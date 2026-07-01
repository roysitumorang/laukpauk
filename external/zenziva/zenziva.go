package sms

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/valyala/fasthttp"
)

const (
	MessageTypeSms MessageType = iota
	MessageTypeOtp
)

const (
	DeliveryStatusSent DeliveryStatus = iota
	DeliveryStatusPending
	DeliveryStatusFailed
	DeliveryStatusRejected
	DeliveryStatusError
	DeliveryStatusUndeliverable
)

type (
	MessageType    int
	DeliveryStatus int

	ZenzivaService struct {
		BaseURL,
		UserKey,
		ApiKey string
		MessageMaxLength int
		Client           *fasthttp.Client
	}

	Request struct {
		UserKey string `json:"userkey"`
		PassKey string `json:"passkey"`
		To      string `json:"to"`
		Message string `json:"message"`
	}

	Response struct {
		MessageID int64  `json:"messageId,string"`
		To        string `json:"to"`
		Status    int    `json:"status,string"`
		Text      string `json:"text"`
	}

	Report struct {
		Status         int            `json:"status,string"`
		MessageID      int64          `json:"messageId,string"`
		Date           string         `json:"date"`
		Time           string         `json:"time"`
		Dest           string         `json:"dest"`
		Message        string         `json:"msg"`
		DeliveryStatus DeliveryStatus `json:"msg-status"`
	}

	Balance struct {
		Balance int64  `json:"balance,string"`
		Status  int    `json:"status,string"`
		Text    string `json:"text"`
	}
)

func NewZenzivaService(
	baseURL,
	userKey,
	apiKey string,
	messageMaxLength int,
	timeout time.Duration,
) *ZenzivaService {
	return new(ZenzivaService{
		BaseURL:          baseURL,
		UserKey:          userKey,
		ApiKey:           apiKey,
		MessageMaxLength: messageMaxLength,
		Client: new(fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		}),
	})
}

func (q *ZenzivaService) Send(ctx context.Context, messageType MessageType, to, message string) (*int64, error) {
	ctxt := "ZenzivaService-Send"
	var sb strings.Builder
	requestBody, err := json.Marshal(Request{
		UserKey: q.UserKey,
		PassKey: q.ApiKey,
		To:      to,
		Message: helper.TruncateString(message, q.MessageMaxLength),
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
	_, _ = fmt.Fprintf(&sb, "%s/masking/api/%s/", q.BaseURL, messageType.String())
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
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
	return new(response.MessageID), nil
}

func (q *ZenzivaService) CheckReport(ctx context.Context, messageID int64) (*Report, error) {
	ctxt := "ZenzivaService-CheckReport"
	var sb strings.Builder
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	_, _ = fmt.Fprintf(&sb, "%s/masking/api/report/?userkey=%s&passkey=%s&messageId=%d", q.BaseURL, q.UserKey, q.ApiKey, messageID)
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodGet)
	err := q.Client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	var response Report
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return new(response), nil
}

func (q *ZenzivaService) CheckBalance(ctx context.Context) (*Balance, error) {
	ctxt := "ZenzivaService-CheckBalance"
	var sb strings.Builder
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	_, _ = fmt.Fprintf(&sb, "%s/api/balance/?userkey=%s&passkey=%s", q.BaseURL, q.UserKey, q.ApiKey)
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodGet)
	err := q.Client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	var response Balance
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return new(response), nil
}

func (q DeliveryStatus) Int() int {
	return int(q)
}

func (q DeliveryStatus) String() string {
	if q == DeliveryStatusPending {
		return "Pending"
	}
	if q == DeliveryStatusFailed {
		return "Failed"
	}
	if q == DeliveryStatusRejected {
		return "Rejected"
	}
	if q == DeliveryStatusError {
		return "Error"
	}
	if q == DeliveryStatusUndeliverable {
		return "Undeliverable"
	}
	return "Sent"
}

func (q MessageType) Int() int {
	return int(q)
}

func (q MessageType) String() string {
	if q == MessageTypeOtp {
		return "sendOTP"
	}
	return "sendsms"
}
