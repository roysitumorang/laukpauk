package externals

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json/v2"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/valyala/fasthttp"
)

type ZenzivaMessageType uint8

const (
	ZenzivaMessageTypeSms ZenzivaMessageType = iota
	ZenzivaMessageTypeOtp
)

func (q ZenzivaMessageType) Int() uint8 {
	return uint8(q)
}

func (q ZenzivaMessageType) String() string {
	if q == ZenzivaMessageTypeOtp {
		return "sendOTP"
	}
	return "sendsms"
}

type ZenzivaRequest struct {
	UserKey string `json:"userkey"`
	PassKey string `json:"passkey"`
	To      string `json:"to"`
	Message string `json:"message"`
}

type ZenzivaResponse struct {
	MessageID string `json:"messageId"`
	To        string `json:"to"`
	Status    string `json:"status"`
	Text      string `json:"text"`
}

type ZenzivaReport struct {
	Status         int    `json:"status,string"`
	MessageID      int64  `json:"messageId,string"`
	Date           string `json:"date"`
	Time           string `json:"time"`
	Dest           string `json:"dest"`
	Message        string `json:"msg"`
	DeliveryStatus string `json:"msg-status"`
}

type ZenzivaBalance struct {
	Balance int64  `json:"balance,string"`
	Status  int    `json:"status,string"`
	Text    string `json:"text"`
}

type ZenzivaClient struct {
	baseURL,
	userKey,
	apiKey string
	messageMaxLength int
	client           *fasthttp.Client
}

func NewZenzivaClient(
	baseURL,
	userKey,
	apiKey string,
	messageMaxLength int,
	timeout time.Duration,
) *ZenzivaClient {
	return &ZenzivaClient{
		baseURL:          baseURL,
		userKey:          userKey,
		apiKey:           apiKey,
		messageMaxLength: messageMaxLength,
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

func (q *ZenzivaClient) SendSMS(ctx context.Context, to, message string) (*int64, error) {
	return q.send(ctx, ZenzivaMessageTypeSms, to, message)
}

func (q *ZenzivaClient) SendOTP(ctx context.Context, to, message string) (*int64, error) {
	return q.send(ctx, ZenzivaMessageTypeOtp, to, message)
}

func (q *ZenzivaClient) send(ctx context.Context, messageType ZenzivaMessageType, to, message string) (*int64, error) {
	ctxt := "ZenzivaClient-send"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
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
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s/masking/api/%s/", q.baseURL, messageType.String())
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
	if response.Status != "1" {
		return nil, errors.New(response.Text)
	}
	messageID, err := strconv.ParseInt(response.MessageID, 10, 64)
	if err != nil {
		return nil, err
	}
	return &messageID, nil
}

func (q *ZenzivaClient) CheckReport(ctx context.Context, messageID int64) (*string, error) {
	ctxt := "ZenzivaConsumer-CheckReport"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s/masking/api/report/?userkey=%s&passkey=%s&messageId=%d", q.baseURL, q.userKey, q.apiKey, messageID)
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

func (q *ZenzivaClient) CheckBalance(ctx context.Context) (*ZenzivaBalance, error) {
	ctxt := "ZenzivaConsumer-CheckBalance"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s/api/balance/?userkey=%s&passkey=%s", q.baseURL, q.userKey, q.apiKey)
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
	return &response, nil
}
