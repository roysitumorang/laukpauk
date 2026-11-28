package repositories

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/models"
)

// fillFilter sets every field of a filter that a query can use: one sample value in each slice, "x" in each string,
// 1 in each number, the time now and flag in each boolean, so the query builder emits all its optional conditions at
// once. Paging, sorting and URL fields are left alone.
func fillFilter(v reflect.Value, flag bool) {
	t := v.Type()
	for i := range t.NumField() {
		f := v.Field(i)
		name := t.Field(i).Name
		if !f.CanSet() {
			continue
		}
		switch name {
		case "Sortings", "MapSortings", "Limit", "Page", "UrlValues", "PaginationURL", "NextRowNo":
			continue
		case "ProductTypes":
			// nothing sets it, and products has no product_type column (on PostgreSQL either)
			continue
		}
		fillValue(f, flag)
	}
}

func fillValue(f reflect.Value, flag bool) {
	switch {
	case f.Type() == reflect.TypeFor[uuid.UUID]():
		f.Set(reflect.ValueOf(uuid.Must(uuid.NewV7())))
		return
	case f.Type() == reflect.TypeFor[time.Time]():
		f.Set(reflect.ValueOf(time.Now()))
		return
	}
	switch f.Kind() {
	case reflect.String:
		f.SetString("x")
	case reflect.Bool:
		f.SetBool(flag)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f.SetUint(1)
	case reflect.Slice:
		s := reflect.MakeSlice(f.Type(), 1, 1)
		fillValue(s.Index(0), flag)
		f.Set(s)
	case reflect.Pointer:
		p := reflect.New(f.Type().Elem())
		fillValue(p.Elem(), flag)
		f.Set(p)
	case reflect.Struct:
		fillFilter(f, flag)
	}
}

