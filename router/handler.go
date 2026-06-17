package router

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/storage/valkey"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/external/kafka"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/migration"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/services"
	"go.uber.org/zap"
)

type (
	Service struct {
		DbWrite            *pgxpool.Pool
		Migration          *migration.Migration
		KafkaService       *kafka.KafkaService
		Storage            fiber.Storage
		UserService        services.UserService
		ArticleService     services.ArticleService
		BlogService        services.BlogService
		WebsiteItemService services.WebsiteItemService
	}
)

func MakeHandler(ctx context.Context) (*Service, error) {
	ctxt := "Router-MakeHandler"
	dbRead, err := config.GetDbReadOnly(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGetDbReadOnly")
		return nil, err
	}
	dbWrite, err := config.GetDbWriteOnly(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGetDbWriteOnly")
		return nil, err
	}
	migration := migration.New(dbRead, dbWrite)
	kafkaService, err := kafka.New(ctx, strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrNew")
		return nil, err
	}
	if err = kafkaService.Ping(ctx); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrPing")
		return nil, err
	}
	topics := make([]kafka.Topic, len(models.SliceTopics))
	for i, topic := range models.SliceTopics {
		topics[i] = kafka.Topic{
			Name:     topic,
			Payloads: []map[string]any{},
		}
	}
	_ = kafkaService.Publish(ctx, topics...)
	storage := valkey.New(valkey.Config{
		URL: os.Getenv("REDIS_URL"),
	})
	repositories.SetDbWrite(dbWrite)
	userRepository := repositories.NewUserRepository(dbRead, dbWrite)
	articleRepository := repositories.NewArticleRepository(dbRead, dbWrite)
	blogRepository := repositories.NewBlogRepository(dbRead, dbWrite)
	websiteItemRepository := repositories.NewWebsiteItemRepository(dbRead, dbWrite)
	userService := services.NewUserService(userRepository)
	articleService := services.NewArticleService(articleRepository)
	blogService := services.NewBlogService(blogRepository)
	websiteItemService := services.NewWebsiteItemService(websiteItemRepository)
	return &Service{
		DbWrite:            dbWrite,
		Migration:          migration,
		KafkaService:       kafkaService,
		Storage:            storage,
		UserService:        userService,
		ArticleService:     articleService,
		BlogService:        blogService,
		WebsiteItemService: websiteItemService,
	}, nil
}

func (q *Service) Consume(ctx context.Context) error {
	ctxt := "Router-Consume"
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRecover")
		}
	}()
	for {
		fetches := q.KafkaService.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			err := errors.New(fmt.Sprint(errs))
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrPollFetches")
			return err
		}
		records := fetches.Records()
		if err := q.KafkaService.CommitUncommittedOffsets(ctx); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrCommitUncommittedOffsets")
			return err
		}
		q.KafkaService.AllowRebalance()
		for _, record := range records {
			now := time.Now()
			duration := time.Since(now)
			helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("consumed message on topic %s[%d]@%d: %s in %s", record.Topic, record.Partition, record.Offset, record.Value, duration.String()), ctxt, "")
		}
	}
}
