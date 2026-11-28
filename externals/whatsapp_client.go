package externals

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WhatsApp Cloud API (graph.facebook.com/<version>/<phone-number-id>/messages), with only the standard library.

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
	return fmt.Sprintf("whatsapp: %s (HTTP %d, code %d, subcode %d, fbtrace_id %s)",
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
	client      *http.Client
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
		endpoint = strings.TrimRight(baseURL, "/") + "/" + apiVersion + "/" + phoneNumberID + "/messages"
	}
	return &WhatsappClient{
		endpoint:    endpoint,
		accessToken: accessToken,
		client:      &http.Client{Timeout: timeout},
	}
}

// NormalizeWhatsappNumber turns a phone number into the digits the Cloud API takes: "+62 811-609-9887" and
// "08116099887" both become "628116099887" (a leading 0 is read as Indonesia's 62).
func NormalizeWhatsappNumber(phone string) string {
	var b strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
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
	if len([]rune(body)) > WhatsappTextMaxLength {
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
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+q.accessToken)
	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var failure struct {
			Error *WhatsappError `json:"error"`
		}
		if err = json.Unmarshal(body, &failure); err != nil || failure.Error == nil {
			return nil, fmt.Errorf("whatsapp: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		failure.Error.StatusCode = resp.StatusCode
		return nil, failure.Error
	}
	var response WhatsappResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
