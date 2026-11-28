package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
)

func TestMain(m *testing.M) {
	helper.NewLogger()
	os.Exit(m.Run())
}

func whatsappApp(appSecret, verifyToken string) *fiber.App {
	app := fiber.New(fiber.Config{StrictRouting: true})
	NewWhatsappController(
		externals.NewWhatsappClient("", "", "", "", appSecret, verifyToken, time.Second),
	).Mount(app.Group("/whatsapp"))
	return app
}

func whatsappDo(t *testing.T, app *fiber.App, req *http.Request) (int, string) {
	t.Helper()
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestWhatsappWebhookVerify(t *testing.T) {
	app := whatsappApp("secret", "verify-me")
	for name, c := range map[string]struct {
		query      string
		wantStatus int
		wantBody   string
	}{
		"valid":     {"hub.mode=subscribe&hub.verify_token=verify-me&hub.challenge=1158201444", fiber.StatusOK, "1158201444"},
		"bad token": {"hub.mode=subscribe&hub.verify_token=nope&hub.challenge=1", fiber.StatusForbidden, ""},
		"bad mode":  {"hub.mode=unsubscribe&hub.verify_token=verify-me&hub.challenge=1", fiber.StatusForbidden, ""},
		"no params": {"", fiber.StatusForbidden, ""},
	} {
		status, body := whatsappDo(t, app, httptest.NewRequest(fiber.MethodGet, "/whatsapp/v1/webhook?"+c.query, nil))
		if status != c.wantStatus || (c.wantBody != "" && body != c.wantBody) {
			t.Errorf("%s: %d %q", name, status, body)
		}
	}
}

func TestWhatsappWebhookReceive(t *testing.T) {
	const payload = `{"object":"whatsapp_business_account","entry":[{"id":"102290129340398","changes":[{"field":"messages","value":{` +
		`"messaging_product":"whatsapp","metadata":{"display_phone_number":"15550783881","phone_number_id":"106540352242922"},` +
		`"contacts":[{"profile":{"name":"Ben"},"wa_id":"628116099887"}],` +
		`"messages":[{"from":"628116099887","id":"wamid.IN","timestamp":"1749416383","type":"text","text":{"body":"halo"}}],` +
		`"statuses":[{"id":"wamid.OUT","status":"failed","timestamp":"1749416384","recipient_id":"628116099887",` +
		`"errors":[{"code":131047,"title":"Re-engagement message","error_data":{"details":"more than 24 hours"}}]}]}}]}]}`
	sign := func(secret, body string) string {
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte(body))
		return "sha256=" + hex.EncodeToString(mac.Sum(nil))
	}
	post := func(app *fiber.App, body, signature string) int {
		req := httptest.NewRequest(fiber.MethodPost, "/whatsapp/v1/webhook", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		if signature != "" {
			req.Header.Set(externals.WhatsappSignatureHeader, signature)
		}
		status, _ := whatsappDo(t, app, req)
		return status
	}
	app := whatsappApp("secret", "verify-me")
	for name, c := range map[string]struct {
		body, signature string
		want            int
	}{
		"valid":           {payload, sign("secret", payload), fiber.StatusOK},
		"other object":    {`{"object":"page"}`, sign("secret", `{"object":"page"}`), fiber.StatusOK},
		"wrong secret":    {payload, sign("other", payload), fiber.StatusUnauthorized},
		"no signature":    {payload, "", fiber.StatusUnauthorized},
		"tampered body":   {payload + " ", sign("secret", payload), fiber.StatusUnauthorized},
		"signed non-JSON": {"nope", sign("secret", "nope"), fiber.StatusBadRequest},
	} {
		if got := post(app, c.body, c.signature); got != c.want {
			t.Errorf("%s: got %d, want %d", name, got, c.want)
		}
	}
}
