package externals

import (
	"context"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/valyala/fasthttp"
)

type (
	XenditClient struct {
		baseURL,
		apiKey string
		client *fasthttp.Client
	}

	XenditPaymentRequestRequest struct {
		ReferenceID        string `json:"reference_id"`
		PaymentChannelType string `json:"type"`
		Country            string `json:"country"`
		Currency           string `json:"currency"`
		RequestAmount      int64  `json:"request_amount"`
		CaptureMethod      string `json:"capture_method"`
		ChannelCode        string `json:"channel_code"`
		ChannelProperties  struct {
			DisplayName string `json:"display_name"`
			CardDetails struct {
				CVN                   string `json:"cvn"`
				CardNumber            string `json:"card_number"`
				ExpiryYear            string `json:"expiry_year"`
				ExpiryMonth           string `json:"expiry_month"`
				CardholderFirstName   string `json:"cardholder_first_name"`
				CardholderLastName    string `json:"cardholder_last_name"`
				CardholderEmail       string `json:"cardholder_email"`
				CardholderPhoneNumber string `json:"cardholder_phone_number"`
			} `json:"card_details"`
		} `json:"channel_properties"`
		Description string         `json:"description"`
		Metadata    map[string]any `json:"metadata"`
	}

	XenditPaymentRequestResponse struct {
		ErrorCode         string         `json:"error_code"`
		Message           string         `json:"message"`
		PaymentRequestID  string         `json:"payment_request_id"`
		PaymentTokenID    *string        `json:"payment_token_id"`
		Country           string         `json:"country"`
		Currency          string         `json:"currency"`
		BusinessID        string         `json:"business_id"`
		ReferenceID       string         `json:"reference_id"`
		Description       string         `json:"description"`
		Metadata          map[string]any `json:"metadata"`
		Created           time.Time      `json:"created"`
		Updated           time.Time      `json:"updated"`
		Status            string         `json:"status"`
		CaptureMethod     string         `json:"capture_method"`
		ChannelCode       string         `json:"channel_code"`
		RequestAmount     int64          `json:"request_amount"`
		ChannelProperties struct {
			DisplayName string    `json:"display_name"`
			ExpiresAt   time.Time `json:"expires_at"`
		} `json:"channel_properties"`
		ChannelType string `json:"type"`
		Actions     []struct {
			Type       string `json:"type"`
			Descriptor string `json:"descriptor"`
			Value      string `json:"value"`
		} `json:"actions"`
		CustomerID      string `json:"customer_id"`
		LatestPaymentID string `json:"latest_payment_id"`
		FailureCode     string `json:"failure_code"`
	}

	XenditPaymentCancellationResponse struct {
		ErrorCode        string `json:"error_code"`
		Message          string `json:"message"`
		PaymentID        string `json:"payment_id"`
		BusinessID       string `json:"business_id"`
		ReferenceID      string `json:"reference_id"`
		PaymentRequestID string `json:"payment_request_id"`
		PaymentTokenID   string `json:"payment_token_id"`
		CustomerID       string `json:"customer_id"`
		ChannelType      string `json:"type"`
		Country          string `json:"country"`
		Currency         string `json:"currency"`
		RequestAmount    int64  `json:"request_amount"`
		CaptureMethod    string `json:"capture_method"`
		ChannelCode      string `json:"channel_code"`
		Captures         []struct {
			CaptureTimestamp time.Time `json:"capture_timestamp"`
			CaptureID        string    `json:"capture_id"`
			CaptureAmount    int64     `json:"capture_amount"`
		} `json:"captures"`

		Status         string `json:"status"`
		PaymentDetails struct {
			AuthorizationData struct {
				AuthorizationCode             string `json:"authorization_code"`
				CVNVerificationResult         string `json:"cvn_verification_result"`
				AddressVerificationResult     string `json:"address_verification_result"`
				RetrievalReferenceNumber      string `json:"retrieval_reference_number"`
				NetworkResponseCode           string `json:"network_response_code"`
				NetworkResponseCodeDescriptor string `json:"network_response_code_descriptor"`
				NetworkTransactionID          string `json:"network_transaction_id"`
				AcquirerMerchantID            string `json:"acquirer_merchant_id"`
				ReconciliationID              string `json:"reconciliation_id"`
			} `json:"authorization_data"`
			AuthenticationData struct {
				Flow string `json:"flow"`
				ARes struct {
					ECI                 string `json:"eci"`
					MessageVersion      string `json:"message_version"`
					AuthenticationValue string `json:"authentication_value"`
					DSTransID           string `json:"ds_trans_id"`
				} `json:"a_res"`
			} `json:"authentication_data"`
			IssuerName         string `json:"issuer_name"`
			PayerAccountNumber string `json:"payer_account_number"`
			PayerName          string `json:"payer_name"`
			ReceiptID          string `json:"receipt_id"`
			Remark             string `json:"remark"`
			Network            string `json:"network"`
			FundSource         string `json:"fund_source"`
		} `json:"payment_details"`
		FailureCode string         `json:"failure_code"`
		Metadata    map[string]any `json:"metadata"`
		Created     time.Time      `json:"created"`
		Updated     time.Time      `json:"updated"`
	}
)

