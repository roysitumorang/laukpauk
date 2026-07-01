package fcm

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/valyala/fasthttp"
)

type (
	FcmService struct {
		BaseURL string
		Client  *fasthttp.Client
	}

	Notification struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Image string `json:"image,omitempty"`
	}

	Request struct {
		Token        string            `json:"token"`
		Notification Notification      `json:"notification"`
		Data         map[string]string `json:"data"`
	}

	Response struct {
		Name string `json:"name"`
	}

	ServiceAccount struct {
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
		ClientEmail  string `json:"client_email"`
		TokenURI     string `json:"token_uri"`
	}

	TokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
)

func NewFcmService(
	projectID string,
	timeout time.Duration,
) *FcmService {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)
	return new(FcmService{
		BaseURL: sb.String(),
		Client: new(fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		}),
	})
}

func (q *FcmService) Send(ctx context.Context, token, title, body, image string, data map[string]string) (*string, error) {
	ctxt := "FcmService-Send"
	bearerToken, err := q.CreateOauth2Token(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrTokenSource")
		return nil, err
	}
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "Bearer %s", bearerToken.AccessToken)
	requestBody, err := json.Marshal(Request{
		Token: token,
		Notification: Notification{
			Title: title,
			Body:  body,
			Image: image,
		},
		Data: data,
	})
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
	req.SetRequestURI(q.BaseURL)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Authorization", sb.String())
	req.SetBodyRaw(requestBody)
	if err = q.Client.Do(req, resp); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	var response Response
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return new(response.Name), nil
}

func (q *FcmService) CreateOauth2Token(ctx context.Context) (*TokenResponse, error) {
	ctxt := "FcmService-CreateOauth2Token"
	jsonBytes, err := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDecodeString")
		return nil, err
	}
	var sa ServiceAccount
	if err = json.Unmarshal(jsonBytes, &sa); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	block, _ := pem.Decode(helper.String2ByteSlice(sa.PrivateKey))
	if block == nil {
		return nil, errors.New("error decoding private key PEM")
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParsePKCS8PrivateKey")
		return nil, err
	}
	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   sa.ClientEmail,
		"scope": "https://googleapis.com",
		"aud":   sa.TokenURI,
		"exp":   now.Add(50 * time.Minute).Unix(),
		"iat":   now.Unix(),
	})
	token.Header["kid"] = sa.PrivateKeyID
	signedJWT, err := token.SignedString(privateKey)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSignedString")
		return nil, err
	}
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", signedJWT)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.SetRequestURI(sa.TokenURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.SetBodyRaw(helper.String2ByteSlice(data.Encode()))
	if err = q.Client.Do(req, resp); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	respBody := resp.Body()
	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("failed to obtain token. status: %d, response: %s", resp.StatusCode(), helper.ByteSlice2String(respBody))
	}
	var tokenResp TokenResponse
	if err = json.Unmarshal(respBody, &tokenResp); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return new(tokenResp), nil
}
