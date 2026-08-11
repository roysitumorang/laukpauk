package helper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/pools"
)

type (
	Response struct {
		RequestID  string    `json:"request_id" example:"6ba3451b-ac73-483e-8481-2ac53f5e75a2"`
		RequestURL string    `json:"request_url" example:"GET http://localhost:19000/ping"`
		StatusCode int       `json:"status_code" example:"200"`
		Status     string    `json:"status" example:"OK"`
		Message    string    `json:"message" example:""`
		Timestamp  time.Time `json:"timestamp" example:"2025-02-05T12:22:47.608963985+07:00"`
		Latency    string    `json:"latency" example:"7.746177ms"`
		Data       any       `json:"data,omitzero"`
		App        string    `json:"app" example:"laukpauk"`
	}
)

func NewResponse(statusCode int) *Response {
	return &Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Timestamp:  time.Now(),
		App:        config.AppName,
	}
}

func (r *Response) SetMessage(message string) *Response {
	if r.StatusCode == fiber.StatusNoContent {
		return r
	}
	r.Message = message
	return r
}

func (r *Response) SetData(data any) *Response {
	if r.StatusCode == fiber.StatusNoContent {
		return r
	}
	r.Data = data
	return r
}

func (r *Response) WriteResponse(c fiber.Ctx) error {
	if r.StatusCode == fiber.StatusNoContent {
		return c.SendStatus(r.StatusCode)
	}
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s %s", c.Method(), c.FullURL())
	r.RequestURL = sb.String()
	r.RequestID = ByteSlice2String(c.Response().Header.Peek(fiber.HeaderXRequestID))
	r.Latency = time.Since(c.RequestCtx().Time()).String()
	return c.Status(r.StatusCode).JSON(r)
}