func NewXenditClient(
	baseURL,
	apiKey string,
	timeout time.Duration,
) *XenditClient {
	sb := pools.BufferPool.Get()
	defer func() {
		sb.Reset()
		pools.BufferPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s:", apiKey)
	return &XenditClient{
		baseURL: baseURL,
		apiKey:  base64.StdEncoding.EncodeToString(sb.Bytes()),
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

func (q *XenditClient) CreatePaymentRequest(
	ctx context.Context,
	buyerID,
	referenceID,
	displayName,
	country,
	currency string,
	requestAmount int64,
	captureMethod,
	channelCode,
	description string,
) (
	paymentRequestID string,
	paymentTokenID *string,
	virtualAccountNumber *string,
	virtualAccountExpiredAt *time.Time,
	err error,
) {
	ctxt := "XenditConsumer-CreatePaymentRequest"
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
	_, _ = fmt.Fprintf(sb, "%s/v3/payment_requests", q.baseURL)
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-version", "2024-11-11")
	req.Header.Set("for-user-id ", buyerID)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "Basic %s", q.apiKey)
	req.Header.Set("Authorization", sb.String())
	request := XenditPaymentRequestRequest{
		ReferenceID:        referenceID,
		PaymentChannelType: "PAY",
		Country:            country,
		Currency:           currency,
		RequestAmount:      requestAmount,
		CaptureMethod:      "AUTOMATIC",
		ChannelCode:        channelCode,
		Description:        description,
	}
	request.ChannelProperties.DisplayName = displayName
	requestBody, err := json.Marshal(request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrMarshal")
		return "", nil, nil, nil, err
	}
	req.SetBodyRaw(requestBody)
	if err = q.client.Do(req, resp); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return "", nil, nil, nil, err
	}
	var response XenditPaymentRequestResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return "", nil, nil, nil, err
	}
	if resp.StatusCode() != fasthttp.StatusCreated {
		return "", nil, nil, nil, fmt.Errorf("%s: %s", response.ErrorCode, response.Message)
	}
	if len(response.Actions) > 0 {
		virtualAccountNumber = &response.Actions[0].Value
	}
	virtualAccountExpiredAt = &response.ChannelProperties.ExpiresAt
	return response.PaymentRequestID, response.PaymentTokenID, virtualAccountNumber, virtualAccountExpiredAt, nil
}

func (q *XenditClient) CancelPaymentRequest(
	ctx context.Context,
	paymentRequestID string,
) (bool, error) {
	ctxt := "XenditConsumer-CancelPaymentRequest"
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
	_, _ = fmt.Fprintf(sb, "%s/v3/payment_requests/%s/cancel", q.baseURL, paymentRequestID)
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-version", "2024-11-11")
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "Basic %s", q.apiKey)
	req.Header.Set("Authorization", sb.String())
	err := q.client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return false, err
	}
	var response XenditPaymentRequestResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return false, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return false, fmt.Errorf("%s: %s", response.ErrorCode, response.Message)
	}
	if response.Status != "SUCCEEDED" {
		return false, errors.New(response.FailureCode)
	}
	return true, nil
}

func (q *XenditClient) CancelPayment(
	ctx context.Context,
	paymentID string,
) (bool, error) {
	ctxt := "XenditConsumer-CancelPayment"
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
	_, _ = fmt.Fprintf(sb, "%s/v3/payments/%s/cancel", q.baseURL, paymentID)
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-version", "2024-11-11")
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "Basic %s", q.apiKey)
	req.Header.Set("Authorization", sb.String())
	err := q.client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return false, err
	}
	var response XenditPaymentCancellationResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return false, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return false, fmt.Errorf("%s: %s", response.ErrorCode, response.Message)
	}
	if response.Status != "SUCCEEDED" {
		return false, errors.New(response.FailureCode)
	}
	return true, nil
}
