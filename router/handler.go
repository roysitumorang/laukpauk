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

	"github.com/goccy/go-json/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/storage/valkey"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/migrations"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/services"
)

type Service struct {
	DbWrite                   *pgxpool.Pool
	Migration                 *migrations.Migration
	KafkaClient               *externals.KafkaClient
	FcmClient                 *externals.FcmClient
	ZenzivaClient             *externals.ZenzivaClient
	RoutesApiClient           *externals.RoutesApiClient
	XenditClient              *externals.XenditClient
	TelegramBotClient         *externals.TelegramBotClient
	Storage                   fiber.Storage
	AdminService              services.AdminService
	BuyerService              services.BuyerService
	ArticleService            services.ArticleService
	BlogService               services.BlogService
	LandingPageService        services.LandingPageService
	LandingPageItemService    services.LandingPageItemService
	MessageService            services.MessageService
	ProductCategoryService    services.ProductCategoryService
	ProductService            services.ProductService
	ProductVariantService     services.ProductVariantService
	TextMessageService        services.TextMessageService
	SettingService            services.SettingService
	DeliveryPlanService       services.DeliveryPlanService
	PointService              services.PointService
	VoucherService            services.VoucherService
	PointEarningCapService    services.PointEarningCapService
	PointSpendingCapService   services.PointSpendingCapService
	TransactionService        services.TransactionService
	JsonWebTokenService       services.JsonWebTokenService
	NotificationTargetService services.NotificationTargetService
	CartService               services.CartService
	TransactionCounterService services.TransactionCounterService
	DeliveryRouteService      services.DeliveryRouteService
	PaymentChannelService     services.PaymentChannelService
	BannerService             services.BannerService
	NotificationService       services.NotificationService
	RecipeService             services.RecipeService
	RecipeGuideService        services.RecipeGuideService
	RecipeItemService         services.RecipeItemService
	CompanyService            services.CompanyService
	SellerService             services.SellerService
	OneTimePasswordService    services.OneTimePasswordService
	PointEarningService       services.PointEarningService
}

