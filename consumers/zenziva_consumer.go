package consumers

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/valyala/fasthttp"
)

const (
	ZensivaMessageTypeSms ZenzivaMessageType = iota
	ZenzivaMessageTypeOtp
)

type (
	ZenzivaMessageType int

	ZenzivaConsumer struct {
		baseURL,
		userKey,
		apiKey string
		messageMaxLength int
		client           *fasthttp.Client
	}

	ZenzivaRequest struct {
		UserKey string `json:"userkey"`
		PassKey string `json:"passkey"`
		To      string `json:"to"`
		Message string `json:"message"`
	}

	ZenzivaResponse struct {
		MessageID int64  `json:"messageId,string"`
		To        string `json:"to"`
		Status    int    `json:"status,string"`
		Text      string `json:"text"`
	}

	ZenzivaReport struct {
		Status         int    `json:"status,string"`
		MessageID      int64  `json:"messageId,string"`
		Date           string `json:"date"`
		Time           string `json:"time"`
		Dest           string `json:"dest"`
		Message        string `json:"msg"`
		DeliveryStatus string `json:"msg-status"`
	}

	ZenzivaBalance struct {
		Balance int64  `json:"balance,string"`
		Status  int    `json:"status,string"`
		Text    string `json:"text"`
	}
)

func NewZenzivaConsumer(
	baseURL,
	userKey,
	apiKey string,
	messageMaxLength int,
	timeout time.Duration,
) *ZenzivaConsumer {
	return new(ZenzivaConsumer{
		baseURL:          baseURL,
		userKey:          userKey,
		apiKey:           apiKey,
		messageMaxLength: messageMaxLength,
		client: new(fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		}),
	})
}

func (q *ZenzivaConsumer) SendSMS(ctx context.Context, to, message string) (*int64, error) {
	return q.send(ctx, ZensivaMessageTypeSms, to, message)
}

func (q *ZenzivaConsumer) SendOTP(ctx context.Context, to, message string) (*int64, error) {
	return q.send(ctx, ZenzivaMessageTypeOtp, to, message)
}

func (q *ZenzivaConsumer) send(ctx context.Context, messageType ZenzivaMessageType, to, message string) (*int64, error) {
	ctxt := "ZenzivaConsumer-send"
	var sb strings.Builder
	requestBody, err := json.Marshal(ZenzivaRequest{
		UserKey: q.userKey,
		PassKey: q.apiKey,
		To:      to,
		Message: helper.TruncateString(message, q.messageMaxLength),
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
	_, _ = fmt.Fprintf(&sb, "%s/masking/api/%s/", q.baseURL, messageType.String())
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBodyRaw(requestBody)
	if err = q.client.Do(req, resp); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	var response ZenzivaResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return new(response.MessageID), nil
}

func (q *ZenzivaConsumer) CheckReport(ctx context.Context, messageID int64) (*string, error) {
	ctxt := "ZenzivaConsumer-CheckReport"
	var sb strings.Builder
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	_, _ = fmt.Fprintf(&sb, "%s/masking/api/report/?userkey=%s&passkey=%s&messageId=%d", q.baseURL, q.userKey, q.apiKey, messageID)
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodGet)
	err := q.client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	var response ZenzivaReport
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return new(strings.ToLower(response.DeliveryStatus)), nil
}

func (q *ZenzivaConsumer) CheckBalance(ctx context.Context) (*ZenzivaBalance, error) {
	ctxt := "ZenzivaConsumer-CheckBalance"
	var sb strings.Builder
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	_, _ = fmt.Fprintf(&sb, "%s/api/balance/?userkey=%s&passkey=%s", q.baseURL, q.userKey, q.apiKey)
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodGet)
	err := q.client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	var response ZenzivaBalance
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return new(response), nil
}

func (q ZenzivaMessageType) Int() int {
	return int(q)
}

func (q ZenzivaMessageType) String() string {
	if q == ZenzivaMessageTypeOtp {
		return "sendOTP"
	}
	return "sendsms"
}
