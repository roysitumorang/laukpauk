package controllers

import (
	"context"
	"crypto/subtle"
	"log/slog"

	"github.com/goccy/go-json/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
)

type whatsappController struct {
	appSecret   string
	verifyToken string
}

// NewWhatsappController handles the WhatsApp Cloud API webhook. appSecret is the Meta app secret that
// signs notifications; verifyToken is the token entered when subscribing the webhook in the Meta app.
func NewWhatsappController(appSecret, verifyToken string) *whatsappController {
	return &whatsappController{
		appSecret:   appSecret,
		verifyToken: verifyToken,
	}
}

func (q *whatsappController) Mount(r fiber.Router) {
	r.Group("/v1/webhook").
		Get("", q.V1WebhookVerify).
		Post("", q.V1WebhookReceive)
}

// V1WebhookVerify answers Meta's subscription check: it echoes hub.challenge when hub.mode is
// "subscribe" and hub.verify_token matches.
func (q *whatsappController) V1WebhookVerify(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "WhatsappController-V1WebhookVerify"
	if q.verifyToken == "" {
		helper.Log(ctx, slog.LevelError, "WHATSAPP_WEBHOOK_VERIFY_TOKEN is not set", ctxt, "ErrVerifyToken")
		return helper.NewResponse(fiber.StatusServiceUnavailable).
			SetMessage("whatsapp webhook is not configured").
			WriteResponse(c)
	}
	if c.Query("hub.mode") != "subscribe" ||
		subtle.ConstantTimeCompare([]byte(c.Query("hub.verify_token")), []byte(q.verifyToken)) != 1 {
		helper.Log(ctx, slog.LevelWarn, "invalid webhook verification request", ctxt, "ErrVerifyToken")
		return helper.NewResponse(fiber.StatusForbidden).
			SetMessage("invalid verify token").
			WriteResponse(c)
	}
	return c.Status(fiber.StatusOK).SendString(c.Query("hub.challenge"))
}

// V1WebhookReceive accepts notifications signed with the app secret. Meta retries anything other than a
// 200, so a valid notification is acknowledged even when nothing acts on it.
func (q *whatsappController) V1WebhookReceive(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "WhatsappController-V1WebhookReceive"
	if q.appSecret == "" {
		helper.Log(ctx, slog.LevelError, "WHATSAPP_APP_SECRET is not set", ctxt, "ErrVerifyWhatsappSignature")
		return helper.NewResponse(fiber.StatusServiceUnavailable).
			SetMessage("whatsapp webhook is not configured").
			WriteResponse(c)
	}
	// the signature covers the raw body, before any decoding
	body := c.Request().Body()
	if !externals.VerifyWhatsappSignature(q.appSecret, body, c.Get(externals.WhatsappSignatureHeader)) {
		helper.Log(ctx, slog.LevelWarn, "invalid webhook signature", ctxt, "ErrVerifyWhatsappSignature")
		return helper.NewResponse(fiber.StatusUnauthorized).
			SetMessage("invalid signature").
			WriteResponse(c)
	}
	var payload externals.WhatsappWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if payload.Object == externals.WhatsappWebhookObject {
		q.handleWebhook(ctx, &payload)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (q *whatsappController) handleWebhook(ctx context.Context, payload *externals.WhatsappWebhook) {
	ctxt := "WhatsappController-handleWebhook"
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}
			value := change.Value
			for _, e := range value.Errors {
				helper.Log(ctx, slog.LevelWarn, helper.Sprintf("whatsapp: error %d: %s %s", e.Code, e.Title, whatsappErrorDetails(e)), ctxt, "")
			}
			for _, status := range value.Statuses {
				if status.Status == externals.WhatsappStatusFailed {
					for _, e := range status.Errors {
						helper.Log(ctx, slog.LevelWarn, helper.Sprintf("whatsapp: message %s to %s failed: %d %s %s", status.ID, status.RecipientID, e.Code, e.Title, whatsappErrorDetails(e)), ctxt, "")
					}
					continue
				}
				helper.Log(ctx, slog.LevelInfo, helper.Sprintf("whatsapp: message %s to %s %s", status.ID, status.RecipientID, status.Status), ctxt, "")
			}
			// message contents aren't logged: they're the customer's
			for _, message := range value.Messages {
				helper.Log(ctx, slog.LevelInfo, helper.Sprintf("whatsapp: received %s message %s from %s", message.Type, message.ID, message.From), ctxt, "")
			}
		}
	}
}

func whatsappErrorDetails(e externals.WhatsappWebhookError) string {
	if e.ErrorData != nil && e.ErrorData.Details != "" {
		return e.ErrorData.Details
	}
	return e.Message
}
