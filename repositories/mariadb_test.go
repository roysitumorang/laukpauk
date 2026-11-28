package repositories

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/migrations"
	"github.com/roysitumorang/laukpauk/models"
)

// These tests run against a real MariaDB when TEST_DATABASE_URL (a go-sql-driver/mysql DSN of an empty database) is
// set, as in CI; the migrations need DEFAULT_PASSWORD too.

var (
	testPoolOnce sync.Once
	testPoolDB   *sql.DB
	testPoolErr  error
)

func testPool(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	testPoolOnce.Do(func() {
		ctx := context.Background()
		helper.NewLogger()
		if testPoolDB, testPoolErr = config.CreateDbConnection(ctx, dsn, 4); testPoolErr != nil {
			return
		}
		testPoolErr = migrations.NewMigrator(testPoolDB).Migrate(ctx)
	})
	if testPoolErr != nil {
		t.Fatalf("open and migrate: %v", testPoolErr)
	}
	return testPoolDB
}

func sellerID(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := db.QueryRowContext(context.Background(), "SELECT id FROM users ORDER BY created_at LIMIT 1").Scan(helper.SQLScan(&id)...); err != nil {
		t.Fatalf("seller: %v", err)
	}
	return id
}

// inTx runs fn in a transaction that is always rolled back, so tests leave no rows behind.
func inTx(t *testing.T, db *sql.DB, fn func(tx *sql.Tx)) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	fn(tx)
}

func TestMariaDBMigrationsAreIdempotent(t *testing.T) {
	db := testPool(t)
	migrations.Migrations = map[int64]migrations.Migration{}
	if err := migrations.NewMigrator(db).Migrate(context.Background()); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}

