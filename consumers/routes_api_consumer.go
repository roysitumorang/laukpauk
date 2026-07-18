package consumers

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/valyala/fasthttp"
)

var (
	ErrRoutesApiEmptyResponse = errors.New("routes API call returns empty response")
)

type (
	RoutesApiConsumer struct {
		endpoint,
		apiKey string
		client *fasthttp.Client
	}

	RoutesApiResponse struct {
		DistanceMeters int64 `json:"distanceMeters"`
	}
)

func NewRoutesApiConsumer(
	endpoint,
	apiKey string,
	timeout time.Duration,
) *RoutesApiConsumer {
	return &RoutesApiConsumer{
		endpoint: endpoint,
		apiKey:   apiKey,
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

func (q *RoutesApiConsumer) ComputeDistanceMeters(
	ctx context.Context,
	originLongitude,
	originLatitude,
	destinationLongitude,
	destinationLatitude string,
) (int64, error) {
	ctxt := "RoutesApiConsumer-ComputeDistanceMeters"
	var sb strings.Builder
	_, _ = fmt.Fprintf(
		&sb,
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
	req.SetBodyRaw(helper.String2ByteSlice(sb.String()))
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
