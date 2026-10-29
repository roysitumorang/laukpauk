package externals

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/goccy/go-json/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/valyala/fasthttp"
)

type TelegramBotClient struct {
	baseURL string
	client  *fasthttp.Client
}

type TelegramBotUpdate struct {
	UpdateID int64               `json:"update_id"`
	Message  *TelegramBotMessage `json:"message"`
}

type TelegramBotMessage struct {
	MessageID int64           `json:"message_id"`
	Text      string          `json:"text"`
	Chat      TelegramBotChat `json:"chat"`
}

type TelegramBotChat struct {
	ID int64 `json:"id"`
}

type TelegramBotGetUpdatesResponse struct {
	OK     bool                `json:"ok"`
	Result []TelegramBotUpdate `json:"result"`
}

type TelegramBotSendMessagePayload struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func NewTelegramBotClient(
	token string,
	timeout time.Duration,
) *TelegramBotClient {
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "https://api.telegram.org/bot%s", token)
	return &TelegramBotClient{
		baseURL: sb.String(),
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

func (q *TelegramBotClient) GetUpdates(ctx context.Context, offset int64) ([]TelegramBotUpdate, error) {
	ctxt := "TelegramBotClient-GetUpdates"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s/getUpdates?timeout=30", q.baseURL)
	if offset > 0 {
		_, _ = fmt.Fprintf(sb, "&offset=%s", strconv.FormatInt(offset, 10))
	}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodGet)
	err := q.client.Do(req, resp)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDo")
		return nil, err
	}
	var data TelegramBotGetUpdatesResponse
	if err = json.Unmarshal(resp.Body(), &data); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	if !data.OK {
		return nil, errors.New("telegram API error response")
	}
	return data.Result, nil
}

func (q *TelegramBotClient) HandleMessage(ctx context.Context, msg *TelegramBotMessage) {
	ctxt := "TelegramBotClient-HandleMessage"
	var replyText string
	switch msg.Text {
	case "/start":
		replyText = "Hello! I am a lightweight bot built entirely with vanilla Go. How can I help you?"
	case "/ping":
		replyText = "Pong!"
	default:
		replyText = fmt.Sprintf("You said: %s", msg.Text)
	}
	if err := q.SendMessage(ctx, msg.Chat.ID, replyText); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSendMessage")
	}
}

func (q *TelegramBotClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	ctxt := "TelegramBotClient-SendMessage"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s/sendMessage", q.baseURL)
	payload, err := json.Marshal(TelegramBotSendMessagePayload{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrMarshal")
		return err
	}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.SetRequestURI(sb.String())
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBodyRaw(payload)
	if resp.StatusCode() != fiber.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	return nil
}
