package helper

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

type (
	FlashMessage struct {
		Data map[string]string
	}
)

const (
	Flash   = "flash"
	danger  = "danger"
	success = "success"
	warning = "warning"
	info    = "info"
)

func NewFlashMessage() *FlashMessage {
	return &FlashMessage{
		Data: map[string]string{},
	}
}

func (f *FlashMessage) Danger(message string) *FlashMessage {
	clear(f.Data)
	f.Data[danger] = message
	return f
}

func (f *FlashMessage) Success(message string) *FlashMessage {
	clear(f.Data)
	f.Data[success] = message
	return f
}

func (f *FlashMessage) Warning(message string) *FlashMessage {
	clear(f.Data)
	f.Data[warning] = message
	return f
}

func (f *FlashMessage) Info(message string) *FlashMessage {
	clear(f.Data)
	f.Data[info] = message
	return f
}

func (f *FlashMessage) Redirect(c fiber.Ctx, sess *session.Session, location string, statusCode ...int) error {
	ctx := c.Context()
	ctxt := "FlashMessage-Redirect"
	sess.Set(Flash, f)
	if err := sess.SaveWithContext(ctx); err != nil {
		Capture(ctx, slog.LevelError, err, ctxt, "ErrSaveWithContext")
	}
	redirect := c.Redirect()
	if len(statusCode) > 0 {
		redirect = redirect.Status(statusCode[0])
	}
	return redirect.To(location)
}

func (f *FlashMessage) Clear(c fiber.Ctx, sess *session.Session) *FlashMessage {
	ctx := c.Context()
	ctxt := "FlashMessage-Clear"
	clear(f.Data)
	sess.Set(Flash, f)
	if err := sess.SaveWithContext(ctx); err != nil {
		Capture(ctx, slog.LevelError, err, ctxt, "ErrSaveWithContext")
	}
	return f
}
