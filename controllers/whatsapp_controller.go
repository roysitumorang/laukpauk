package controllers

import (
	"crypto/subtle"
	"log/slog"

	"github.com/goccy/go-json/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
)

type whatsappController struct {
	whatsappClient *externals.WhatsappClient
}

func NewWhatsappController(
	whatsappClient *externals.WhatsappClient,
) *whatsappController {
	return &whatsappController{
		whatsappClient: whatsappClient,
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
	if c.Query("hub.mode") != "subscribe" ||
		subtle.ConstantTimeCompare(
			helper.String2ByteSlice(c.Query("hub.verify_token")),
			q.whatsappClient.VerifyToken,
		) != 1 {
		return helper.NewResponse(fiber.StatusForbidden).
			SetMessage("invalid verify token").
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SendString(c, c.Query("hub.challenge"))
}

// V1WebhookReceive accepts notifications signed with the app secret. Meta retries anything other than a
// 200, so a valid notification is acknowledged even when nothing acts on it.
func (q *whatsappController) V1WebhookReceive(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "WhatsappController-V1WebhookReceive"
	// the signature covers the raw body, before any decoding
	body := c.Request().Body()
	if !q.whatsappClient.VerifyWhatsappSignature(ctx, body, c.Get(externals.WhatsappSignatureHeader)) {
		return helper.NewResponse(fiber.StatusUnauthorized).
			SetMessage("invalid signature").
			JSON(c)
	}
	var payload externals.WhatsappWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			JSON(c)
	}
	if payload.Object == externals.WhatsappWebhookObject {
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
	return helper.NewResponse(fiber.StatusOK).
		SendStatus(c)
}

func whatsappErrorDetails(e externals.WhatsappWebhookError) string {
	if e.ErrorData != nil && e.ErrorData.Details != "" {
		return e.ErrorData.Details
	}
	return e.Message
}