func TestMariaDBTransactionCounterUpsert(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	repo := NewTransactionCounterRepository(db, db)
	window := "test-" + time.Now().Format("150405.000000000")
	now := time.Now()
	first, err := repo.SaveTransactionCounter(ctx, &models.TransactionCounter{TransactionWindow: window, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("first save: %v", err)
	}
	second, err := repo.SaveTransactionCounter(ctx, &models.TransactionCounter{TransactionWindow: window, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	if first.Hits != 1 || second.Hits != 2 || first.ID.Compare(second.ID) != 0 || second.TransactionWindow != window {
		t.Fatalf("got %+v then %+v; want hits 1 then 2 on the same row", first, second)
	}
}

func TestMariaDBProductCategoryUniqueName(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	seller := sellerID(t, db)
	repo := NewProductCategoryRepository(db, db)
	inTx(t, db, func(tx *sql.Tx) {
		now := time.Now()
		category := func(name string) *models.ProductCategory {
			return &models.ProductCategory{
				RecordStatus: models.RecordStatusActive,
				Name:         name,
				Description:  "test",
				Photo:        "test.webp",
				CreatedBy:    seller,
				CreatedAt:    now,
				UpdatedBy:    seller,
				UpdatedAt:    now,
			}
		}
		created, err := repo.CreateProductCategory(ctx, tx, category("Sayur Test 10"))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.ID.Compare(uuid.Nil()) == 0 || created.Name != "Sayur Test 10" || created.CreatedAt.Location() != time.Local {
			t.Fatalf("created = %+v", created)
		}
		if _, err = repo.CreateProductCategory(ctx, tx, category("sayur test 10")); !errors.Is(err, models.ErrUniqueProductCategoryNameViolation) {
			t.Fatalf("duplicate name: err = %v; want ErrUniqueProductCategoryNameViolation", err)
		}
	})
	inTx(t, db, func(tx *sql.Tx) {
		now := time.Now()
		for _, name := range []string{"Item 10", "Item 2", "Item 1"} {
			if _, err := repo.CreateProductCategory(ctx, tx, &models.ProductCategory{
				RecordStatus: models.RecordStatusActive, Name: name, Description: "test", Photo: "test.webp",
				CreatedBy: seller, CreatedAt: now, UpdatedBy: seller, UpdatedAt: now,
			}); err != nil {
				t.Fatalf("create %q: %v", name, err)
			}
		}
		rows, err := tx.QueryContext(ctx, "SELECT name FROM product_categories WHERE name LIKE ? ORDER BY "+sortExpression("name"), helper.SQLArgs("Item %")...)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		var names []string
		for rows.Next() {
			var name string
			if err = rows.Scan(helper.SQLScan(&name)...); err != nil {
				t.Fatalf("scan: %v", err)
			}
			names = append(names, name)
		}
		if strings.Join(names, ",") != "Item 1,Item 2,Item 10" {
			t.Fatalf("natural order = %v", names)
		}
	})
	if _, _, _, err := repo.FindProductCategories(ctx, models.NewProductCategoryFilter().
		WithKeyword("sayur").
		WithSortings(models.Sorting{Field: models.SortingFieldName, Direction: models.SortingDirectionAsc}).
		WithLimit(10)); err != nil {
		t.Fatalf("find: %v", err)
	}
}

func TestMariaDBDeliveryPlanOverlap(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	seller := sellerID(t, db)
	repo := NewDeliveryPlanRepository(db, db)
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	plan := func(weekday models.DeliveryWeekday, from, to int) *models.DeliveryPlan {
		now := time.Now()
		return &models.DeliveryPlan{
			RecordStatus:         models.RecordStatusActive,
			Name:                 "test",
			Weekday:              weekday,
			TransactionWindowMin: day.Add(time.Duration(from) * time.Hour),
			TransactionWindowMax: day.Add(time.Duration(to) * time.Hour),
			DeliveryTimeInterval: models.Duration(time.Hour),
			CreatedBy:            seller,
			CreatedAt:            now,
			UpdatedBy:            seller,
			UpdatedAt:            now,
		}
	}
	inTx(t, db, func(tx *sql.Tx) {
		first, err := repo.CreateDeliveryPlan(ctx, tx, plan(99, 1, 5))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if first.SameDayWindowMin != nil || !first.TransactionWindowMin.Equal(day.Add(time.Hour)) {
			t.Fatalf("created = %+v", first)
		}
		if _, err = repo.CreateDeliveryPlan(ctx, tx, plan(99, 5, 8)); !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
			t.Fatalf("overlapping plan: err = %v", err)
		}
		if _, err = repo.CreateDeliveryPlan(ctx, tx, plan(98, 1, 5)); err != nil {
			t.Fatalf("other weekday: %v", err)
		}
		second, err := repo.CreateDeliveryPlan(ctx, tx, plan(99, 6, 8))
		if err != nil {
			t.Fatalf("next window: %v", err)
		}
		moved := plan(99, 4, 8)
		moved.ID = second.ID
		if _, err = repo.UpdateDeliveryPlan(ctx, tx, moved); !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
			t.Fatalf("update into overlap: err = %v", err)
		}
	})
	if _, _, _, err := repo.FindDeliveryPlans(ctx, models.NewDeliveryPlanFilter().
		WithTransactionDatetime(time.Now()).
		WithSortings(models.Sorting{Field: models.SortingFieldName, Direction: models.SortingDirectionAsc})); err != nil {
		t.Fatalf("find by datetime: %v", err)
	}
	if _, _, _, err := repo.FindDeliveryPlans(ctx, models.NewDeliveryPlanFilter().
		WithTransactionWindow(day, day.Add(time.Hour))); err != nil {
		t.Fatalf("find by window: %v", err)
	}
}

func TestMariaDBPointEarningCapOverlap(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	seller := sellerID(t, db)
	repo := NewPointEarningCapRepository(db, db)
	limit := int64(1_000_000_000_000)
	pointCap := func(minimum, maximum int64) *models.PointEarningCap {
		now := time.Now()
		return &models.PointEarningCap{
			RecordStatus:        models.RecordStatusActive,
			MinimumShoppingCost: minimum,
			MaximumShoppingCost: maximum,
			EarnedPoints:        1,
			CreatedBy:           seller,
			CreatedAt:           now,
			UpdatedBy:           seller,
			UpdatedAt:           now,
		}
	}
	inTx(t, db, func(tx *sql.Tx) {
		created, err := repo.CreatePointEarningCap(ctx, tx, pointCap(limit, limit+100))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.MinimumShoppingCost != limit || created.MaximumShoppingCost != limit+100 {
			t.Fatalf("created = %+v", created)
		}
		if _, err = repo.CreatePointEarningCap(ctx, tx, pointCap(limit+100, limit+200)); !errors.Is(err, models.ErrPointEarningCapShoppingCostOverlapping) {
			t.Fatalf("overlapping cap: err = %v", err)
		}
		if _, err = repo.CreatePointEarningCap(ctx, tx, pointCap(limit+101, limit+200)); err != nil {
			t.Fatalf("adjacent cap: %v", err)
		}
	})
	if _, _, _, err := repo.FindPointEarningCaps(ctx, models.NewPointEarningCapFilter().WithShoppingCost(1, 100)); err != nil {
		t.Fatalf("find: %v", err)
	}
}

func TestMariaDBSetting(t *testing.T) {
	db := testPool(t)
	setting, err := NewSettingRepository(db, db).FindSetting(context.Background())
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(setting.ReferralPointsSlots) == 0 || setting.TransactionAutoWithdrawalWindowMin == nil {
		t.Fatalf("setting = %+v", setting)
	}
}