// TestMariaDBQueries runs every repository's list and single-row queries with all filter conditions on, to check
// their SQL against MariaDB. The rows found don't matter.
func TestMariaDBQueries(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	type finder struct {
		name   string
		filter func() any
		list   func(any) error
		one    func(any) error
	}
	// only errors from MariaDB, the driver or the scanning count: "not found" and other model errors don't
	check := func(err error) error {
		if err == nil || errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) || strings.Contains(err.Error(), "sql:") ||
			strings.Contains(err.Error(), "database:") || strings.Contains(err.Error(), "converting") {
			return err
		}
		return nil
	}
	finders := []finder{
		{"banners", func() any { return models.NewBannerFilter() },
			func(f any) error {
				_, _, _, err := NewBannerRepository(db, db).FindBanners(ctx, f.(*models.BannerFilter))
				return err
			},
			func(f any) error {
				_, err := NewBannerRepository(db, db).FindBanner(ctx, f.(*models.BannerFilter))
				return err
			}},
		{"carts", func() any { return models.NewCartFilter() },
			func(f any) error {
				_, _, _, err := NewCartRepository(db, db).FindCarts(ctx, f.(*models.CartFilter))
				return err
			},
			func(f any) error {
				_, err := NewCartRepository(db, db).FindCart(ctx, f.(*models.CartFilter))
				return err
			}},
		{"delivery addresses", func() any { return models.NewDeliveryAddressFilter() },
			func(f any) error {
				_, _, _, err := NewDeliveryAddressRepository(db, db).FindDeliveryAddresses(ctx, f.(*models.DeliveryAddressFilter))
				return err
			},
			func(f any) error {
				_, err := NewDeliveryAddressRepository(db, db).FindDeliveryAddress(ctx, f.(*models.DeliveryAddressFilter))
				return err
			}},
		{"delivery plans", func() any { return models.NewDeliveryPlanFilter() },
			func(f any) error {
				_, _, _, err := NewDeliveryPlanRepository(db, db).FindDeliveryPlans(ctx, f.(*models.DeliveryPlanFilter))
				return err
			},
			func(f any) error {
				_, err := NewDeliveryPlanRepository(db, db).FindDeliveryPlan(ctx, f.(*models.DeliveryPlanFilter))
				return err
			}},
		{"json web tokens", func() any { return models.NewJsonWebTokenFilter() },
			func(f any) error {
				_, _, _, err := NewJsonWebTokenRepository(db, db).FindJsonWebTokens(ctx, f.(*models.JsonWebTokenFilter))
				return err
			},
			func(f any) error {
				_, err := NewJsonWebTokenRepository(db, db).FindJsonWebToken(ctx, f.(*models.JsonWebTokenFilter))
				return err
			}},
		{"notifications", func() any { return models.NewNotificationFilter() },
			func(f any) error {
				_, _, _, err := NewNotificationRepository(db, db).FindNotifications(ctx, f.(*models.NotificationFilter))
				return err
			},
			func(f any) error {
				_, err := NewNotificationRepository(db, db).FindNotification(ctx, f.(*models.NotificationFilter))
				return err
			}},
		{"notification targets", func() any { return models.NewNotificationTargetFilter() },
			func(f any) error {
				_, _, _, err := NewNotificationTargetRepository(db, db).FindNotificationTargets(ctx, f.(*models.NotificationTargetFilter))
				return err
			},
			func(f any) error {
				_, err := NewNotificationTargetRepository(db, db).FindNotificationTarget(ctx, f.(*models.NotificationTargetFilter))
				return err
			}},
		{"payment channels", func() any { return models.NewPaymentChannelFilter() },
			func(f any) error {
				_, _, _, err := NewPaymentChannelRepository(db, db).FindPaymentChannels(ctx, f.(*models.PaymentChannelFilter))
				return err
			},
			func(f any) error {
				_, err := NewPaymentChannelRepository(db, db).FindPaymentChannel(ctx, f.(*models.PaymentChannelFilter))
				return err
			}},
		{"point earning caps", func() any { return models.NewPointEarningCapFilter() },
			func(f any) error {
				_, _, _, err := NewPointEarningCapRepository(db, db).FindPointEarningCaps(ctx, f.(*models.PointEarningCapFilter))
				return err
			},
			func(f any) error {
				_, err := NewPointEarningCapRepository(db, db).FindPointEarningCap(ctx, f.(*models.PointEarningCapFilter))
				return err
			}},
		{"point earnings", func() any { return models.NewPointEarningFilter() },
			func(f any) error {
				_, _, _, err := NewPointEarningRepository(db, db).FindPointEarnings(ctx, f.(*models.PointEarningFilter))
				return err
			},
			func(f any) error {
				_, err := NewPointEarningRepository(db, db).FindPointEarning(ctx, f.(*models.PointEarningFilter))
				return err
			}},
		{"points", func() any { return models.NewPointFilter() },
			func(f any) error {
				_, _, _, err := NewPointRepository(db, db).FindPoints(ctx, f.(*models.PointFilter))
				return err
			},
			func(f any) error {
				_, err := NewPointRepository(db, db).FindPoint(ctx, f.(*models.PointFilter))
				return err
			}},
		{"point spending caps", func() any { return models.NewPointSpendingCapFilter() },
			func(f any) error {
				_, _, _, err := NewPointSpendingCapRepository(db, db).FindPointSpendingCaps(ctx, f.(*models.PointSpendingCapFilter))
				return err
			},
			func(f any) error {
				_, err := NewPointSpendingCapRepository(db, db).FindPointSpendingCap(ctx, f.(*models.PointSpendingCapFilter))
				return err
			}},
		{"product categories", func() any { return models.NewProductCategoryFilter() },
			func(f any) error {
				_, _, _, err := NewProductCategoryRepository(db, db).FindProductCategories(ctx, f.(*models.ProductCategoryFilter))
				return err
			},
			func(f any) error {
				_, err := NewProductCategoryRepository(db, db).FindProductCategory(ctx, f.(*models.ProductCategoryFilter))
				return err
			}},
		{"products", func() any { return models.NewProductFilter() },
			func(f any) error {
				_, _, _, err := NewProductRepository(db, db).FindProducts(ctx, f.(*models.ProductFilter))
				return err
			},
			func(f any) error {
				_, err := NewProductRepository(db, db).FindProduct(ctx, f.(*models.ProductFilter))
				return err
			}},
		{"product variants", func() any { return models.NewProductVariantFilter() },
			func(f any) error {
				_, _, _, err := NewProductVariantRepository(db, db).FindProductVariants(ctx, f.(*models.ProductVariantFilter))
				return err
			},
			func(f any) error {
				_, err := NewProductVariantRepository(db, db).FindProductVariant(ctx, f.(*models.ProductVariantFilter))
				return err
			}},
		{"promotions", func() any { return models.NewPromotionFilter() },
			func(f any) error {
				_, _, _, err := NewPromotionRepository(db, db).FindPromotions(ctx, f.(*models.PromotionFilter))
				return err
			},
			func(f any) error {
				_, err := NewPromotionRepository(db, db).FindPromotion(ctx, f.(*models.PromotionFilter))
				return err
			}},
		{"recipe guides", func() any { return models.NewRecipeGuideFilter() },
			func(f any) error {
				_, _, _, err := NewRecipeGuideRepository(db, db).FindRecipeGuides(ctx, f.(*models.RecipeGuideFilter))
				return err
			},
			func(f any) error {
				_, err := NewRecipeGuideRepository(db, db).FindRecipeGuide(ctx, f.(*models.RecipeGuideFilter))
				return err
			}},
		{"recipe items", func() any { return models.NewRecipeItemFilter() },
			func(f any) error {
				_, _, _, err := NewRecipeItemRepository(db, db).FindRecipeItems(ctx, f.(*models.RecipeItemFilter))
				return err
			},
			func(f any) error {
				_, err := NewRecipeItemRepository(db, db).FindRecipeItem(ctx, f.(*models.RecipeItemFilter))
				return err
			}},
		{"recipes", func() any { return models.NewRecipeFilter() },
			func(f any) error {
				_, _, _, err := NewRecipeRepository(db, db).FindRecipes(ctx, f.(*models.RecipeFilter))
				return err
			},
			func(f any) error {
				_, err := NewRecipeRepository(db, db).FindRecipe(ctx, f.(*models.RecipeFilter))
				return err
			}},
		{"text messages", func() any { return models.NewTextMessageFilter() },
			func(f any) error {
				_, _, _, err := NewTextMessageRepository(db, db).FindTextMessages(ctx, f.(*models.TextMessageFilter))
				return err
			},
			func(f any) error {
				_, err := NewTextMessageRepository(db, db).FindTextMessage(ctx, f.(*models.TextMessageFilter))
				return err
			}},
		{"transactions", func() any { return models.NewTransactionFilter() },
			func(f any) error {
				_, _, _, err := NewTransactionRepository(db, db).FindTransactions(ctx, f.(*models.TransactionFilter))
				return err
			},
			func(f any) error {
				_, err := NewTransactionRepository(db, db).FindTransaction(ctx, f.(*models.TransactionFilter))
				return err
			}},
		{"users", func() any { return models.NewUserFilter() },
			func(f any) error {
				_, _, _, err := NewUserRepository(db, db).FindUsers(ctx, f.(*models.UserFilter))
				return err
			},
			func(f any) error {
				_, err := NewUserRepository(db, db).FindUser(ctx, f.(*models.UserFilter))
				return err
			}},
		{"vouchers", func() any { return models.NewVoucherFilter() },
			func(f any) error {
				_, _, _, err := NewVoucherRepository(db, db).FindVouchers(ctx, f.(*models.VoucherFilter))
				return err
			},
			func(f any) error {
				_, err := NewVoucherRepository(db, db).FindVoucher(ctx, f.(*models.VoucherFilter))
				return err
			}},
	}
	for _, fd := range finders {
		t.Run(fd.name, func(t *testing.T) {
			// no conditions, then every condition with booleans off and on
			for _, mode := range []string{"plain", "false", "true"} {
				for _, run := range []func(any) error{fd.list, fd.one} {
					f := fd.filter()
					if mode != "plain" {
						fillFilter(reflect.ValueOf(f).Elem(), mode == "true")
					}
					if err := check(run(f)); err != nil {
						t.Errorf("%s filter: %v", mode, err)
					}
				}
			}
		})
	}
	trx := NewTransactionRepository(db, db)
	if _, err := trx.FindVouchersUsages(ctx, uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())); err != nil {
		t.Errorf("vouchers usages: %v", err)
	}
	if _, err := trx.FindPromotionsUsages(ctx, uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())); err != nil {
		t.Errorf("promotions usages: %v", err)
	}
}

