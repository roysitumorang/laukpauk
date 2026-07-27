package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/storage/valkey"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/consumers"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/migration"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	Service struct {
		DbWrite                   *pgxpool.Pool
		Migration                 *migration.Migration
		KafkaConsumer             *consumers.KafkaConsumer
		FcmConsumer               *consumers.FcmConsumer
		ZenzivaConsumer           *consumers.ZenzivaConsumer
		RoutesApiConsumer         *consumers.RoutesApiConsumer
		XenditConsumer            *consumers.XenditConsumer
		Storage                   fiber.Storage
		AdminService              services.AdminService
		BuyerService              services.BuyerService
		ArticleService            services.ArticleService
		BlogService               services.BlogService
		BlockService              services.BlockService
		MessageService            services.MessageService
		ProductCategoryService    services.ProductCategoryService
		ProductService            services.ProductService
		ProductVariantService     services.ProductVariantService
		SmsMessageService         services.SmsMessageService
		SettingService            services.SettingService
		DeliveryPlanService       services.DeliveryPlanService
		PointService              services.PointService
		VoucherService            services.VoucherService
		PointEarningCapService    services.PointEarningCapService
		PointSpendingCapService   services.PointSpendingCapService
		TransactionService        services.TransactionService
		JsonWebTokenService       services.JsonWebTokenService
		FcmRegistrationService    services.FcmRegistrationService
		CartService               services.CartService
		TransactionCounterService services.TransactionCounterService
		DeliveryRouteService      services.DeliveryRouteService
		PaymentChannelService     services.PaymentChannelService
	}
)

