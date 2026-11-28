package externals

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	endpoint,
	accessToken string
	appSecret,
	VerifyToken []byte
	client *fasthttp.Client
}

// NewWhatsappClient sends through the Cloud API with the business phone number phoneNumberID. An empty baseURL or
// apiVersion uses graph.facebook.com and WhatsappDefaultAPIVersion; without phoneNumberID or accessToken every send
// returns ErrWhatsappNotConfigured, so a deployment that doesn't use WhatsApp still starts.
func NewWhatsappClient(
	baseURL,
	apiVersion,
	phoneNumberID,
	accessToken,
	appSecret,
	verifyToken string,
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
		appSecret:   helper.String2ByteSlice(appSecret),
		VerifyToken: helper.String2ByteSlice(verifyToken),
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
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

// WhatsApp Cloud API webhook notifications (the "messages" field of a WhatsApp Business Account):
// https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/components

const (
	WhatsappSignatureHeader = "X-Hub-Signature-256"
	WhatsappWebhookObject   = "whatsapp_business_account"

	WhatsappStatusSent      = "sent"
	WhatsappStatusDelivered = "delivered"
	WhatsappStatusRead      = "read"
	WhatsappStatusFailed    = "failed"
)

type WhatsappWebhook struct {
	Object string                 `json:"object"`
	Entry  []WhatsappWebhookEntry `json:"entry"`
}

type WhatsappWebhookEntry struct {
	// ID is the WhatsApp Business Account ID.
	ID      string                  `json:"id"`
	Changes []WhatsappWebhookChange `json:"changes"`
}

type WhatsappWebhookChange struct {
	Field string               `json:"field"`
	Value WhatsappWebhookValue `json:"value"`
}

type WhatsappWebhookValue struct {
	MessagingProduct string                   `json:"messaging_product"`
	Metadata         WhatsappWebhookMetadata  `json:"metadata"`
	Contacts         []WhatsappWebhookContact `json:"contacts,omitempty"`
	Messages         []WhatsappWebhookMessage `json:"messages,omitempty"`
	Statuses         []WhatsappWebhookStatus  `json:"statuses,omitempty"`
	Errors           []WhatsappWebhookError   `json:"errors,omitempty"`
}

type WhatsappWebhookMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type WhatsappWebhookContact struct {
	Profile struct {
		Name string `json:"name"`
	} `json:"profile"`
	WaID string `json:"wa_id"`
}

// WhatsappWebhookMessage is a message a customer sent to the business number.
type WhatsappWebhookMessage struct {
	From      string `json:"from"`
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	// Type is text, button, interactive, image, audio, video, document, sticker, location, contacts,
	// reaction, order, system or unsupported.
	Type string `json:"type"`
	Text *struct {
		Body string `json:"body"`
	} `json:"text,omitempty"`
	// Button is a quick-reply button tapped on a template message.
	Button *struct {
		Text    string `json:"text"`
		Payload string `json:"payload"`
	} `json:"button,omitempty"`
	Interactive *WhatsappWebhookInteractive `json:"interactive,omitempty"`
	// Context refers to the message being replied to.
	Context *struct {
		From string `json:"from"`
		ID   string `json:"id"`
	} `json:"context,omitempty"`
	Errors []WhatsappWebhookError `json:"errors,omitempty"`
}

type WhatsappWebhookInteractive struct {
	// Type is button_reply or list_reply.
	Type        string                   `json:"type"`
	ButtonReply *WhatsappWebhookReplyRow `json:"button_reply,omitempty"`
	ListReply   *WhatsappWebhookReplyRow `json:"list_reply,omitempty"`
}

type WhatsappWebhookReplyRow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// WhatsappWebhookStatus is a delivery update for a message the business sent.
type WhatsappWebhookStatus struct {
	// ID is the wamid returned when the message was sent.
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	Timestamp   string                 `json:"timestamp"`
	RecipientID string                 `json:"recipient_id"`
	Errors      []WhatsappWebhookError `json:"errors,omitempty"`
}

type WhatsappWebhookError struct {
	Code      int    `json:"code"`
	Title     string `json:"title"`
	Message   string `json:"message,omitempty"`
	ErrorData *struct {
		Details string `json:"details"`
	} `json:"error_data,omitempty"`
}

// VerifyWhatsappSignature reports whether signature, the X-Hub-Signature-256 header ("sha256=<hex>"),
// is the HMAC-SHA256 of the raw request body keyed with the Meta app secret.
func (q *WhatsappClient) VerifyWhatsappSignature(ctx context.Context, body []byte, signature string) bool {
	ctxt := "WhatsappClient-VerifyWhatsappSignature"
	digest, ok := strings.CutPrefix(signature, "sha256=")
	if !ok {
		return false
	}
	expected, err := hex.DecodeString(digest)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDecodeString")
		return false
	}
	mac := hmac.New(sha256.New, q.appSecret)
	_, _ = mac.Write(body)
	return hmac.Equal(mac.Sum(nil), expected)
}
