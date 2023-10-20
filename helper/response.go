package helper

import (
	"github.com/gofiber/fiber/v2"
)

const (
	APP = "laukpauk"
)

type (
	Response struct {
		Code    int         `json:"code"`
		Message string      `json:"message,omitempty"`
		Data    interface{} `json:"data,omitempty"`
		App     string      `json:"app"`
	}
)

func NewResponse(code int, msg string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: msg,
		Data:    data,
		App:     APP,
	}
}

func (r *Response) WriteResponse(c *fiber.Ctx) error {
	return c.Status(r.Code).JSON(r)
}
