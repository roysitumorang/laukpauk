package externals

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
)

// TestMain sets up the logger: the client logs failed sends with helper.Log.
func TestMain(m *testing.M) {
	helper.NewLogger()
	os.Exit(m.Run())
}

// whatsappServer records the last request and answers with status and body.
func whatsappServer(t *testing.T, status int, body string) (*httptest.Server, *http.Request, *map[string]any) {
	t.Helper()
	var got http.Request
	payload := map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(context.Background())
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv, &got, &payload
}

func TestWhatsappSendText(t *testing.T) {
	srv, req, payload := whatsappServer(t, http.StatusOK,
		`{"messaging_product":"whatsapp","contacts":[{"input":"628116099887","wa_id":"628116099887"}],"messages":[{"id":"wamid.ABC"}]}`)
	client := NewWhatsappClient(srv.URL, "", "106540352242922", "token", time.Second)
	resp, err := client.SendText(context.Background(), "+62 811-609-9887", "Pesanan Anda sudah dikirim", false)
	if err != nil {
		t.Fatal(err)
	}
	if resp.MessageID() != "wamid.ABC" {
		t.Fatalf("message id = %q", resp.MessageID())
	}
	if req.Method != http.MethodPost || req.URL.Path != "/v27.0/106540352242922/messages" {
		t.Fatalf("request = %s %s", req.Method, req.URL.Path)
	}
	if req.Header.Get("Authorization") != "Bearer token" || req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("headers = %v", req.Header)
	}
	p := *payload
	text, _ := p["text"].(map[string]any)
	if p["messaging_product"] != "whatsapp" || p["recipient_type"] != "individual" || p["to"] != "628116099887" ||
		p["type"] != "text" || text["body"] != "Pesanan Anda sudah dikirim" || text["preview_url"] != false {
		t.Fatalf("payload = %v", p)
	}
	if _, ok := p["template"]; ok {
		t.Fatalf("payload has a template: %v", p)
	}
}

func TestWhatsappSendTemplate(t *testing.T) {
	srv, _, payload := whatsappServer(t, http.StatusOK, `{"messages":[{"id":"wamid.T"}]}`)
	client := NewWhatsappClient(srv.URL, "v26.0", "1", "token", time.Second)
	if _, err := client.SendTemplate(context.Background(), "08116099887", "order_shipped", "id", "INV-001", "Ben"); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal((*payload)["template"])
	want := `{"components":[{"parameters":[{"text":"INV-001","type":"text"},{"text":"Ben","type":"text"}],"type":"body"}],` +
		`"language":{"code":"id"},"name":"order_shipped"}`
	if string(b) != want || (*payload)["to"] != "628116099887" || (*payload)["type"] != "template" {
		t.Fatalf("template = %s, payload = %v", b, *payload)
	}
}

func TestWhatsappErrors(t *testing.T) {
	srv, _, _ := whatsappServer(t, http.StatusBadRequest,
		`{"error":{"message":"(#131030) Recipient phone number not in allowed list","type":"OAuthException","code":131030,"error_subcode":2655007,"fbtrace_id":"Az8"}}`)
	client := NewWhatsappClient(srv.URL, "", "1", "token", time.Second)
	_, err := client.SendText(context.Background(), "628116099887", "hi", false)
	var apiErr *WhatsappError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest || apiErr.Code != 131030 ||
		!strings.Contains(apiErr.Error(), "allowed list") {
		t.Fatalf("err = %v", err)
	}

	// a non-JSON body fails to unmarshal, whatever the status
	srv2, _, _ := whatsappServer(t, http.StatusBadGateway, `upstream down`)
	if _, err = NewWhatsappClient(srv2.URL, "", "1", "token", time.Second).SendText(context.Background(), "1", "hi", false); err == nil ||
		errors.As(err, &apiErr) || strings.Contains(err.Error(), "HTTP 502") {
		t.Fatalf("non-JSON error = %v", err)
	}

	// an error status whose JSON body has no "error" object
	srv3, _, _ := whatsappServer(t, http.StatusInternalServerError, `{}`)
	if _, err = NewWhatsappClient(srv3.URL, "", "1", "token", time.Second).SendText(context.Background(), "1", "hi", false); err == nil ||
		!strings.Contains(err.Error(), "HTTP 500: {}") {
		t.Fatalf("JSON error without details = %v", err)
	}

	unconfigured := NewWhatsappClient("", "", "", "", time.Second)
	if _, err = unconfigured.SendText(context.Background(), "1", "hi", false); !errors.Is(err, ErrWhatsappNotConfigured) {
		t.Fatalf("unconfigured = %v", err)
	}
	if _, err = client.SendText(context.Background(), "1", " ", false); !errors.Is(err, ErrWhatsappEmptyBody) {
		t.Fatalf("empty body = %v", err)
	}
	if _, err = client.SendText(context.Background(), "1", strings.Repeat("a", WhatsappTextMaxLength+1), false); !errors.Is(err, ErrWhatsappBodyTooLong) {
		t.Fatalf("long body = %v", err)
	}
	if _, err = client.SendText(context.Background(), "+-", "hi", false); !errors.Is(err, ErrWhatsappEmptyRecipient) {
		t.Fatalf("empty recipient = %v", err)
	}
	if _, err = client.SendTemplate(context.Background(), "1", "", "id"); !errors.Is(err, ErrWhatsappEmptyTemplate) {
		t.Fatalf("empty template = %v", err)
	}
}

func TestNormalizeWhatsappNumber(t *testing.T) {
	for in, want := range map[string]string{
		"+62 811-609-9887":  "628116099887",
		"08116099887":       "628116099887",
		"628116099887":      "628116099887",
		"(+1) 650 555 1234": "16505551234",
	} {
		if got := NormalizeWhatsappNumber(in); got != want {
			t.Errorf("NormalizeWhatsappNumber(%q) = %q; want %q", in, got, want)
		}
	}
}

func whatsappSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWhatsappSignature(t *testing.T) {
	body := []byte(`{"object":"whatsapp_business_account"}`)
	valid := whatsappSignature("secret", body)
	for name, c := range map[string]struct {
		secret, signature string
		body              []byte
		want              bool
	}{
		"valid":         {"secret", valid, body, true},
		"other secret":  {"other", valid, body, false},
		"changed body":  {"secret", valid, []byte(`{"object":"page"}`), false},
		"no prefix":     {"secret", valid[len("sha256="):], body, false},
		"not hex":       {"secret", "sha256=zz", body, false},
		"missing":       {"secret", "", body, false},
		"no app secret": {"", whatsappSignature("", body), body, false},
	} {
		if got := VerifyWhatsappSignature(c.secret, c.body, c.signature); got != c.want {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
}
