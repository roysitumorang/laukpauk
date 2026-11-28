package externals

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

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

type (
	WhatsappWebhook struct {
		Object string                 `json:"object"`
		Entry  []WhatsappWebhookEntry `json:"entry"`
	}

	WhatsappWebhookEntry struct {
		// ID is the WhatsApp Business Account ID.
		ID      string                  `json:"id"`
		Changes []WhatsappWebhookChange `json:"changes"`
	}

	WhatsappWebhookChange struct {
		Field string               `json:"field"`
		Value WhatsappWebhookValue `json:"value"`
	}

	WhatsappWebhookValue struct {
		MessagingProduct string                   `json:"messaging_product"`
		Metadata         WhatsappWebhookMetadata  `json:"metadata"`
		Contacts         []WhatsappWebhookContact `json:"contacts,omitempty"`
		Messages         []WhatsappWebhookMessage `json:"messages,omitempty"`
		Statuses         []WhatsappWebhookStatus  `json:"statuses,omitempty"`
		Errors           []WhatsappWebhookError   `json:"errors,omitempty"`
	}

	WhatsappWebhookMetadata struct {
		DisplayPhoneNumber string `json:"display_phone_number"`
		PhoneNumberID      string `json:"phone_number_id"`
	}

	WhatsappWebhookContact struct {
		Profile struct {
			Name string `json:"name"`
		} `json:"profile"`
		WaID string `json:"wa_id"`
	}

	// WhatsappWebhookMessage is a message a customer sent to the business number.
	WhatsappWebhookMessage struct {
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

	WhatsappWebhookInteractive struct {
		// Type is button_reply or list_reply.
		Type        string                   `json:"type"`
		ButtonReply *WhatsappWebhookReplyRow `json:"button_reply,omitempty"`
		ListReply   *WhatsappWebhookReplyRow `json:"list_reply,omitempty"`
	}

	WhatsappWebhookReplyRow struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
	}

	// WhatsappWebhookStatus is a delivery update for a message the business sent.
	WhatsappWebhookStatus struct {
		// ID is the wamid returned when the message was sent.
		ID          string                 `json:"id"`
		Status      string                 `json:"status"`
		Timestamp   string                 `json:"timestamp"`
		RecipientID string                 `json:"recipient_id"`
		Errors      []WhatsappWebhookError `json:"errors,omitempty"`
	}

	WhatsappWebhookError struct {
		Code      int    `json:"code"`
		Title     string `json:"title"`
		Message   string `json:"message,omitempty"`
		ErrorData *struct {
			Details string `json:"details"`
		} `json:"error_data,omitempty"`
	}
)

// VerifyWhatsappSignature reports whether signature, the X-Hub-Signature-256 header ("sha256=<hex>"),
// is the HMAC-SHA256 of the raw request body keyed with the Meta app secret.
func VerifyWhatsappSignature(appSecret string, body []byte, signature string) bool {
	if appSecret == "" {
		return false
	}
	digest, ok := strings.CutPrefix(signature, "sha256=")
	if !ok {
		return false
	}
	expected, err := hex.DecodeString(digest)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(appSecret))
	_, _ = mac.Write(body)
	return hmac.Equal(mac.Sum(nil), expected)
}