// TestMariaDBInfiniteScrollProducts pages through products with ROW_NUMBER() and FETCH FIRST ? ROWS ONLY.
func TestMariaDBInfiniteScrollProducts(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	seller := sellerID(t, db)
	now := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	category, err := NewProductCategoryRepository(db, db).CreateProductCategory(ctx, tx, &models.ProductCategory{
		RecordStatus: models.RecordStatusActive, Name: "Scroll Test", Description: "test", Photo: "test.webp",
		CreatedBy: seller, CreatedAt: now, UpdatedBy: seller, UpdatedAt: now,
	})
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("category: %v", err)
	}
	for _, name := range []string{"Scroll 1", "Scroll 2", "Scroll 3"} {
		if _, err = NewProductRepository(db, db).CreateProduct(ctx, tx, &models.Product{
			RecordStatus: models.RecordStatusActive, ProductCategoryID: category.ID, ProductCategoryName: category.Name,
			Name: name, Photo: "test.webp", Description: "test", ReadyStock: true,
			CreatedBy: seller, CreatedAt: now, UpdatedBy: seller, UpdatedAt: now,
		}); err != nil {
			_ = tx.Rollback()
			t.Fatalf("product %q: %v", name, err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	next := int64(2)
	products, _, err := NewProductRepository(db, db).InfiniteScrollProducts(ctx, models.NewProductFilter().
		WithProductCategoryIDs(category.ID).
		WithNextRowNo(&next).
		WithSortings(models.Sorting{Field: models.SortingFieldName, Direction: models.SortingDirectionAsc}).
		WithLimit(1))
	if err != nil {
		t.Fatalf("scroll: %v", err)
	}
	if len(products) != 1 || products[0].Name != "Scroll 2" {
		t.Fatalf("products = %+v; want only Scroll 2", products)
	}
}
