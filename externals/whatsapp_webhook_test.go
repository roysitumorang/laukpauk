package externals

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

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