func MakeHandler(ctx context.Context) (*Service, error) {
	ctxt := "Router-MakeHandler"
	dbRead, err := config.GetDbReadOnly(ctx)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrGetDbReadOnly")
		return nil, err
	}
	dbWrite, err := config.GetDbWriteOnly(ctx)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrGetDbWriteOnly")
		return nil, err
	}
	migration := migration.New(dbRead, dbWrite)
	kafkaConsumer, err := consumers.NewKafkaConsumer(ctx, strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewKafkaConsumer")
		return nil, err
	}
	if err = kafkaConsumer.Ping(ctx); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrPing")
		return nil, err
	}
	topics := make([]models.TopicRequest, len(models.SliceTopics))
	for i, topic := range models.SliceTopics {
		topics[i] = models.TopicRequest{
			Topic:    models.Topic(topic),
			Messages: []models.TopicMessage{},
		}
	}
	_ = kafkaConsumer.Publish(ctx, topics...)
	fcmTimeoutDuration, err := time.ParseDuration(os.Getenv("FCM_TIMEOUT_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return nil, err
	}
	fcmConsumer := consumers.NewFcmConsumer(os.Getenv("FIREBASE_PROJECT_ID"), fcmTimeoutDuration)
	zenzivaMessageMaxLength, err := strconv.Atoi(os.Getenv("ZENZIVA_MESSAGE_MAX_LENGTH"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrAtoi")
		return nil, err
	}
	zenzivaTimeoutDuration, err := time.ParseDuration(os.Getenv("ZENZIVA_TIMEOUT_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return nil, err
	}
	zenzivaConsumer := consumers.NewZenzivaConsumer(
		os.Getenv("ZENZIVA_BASE_URL"),
		os.Getenv("ZENZIVA_USER_KEY"),
		os.Getenv("ZENZIVA_API_KEY"),
		zenzivaMessageMaxLength,
		zenzivaTimeoutDuration,
	)
	routesApiTimeoutDuration, err := time.ParseDuration(os.Getenv("ROUTES_API_TIMEOUT_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return nil, err
	}
	routesApiConsumer := consumers.NewRoutesApiConsumer(
		os.Getenv("ROUTES_API_ENDPOINT"),
		os.Getenv("GOOGLE_API_KEY"),
		routesApiTimeoutDuration,
	)
	xenditTimeoutDuration, err := time.ParseDuration(os.Getenv("XENDIT_TIMEOUT_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return nil, err
	}
	xenditConsumer := consumers.NewXenditConsumer(
		os.Getenv("XENDIT_BASE_URL"),
		os.Getenv("XENDIT_API_KEY"),
		xenditTimeoutDuration,
	)
	storage := valkey.New(valkey.Config{
		URL: os.Getenv("REDIS_URL"),
	})
	repositories.SetDbWrite(dbWrite)
	adminRepository := repositories.NewAdminRepository(dbRead, dbWrite)
	buyerRepository := repositories.NewBuyerRepository(dbRead, dbWrite)
	articleRepository := repositories.NewArticleRepository(dbRead, dbWrite)
	blogRepository := repositories.NewBlogRepository(dbRead, dbWrite)
	blockRepository := repositories.NewBlockRepository(dbRead, dbWrite)
	messageRepository := repositories.NewMessageRepository(dbRead, dbWrite)
	productCategoryRepository := repositories.NewProductCategoryRepository(dbRead, dbWrite)
	productRepository := repositories.NewProductRepository(dbRead, dbWrite)
	productVariantRepository := repositories.NewProductVariantRepository(dbRead, dbWrite)
	smsMessageRepository := repositories.NewSmsMessageRepository(dbRead, dbWrite)
	settingRepository := repositories.NewSettingRepository(dbRead, dbWrite)
	deliveryPlanRepository := repositories.NewDeliveryPlanRepository(dbRead, dbWrite)
	pointRepository := repositories.NewPointRepository(dbRead, dbWrite)
	voucherRepository := repositories.NewVoucherRepository(dbRead, dbWrite)
	pointEarningCapRepository := repositories.NewPointEarningCapRepository(dbRead, dbWrite)
	pointSpendingCapRepository := repositories.NewPointSpendingCapRepository(dbRead, dbWrite)
	transactionRepository := repositories.NewTransactionRepository(dbRead, dbWrite)
	jsonWebTokenRepository := repositories.NewJsonWebTokenRepository(dbRead, dbWrite)
	fcmRegistrationRepository := repositories.NewFcmRegistrationRepository(dbRead, dbWrite)
	cartRepository := repositories.NewCartRepository(dbRead, dbWrite)
	transactionCounterRepository := repositories.NewTransactionCounterRepository(dbRead, dbWrite)
	deliveryRouteRepository := repositories.NewDeliveryRouteRepository(dbRead, dbWrite)
	paymentChannelRepository := repositories.NewPaymentChannelRepository(dbRead, dbWrite)
	adminService := services.NewAdminService(adminRepository)
	buyerService := services.NewBuyerService(buyerRepository)
	articleService := services.NewArticleService(articleRepository)
	blogService := services.NewBlogService(blogRepository)
	blockService := services.NewBlockService(blockRepository)
	messageService := services.NewMessageService(messageRepository)
	productCategoryService := services.NewProductCategoryService(productCategoryRepository)
	productService := services.NewProductService(productRepository, productCategoryRepository)
	productVariantService := services.NewProductVariantService(productVariantRepository, productRepository)
	smsMessageService := services.NewSmsMessageService(smsMessageRepository, kafkaConsumer, zenzivaConsumer)
	settingService := services.NewSettingService(settingRepository)
	deliveryPlanService := services.NewDeliveryPlanService(deliveryPlanRepository)
	pointService := services.NewPointService(pointRepository)
	voucherService := services.NewVoucherService(voucherRepository)
	pointEarningCapService := services.NewPointEarningCapService(pointEarningCapRepository)
	pointSpendingCapService := services.NewPointSpendingCapService(pointSpendingCapRepository)
	transactionService := services.NewTransactionService(
		transactionRepository,
		buyerRepository,
		transactionCounterRepository,
		cartRepository,
		deliveryPlanRepository,
		deliveryRouteRepository,
		settingRepository,
		voucherRepository,
		pointSpendingCapRepository,
		pointRepository,
		productVariantRepository,
		paymentChannelRepository,
		xenditConsumer,
	)
	jsonWebTokenService := services.NewJsonWebTokenService(jsonWebTokenRepository)
	fcmRegistrationService := services.NewFcmRegistrationService(fcmRegistrationRepository)
	cartService := services.NewCartService(cartRepository)
	transactionCounterService := services.NewTransactionCounterService(transactionCounterRepository)
	deliveryRouteService := services.NewDeliveryRouteService(deliveryRouteRepository)
	paymentChannelService := services.NewPaymentChannelService(paymentChannelRepository)
	return &Service{
		DbWrite:                   dbWrite,
		Migration:                 migration,
		KafkaConsumer:             kafkaConsumer,
		FcmConsumer:               fcmConsumer,
		ZenzivaConsumer:           zenzivaConsumer,
		RoutesApiConsumer:         routesApiConsumer,
		XenditConsumer:            xenditConsumer,
		Storage:                   storage,
		AdminService:              adminService,
		BuyerService:              buyerService,
		ArticleService:            articleService,
		BlogService:               blogService,
		BlockService:              blockService,
		MessageService:            messageService,
		ProductCategoryService:    productCategoryService,
		ProductService:            productService,
		ProductVariantService:     productVariantService,
		SmsMessageService:         smsMessageService,
		SettingService:            settingService,
		DeliveryPlanService:       deliveryPlanService,
		PointService:              pointService,
		VoucherService:            voucherService,
		PointEarningCapService:    pointEarningCapService,
		PointSpendingCapService:   pointSpendingCapService,
		TransactionService:        transactionService,
		JsonWebTokenService:       jsonWebTokenService,
		FcmRegistrationService:    fcmRegistrationService,
		CartService:               cartService,
		TransactionCounterService: transactionCounterService,
		DeliveryRouteService:      deliveryRouteService,
		PaymentChannelService:     paymentChannelService,
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
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrRecover")
		}
	}()
	for {
		fetches := q.KafkaConsumer.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			err := errors.New(fmt.Sprint(errs))
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrPollFetches")
			return err
		}
		records := fetches.Records()
		if err := q.KafkaConsumer.CommitUncommittedOffsets(ctx); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCommitUncommittedOffsets")
			return err
		}
		q.KafkaConsumer.AllowRebalance()
		for _, record := range records {
			now := time.Now()
			duration := time.Since(now)
			helper.Log(ctx, slog.LevelInfo, helper.Sprintf("consumed message on topic %s[%d]@%d: %s in %s", record.Topic, record.Partition, record.Offset, record.Value, duration.String()), ctxt, "")
		}
	}
}