func MakeHandler(ctx context.Context) (*Service, error) {
	ctxt := "Router-MakeHandler"
	maxConns64, err := strconv.ParseInt(os.Getenv("DB_MAX_CONNECTIONS"), 10, 32)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrParseInt")
		return nil, err
	}
	if maxConns64 < 1 {
		return nil, errors.New("db: env DB_MAX_CONNECTIONS requires a positive integer")
	}
	maxConns := int32(maxConns64)
	dbRead, err := config.CreateDbConnection(ctx, os.Getenv("DB_READ_URL"), maxConns)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCreateDbConnection")
		return nil, err
	}
	dbWrite, err := config.CreateDbConnection(ctx, os.Getenv("DB_WRITE_URL"), maxConns)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCreateDbConnection")
		return nil, err
	}
	migration := migrations.NewMigration(dbWrite)
	kafkaClient, err := externals.NewKafkaClient(ctx, strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewKafkaClient")
		return nil, err
	}
	if err = kafkaClient.Ping(ctx); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrPing")
		return nil, err
	}
	messages := make([]models.TopicMessage, len(models.SliceTopics))
	for i, topic := range models.SliceTopics {
		messages[i] = models.TopicMessage{
			Topic: models.Topic(topic),
		}
	}
	if err = kafkaClient.Publish(ctx, messages...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrPublish")
		return nil, err
	}
	externalApiCallTimeoutDuration, err := time.ParseDuration(os.Getenv("EXTERNAL_API_CALL_TIMEOUT_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return nil, err
	}
	fcmClient := externals.NewFcmClient(os.Getenv("FIREBASE_PROJECT_ID"), externalApiCallTimeoutDuration)
	zenzivaMessageMaxLength, err := strconv.Atoi(os.Getenv("ZENZIVA_MESSAGE_MAX_LENGTH"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrAtoi")
		return nil, err
	}
	zenzivaClient := externals.NewZenzivaClient(
		os.Getenv("ZENZIVA_BASE_URL"),
		os.Getenv("ZENZIVA_USER_KEY"),
		os.Getenv("ZENZIVA_API_KEY"),
		zenzivaMessageMaxLength,
		externalApiCallTimeoutDuration,
	)
	routesApiClient := externals.NewRoutesApiClient(
		os.Getenv("ROUTES_API_ENDPOINT"),
		os.Getenv("GOOGLE_API_KEY"),
		externalApiCallTimeoutDuration,
	)
	xenditClient := externals.NewXenditClient(
		os.Getenv("XENDIT_BASE_URL"),
		os.Getenv("XENDIT_API_KEY"),
		externalApiCallTimeoutDuration,
	)
	telegramBotClient := externals.NewTelegramBotClient(
		os.Getenv("TELEGRAM_BOT_TOKEN"),
		externalApiCallTimeoutDuration,
	)
	storage := valkey.New(valkey.Config{
		URL: os.Getenv("REDIS_URL"),
	})
	repositories.SetDbWrite(dbWrite)
	adminRepository := repositories.NewAdminRepository(dbRead, dbWrite)
	buyerRepository := repositories.NewBuyerRepository(dbRead, dbWrite)
	articleRepository := repositories.NewArticleRepository(dbRead, dbWrite)
	blogRepository := repositories.NewBlogRepository(dbRead, dbWrite)
	landingPageRepository := repositories.NewLandingPageRepository(dbRead, dbWrite)
	landingPageItemRepository := repositories.NewLandingPageItemRepository(dbRead, dbWrite)
	messageRepository := repositories.NewMessageRepository(dbRead, dbWrite)
	productCategoryRepository := repositories.NewProductCategoryRepository(dbRead, dbWrite)
	productRepository := repositories.NewProductRepository(dbRead, dbWrite)
	productVariantRepository := repositories.NewProductVariantRepository(dbRead, dbWrite)
	textMessageRepository := repositories.NewTextMessageRepository(dbRead, dbWrite)
	settingRepository := repositories.NewSettingRepository(dbRead, dbWrite)
	deliveryPlanRepository := repositories.NewDeliveryPlanRepository(dbRead, dbWrite)
	pointRepository := repositories.NewPointRepository(dbRead, dbWrite)
	voucherRepository := repositories.NewVoucherRepository(dbRead, dbWrite)
	pointEarningCapRepository := repositories.NewPointEarningCapRepository(dbRead, dbWrite)
	pointSpendingCapRepository := repositories.NewPointSpendingCapRepository(dbRead, dbWrite)
	transactionRepository := repositories.NewTransactionRepository(dbRead, dbWrite)
	jsonWebTokenRepository := repositories.NewJsonWebTokenRepository(dbRead, dbWrite)
	notificationTargetRepository := repositories.NewNotificationTargetRepository(dbRead, dbWrite)
	cartRepository := repositories.NewCartRepository(dbRead, dbWrite)
	transactionCounterRepository := repositories.NewTransactionCounterRepository(dbRead, dbWrite)
	deliveryRouteRepository := repositories.NewDeliveryRouteRepository(dbRead, dbWrite)
	paymentChannelRepository := repositories.NewPaymentChannelRepository(dbRead, dbWrite)
	bannerRepository := repositories.NewBannerRepository(dbRead, dbWrite)
	notificationRepository := repositories.NewNotificationRepository(dbRead, dbWrite)
	recipeRepository := repositories.NewRecipeRepository(dbRead, dbWrite)
	recipeGuideRepository := repositories.NewRecipeGuideRepository(dbRead, dbWrite)
	recipeItemRepository := repositories.NewRecipeItemRepository(dbRead, dbWrite)
	companyRepository := repositories.NewCompanyRepository(dbRead, dbWrite)
	sellerRepository := repositories.NewSellerRepository(dbRead, dbWrite)
	oneTimePasswordRepository := repositories.NewOneTimePasswordRepository(dbRead, dbWrite)
	pointEarningRepository := repositories.NewPointEarningRepository(dbRead, dbWrite)
	adminService := services.NewAdminService(adminRepository)
	buyerService := services.NewBuyerService(buyerRepository)
	articleService := services.NewArticleService(articleRepository)
	blogService := services.NewBlogService(blogRepository)
	landingPageService := services.NewLandingPageService(landingPageRepository)
	landingPageItemService := services.NewLandingPageItemService(landingPageItemRepository)
	messageService := services.NewMessageService(messageRepository)
	productCategoryService := services.NewProductCategoryService(productCategoryRepository)
	productService := services.NewProductService(productRepository)
	productVariantService := services.NewProductVariantService(productVariantRepository, productRepository)
	textMessageService := services.NewTextMessageService(textMessageRepository, kafkaClient, zenzivaClient)
	settingService := services.NewSettingService(settingRepository)
	deliveryPlanService := services.NewDeliveryPlanService(deliveryPlanRepository)
	pointService := services.NewPointService(
		pointRepository,
		buyerRepository,
		transactionRepository,
	)
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
		companyRepository,
		voucherRepository,
		pointSpendingCapRepository,
		pointRepository,
		productVariantRepository,
		paymentChannelRepository,
		sellerRepository,
		notificationRepository,
		xenditClient,
	)
	jsonWebTokenService := services.NewJsonWebTokenService(jsonWebTokenRepository)
	notificationTargetService := services.NewNotificationTargetService(notificationTargetRepository)
	cartService := services.NewCartService(cartRepository)
	transactionCounterService := services.NewTransactionCounterService(transactionCounterRepository)
	deliveryRouteService := services.NewDeliveryRouteService(deliveryRouteRepository)
	paymentChannelService := services.NewPaymentChannelService(paymentChannelRepository)
	bannerService := services.NewBannerService(bannerRepository)
	notificationService := services.NewNotificationService(
		notificationRepository,
		notificationTargetRepository,
		kafkaClient,
		fcmClient,
	)
	recipeService := services.NewRecipeService(recipeRepository)
	recipeGuideService := services.NewRecipeGuideService(recipeGuideRepository)
	recipeItemService := services.NewRecipeItemService(recipeItemRepository)
	companyService := services.NewCompanyService(companyRepository)
	sellerService := services.NewSellerService(sellerRepository)
	oneTimePasswordService := services.NewOneTimePasswordService(oneTimePasswordRepository)
	pointEarningService := services.NewPointEarningService(pointEarningRepository)
	return &Service{
		DbWrite:                   dbWrite,
		Migration:                 migration,
		KafkaClient:               kafkaClient,
		FcmClient:                 fcmClient,
		ZenzivaClient:             zenzivaClient,
		RoutesApiClient:           routesApiClient,
		XenditClient:              xenditClient,
		TelegramBotClient:         telegramBotClient,
		Storage:                   storage,
		AdminService:              adminService,
		BuyerService:              buyerService,
		ArticleService:            articleService,
		BlogService:               blogService,
		LandingPageService:        landingPageService,
		LandingPageItemService:    landingPageItemService,
		MessageService:            messageService,
		ProductCategoryService:    productCategoryService,
		ProductService:            productService,
		ProductVariantService:     productVariantService,
		TextMessageService:        textMessageService,
		SettingService:            settingService,
		DeliveryPlanService:       deliveryPlanService,
		PointService:              pointService,
		VoucherService:            voucherService,
		PointEarningCapService:    pointEarningCapService,
		PointSpendingCapService:   pointSpendingCapService,
		TransactionService:        transactionService,
		JsonWebTokenService:       jsonWebTokenService,
		NotificationTargetService: notificationTargetService,
		CartService:               cartService,
		TransactionCounterService: transactionCounterService,
		DeliveryRouteService:      deliveryRouteService,
		PaymentChannelService:     paymentChannelService,
		BannerService:             bannerService,
		NotificationService:       notificationService,
		RecipeService:             recipeService,
		RecipeGuideService:        recipeGuideService,
		RecipeItemService:         recipeItemService,
		CompanyService:            companyService,
		SellerService:             sellerService,
		OneTimePasswordService:    oneTimePasswordService,
		PointEarningService:       pointEarningService,
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
		fetches := q.KafkaClient.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			err := errors.New(fmt.Sprint(errs))
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrPollFetches")
			return err
		}
		records := fetches.Records()
		if err := q.KafkaClient.CommitUncommittedOffsets(ctx); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCommitUncommittedOffsets")
			return err
		}
		q.KafkaClient.AllowRebalance()
		for _, record := range records {
			now := time.Now()
			var message models.TopicMessage
			err := json.Unmarshal(record.Value, &message)
			if err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrUnmarshal")
			} else {
				if err = q.TextMessageService.ConsumeMessage(ctx, message.Topic, message); err != nil {
					helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrConsumeMessage")
				}
				if err = q.NotificationService.ConsumeMessage(ctx, message.Topic, message); err != nil {
					helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrConsumeMessage")
				}
			}
			duration := time.Since(now)
			helper.Log(ctx, slog.LevelInfo, helper.Sprintf("consumed message on topic %s[%d]@%d: %s in %s", record.Topic, record.Partition, record.Offset, record.Value, duration.String()), ctxt, "")
		}
	}
}
