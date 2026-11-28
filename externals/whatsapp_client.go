package externals

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/goccy/go-json/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/valyala/fasthttp"
)

// WhatsApp Cloud API (graph.facebook.com/<version>/<phone-number-id>/messages).

const (
	WhatsappDefaultBaseURL    = "https://graph.facebook.com"
	WhatsappDefaultAPIVersion = "v27.0"
	// WhatsappTextMaxLength is the most characters a text message body may have.
	WhatsappTextMaxLength = 4096
)

var (
	// ErrWhatsappNotConfigured is returned by every send when the phone number ID or access token is missing.
	ErrWhatsappNotConfigured  = errors.New("whatsapp: WHATSAPP_PHONE_NUMBER_ID and WHATSAPP_ACCESS_TOKEN are required")
	ErrWhatsappEmptyRecipient = errors.New("whatsapp: recipient phone number is empty")
	ErrWhatsappEmptyBody      = errors.New("whatsapp: message body is empty")
	ErrWhatsappBodyTooLong    = fmt.Errorf("whatsapp: message body is longer than %d characters", WhatsappTextMaxLength)
	ErrWhatsappEmptyTemplate  = errors.New("whatsapp: template name and language code are required")
)

// WhatsappError is an error the Cloud API returned ({"error": {...}}).
type WhatsappError struct {
	StatusCode   int    `json:"-"`
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode"`
	FbtraceID    string `json:"fbtrace_id"`
}

func (e *WhatsappError) Error() string {
	return helper.Sprintf("whatsapp: %s (HTTP %d, code %d, subcode %d, fbtrace_id %s)",
		e.Message, e.StatusCode, e.Code, e.ErrorSubcode, e.FbtraceID)
}

type WhatsappText struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

type WhatsappTemplateLanguage struct {
	Code string `json:"code"`
}

type WhatsappTemplateParameter struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type WhatsappTemplateComponent struct {
	Type       string                      `json:"type"`
	Parameters []WhatsappTemplateParameter `json:"parameters,omitempty"`
}

type WhatsappTemplate struct {
	Name       string                      `json:"name"`
	Language   WhatsappTemplateLanguage    `json:"language"`
	Components []WhatsappTemplateComponent `json:"components,omitempty"`
}

// WhatsappMessage is the body of POST /<phone-number-id>/messages for a text or template message.
type WhatsappMessage struct {
	MessagingProduct string            `json:"messaging_product"`
	RecipientType    string            `json:"recipient_type"`
	To               string            `json:"to"`
	Type             string            `json:"type"`
	Text             *WhatsappText     `json:"text,omitempty"`
	Template         *WhatsappTemplate `json:"template,omitempty"`
}

type WhatsappContact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type WhatsappSentMessage struct {
	ID            string `json:"id"`
	MessageStatus string `json:"message_status,omitempty"`
}

// WhatsappResponse is what a successful send returns; Messages[0].ID is the message ID (wamid.…).
type WhatsappResponse struct {
	MessagingProduct string                `json:"messaging_product"`
	Contacts         []WhatsappContact     `json:"contacts"`
	Messages         []WhatsappSentMessage `json:"messages"`
	Error            *WhatsappError        `json:"error"`
}

// MessageID is the ID of the sent message, or "" when the response has none.
func (q *WhatsappResponse) MessageID() string {
	if q == nil || len(q.Messages) == 0 {
		return ""
	}
	return q.Messages[0].ID
}

type WhatsappClient struct {
	endpoint    string
	accessToken string
	client      *fasthttp.Client
}

