package externals

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/valyala/fasthttp"
)

var (
	ErrRoutesApiEmptyResponse = errors.New("routes API call returns empty response")
)

type (
	RoutesApiClient struct {
		endpoint,
		apiKey string
		client *fasthttp.Client
	}

	RoutesApiResponse struct {
		DistanceMeters int64 `json:"distanceMeters"`
	}
)

func NewRoutesApiClient(
	endpoint,
	apiKey string,
	timeout time.Duration,
) *RoutesApiClient {
	return &RoutesApiClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

func (q *RoutesApiClient) ComputeDistanceMeters(
	ctx context.Context,
	originLongitude,
	originLatitude,
	destinationLongitude,
	destinationLatitude string,
) (int64, error) {
	ctxt := "RoutesApiClient-ComputeDistanceMeters"
	sb := pools.BufferPool.Get()
	defer func() {
		sb.Reset()
		pools.BufferPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`{"origins":[{"waypoint":{"location":{"latLng":{"latitude":%s,"longitude":%s}}}}],"destinations":[{"waypoint":{"location":{"latLng":{"latitude":%s,"longitude":%s}}}}],"travelMode":"TWO_WHEELER","routingPreference":"TRAFFIC_AWARE"}`,
		originLatitude,
		originLongitude,
		destinationLatitude,
		destinationLongitude,
	)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.SetRequestURI(q.endpoint)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Add("X-Goog-Api-Key", q.apiKey)
	req.Header.Add("X-Goog-FieldMask", "distanceMeters")
	req.SetBodyRaw(sb.Bytes())
	err := q.client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return 0, err
	}
	var response []RoutesApiResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return 0, err
	}
	if len(response) == 0 {
		return 0, ErrRoutesApiEmptyResponse
	}
	return response[0].DistanceMeters, nil
}
