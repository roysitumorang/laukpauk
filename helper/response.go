package helper

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
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
		Status:     utils.StatusMessage(statusCode),
		Timestamp:  time.Now(),
		App:        config.AppName,
	}
}

func (r *Response) SetMessage(message string) *Response {
	if r.StatusCode != fiber.StatusNoContent {
		r.Message = message
	}
	return r
}

func (r *Response) SetData(data any) *Response {
	if r.StatusCode != fiber.StatusNoContent {
		r.Data = data
	}
	return r
}

func (r *Response) JSON(c fiber.Ctx) error {
	if r.StatusCode == fiber.StatusNoContent {
		return r.SendStatus(c)
	}
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	_, _ = fmt.Fprintf(sb, "%s %s", c.Method(), c.FullURL())
	r.RequestURL = sb.String()
	r.RequestID = ByteSlice2String(c.Response().Header.Peek(fiber.HeaderXRequestID))
	r.Latency = time.Since(c.RequestCtx().Time()).String()
	return c.Status(r.StatusCode).JSON(r)
}

func (r *Response) SendString(c fiber.Ctx, body string) error {
	if r.StatusCode == fiber.StatusNoContent {
		return r.SendStatus(c)
	}
	return c.Status(r.StatusCode).SendString(body)
}

func (r *Response) SendStatus(c fiber.Ctx) error {
	return c.SendStatus(r.StatusCode)
}