// NewWhatsappClient sends through the Cloud API with the business phone number phoneNumberID. An empty baseURL or
// apiVersion uses graph.facebook.com and WhatsappDefaultAPIVersion; without phoneNumberID or accessToken every send
// returns ErrWhatsappNotConfigured, so a deployment that doesn't use WhatsApp still starts.
func NewWhatsappClient(
	baseURL,
	apiVersion,
	phoneNumberID,
	accessToken string,
	timeout time.Duration,
) *WhatsappClient {
	if baseURL == "" {
		baseURL = WhatsappDefaultBaseURL
	}
	if apiVersion == "" {
		apiVersion = WhatsappDefaultAPIVersion
	}
	var endpoint string
	if phoneNumberID != "" && accessToken != "" {
		endpoint = helper.Sprintf("%s/%s/%s/messages", strings.TrimRight(baseURL, "/"), apiVersion, phoneNumberID)
	}
	return &WhatsappClient{
		endpoint:    endpoint,
		accessToken: accessToken,
		client:      &fasthttp.Client{ReadTimeout: timeout, WriteTimeout: timeout},
	}
}

// NormalizeWhatsappNumber turns a phone number into the digits the Cloud API takes: "+62 811-609-9887" and
// "08116099887" both become "628116099887" (a leading 0 is read as Indonesia's 62).
func NormalizeWhatsappNumber(phone string) string {
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			_, _ = sb.WriteRune(r)
		}
	}
	digits := sb.String()
	if strings.HasPrefix(digits, "0") {
		digits = "62" + digits[1:]
	}
	return digits
}

// SendText sends a text message. Outside the 24 hours after the user last wrote, WhatsApp only delivers
// templates; use SendTemplate then.
func (q *WhatsappClient) SendText(ctx context.Context, to, body string, previewURL bool) (*WhatsappResponse, error) {
	if strings.TrimSpace(body) == "" {
		return nil, ErrWhatsappEmptyBody
	}
	if utf8.RuneCountInString(body) > WhatsappTextMaxLength {
		return nil, ErrWhatsappBodyTooLong
	}
	return q.send(ctx, &WhatsappMessage{
		To:   to,
		Type: "text",
		Text: &WhatsappText{PreviewURL: previewURL, Body: body},
	})
}

// SendTemplate sends an approved message template, filling its body variables ({{1}}, {{2}}, …) with
// bodyParameters in order.
func (q *WhatsappClient) SendTemplate(
	ctx context.Context,
	to,
	name,
	languageCode string,
	bodyParameters ...string,
) (*WhatsappResponse, error) {
	if name == "" || languageCode == "" {
		return nil, ErrWhatsappEmptyTemplate
	}
	template := &WhatsappTemplate{Name: name, Language: WhatsappTemplateLanguage{Code: languageCode}}
	if len(bodyParameters) > 0 {
		parameters := make([]WhatsappTemplateParameter, len(bodyParameters))
		for i, p := range bodyParameters {
			parameters[i] = WhatsappTemplateParameter{Type: "text", Text: p}
		}
		template.Components = []WhatsappTemplateComponent{{Type: "body", Parameters: parameters}}
	}
	return q.send(ctx, &WhatsappMessage{To: to, Type: "template", Template: template})
}

func (q *WhatsappClient) send(ctx context.Context, message *WhatsappMessage) (*WhatsappResponse, error) {
	ctxt := "WhatsappClient-send"
	if q.endpoint == "" {
		return nil, ErrWhatsappNotConfigured
	}
	message.MessagingProduct = "whatsapp"
	message.RecipientType = "individual"
	message.To = NormalizeWhatsappNumber(message.To)
	if message.To == "" {
		return nil, ErrWhatsappEmptyRecipient
	}
	payload, err := json.Marshal(message)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrMarshal")
		return nil, err
	}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.SetRequestURI(q.endpoint)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Authorization", helper.Sprintf("Bearer %s", q.accessToken))
	req.SetBodyRaw(payload)
	if err = q.client.Do(req, resp); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	statusCode := resp.StatusCode()
	body := resp.Body()
	var response WhatsappResponse
	if err = json.Unmarshal(body, &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	if statusCode < fiber.StatusOK || statusCode >= fiber.StatusMultipleChoices {
		if response.Error == nil {
			return nil, fmt.Errorf("whatsapp: HTTP %d: %s", statusCode, helper.ByteSlice2String(body))
		}
		response.Error.StatusCode = statusCode
		return nil, response.Error
	}
	return &response, nil
}
