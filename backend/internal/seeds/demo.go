package seeds

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/accounts"
	"github.com/savio/savio/backend/internal/auth"
	"github.com/savio/savio/backend/internal/budgets"
	"github.com/savio/savio/backend/internal/goals"
	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/platform/password"
	"github.com/savio/savio/backend/internal/recurring"
	"github.com/savio/savio/backend/internal/transactions"
	"github.com/savio/savio/backend/internal/users"
)

const (
	demoEmail    = "user@user.com"
	demoPassword = "password"
)

// SeedDemo creates a demo user + workspace and a coherent ~4 month finance
// history via the real domain services (balances stay derived-consistent) so
// the foreground demo story works end to end. Never real personal data.
func SeedDemo(ctx context.Context, db *gorm.DB) error {
	if err := SeedSystemCategories(ctx, db); err != nil {
		return err
	}
	var existing users.User
	if err := db.Where("email = ?", demoEmail).First(&existing).Error; err == nil {
		// Demo credentials must remain usable after source credential changes.
		hash, hashErr := password.Hash(demoPassword)
		if hashErr != nil {
			return fmt.Errorf("hash demo password: %w", hashErr)
		}
		if updateErr := db.WithContext(ctx).Model(&existing).Update("password_hash", hash).Error; updateErr != nil {
			return fmt.Errorf("reset demo password: %w", updateErr)
		}
		fmt.Printf("demo user ready -> email: %s password: %s\n", demoEmail, demoPassword)
		return nil
	}

	cfg := &config.Config{
		AppEnv:          "development",
		JWTSecret:       "demo-seed-demo-seed-demo-seed-demo-seed",
		CSRFSecret:      "demo-csrf-demo-csrf-demo-csrf",
		AIProvider:      "mock",
		AIModel:         "mock",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}
	res, err := auth.NewService(db, cfg).Register(ctx, "Demo User", demoEmail, demoPassword, "seed", "127.0.0.1")
	if err != nil {
		return fmt.Errorf("register demo: %w", err)
	}
	wsID, userID := res.WorkspaceID, res.UserID

	acctSvc := accounts.NewService(db)
	cashID := mustAcct(ctx, acctSvc, wsID, userID, "Daily Wallet", "EWALLET", 2500000)
	bankID := mustAcct(ctx, acctSvc, wsID, userID, "Salary Bank", "BANK", 45000000)

	txSvc := transactions.NewService(db)
	income := mustCatType(ctx, db, "Salary", "INCOME")
	food := mustCatType(ctx, db, "Food & Dining", "EXPENSE")
	travel := mustCatType(ctx, db, "Transportation", "EXPENSE")
	util := mustCatType(ctx, db, "Utilities", "EXPENSE")

	start := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	for m := 0; m < 4; m++ {
		base := start.AddDate(0, m, 0)
		incomeDay := dayOf(base, 15)
		post(ctx, txSvc, wsID, userID, bankID, &income, "INCOME", "15000000.00", incomeDay, "Monthly salary")
		post(ctx, txSvc, wsID, userID, cashID, &food, "EXPENSE", "350000.00", dayOf(base, 13), "Groceries run")
		post(ctx, txSvc, wsID, userID, cashID, &food, "EXPENSE", "150000.00", dayOf(base, 10), "Restaurant dinner")
		post(ctx, txSvc, wsID, userID, cashID, &travel, "EXPENSE", "600000.00", dayOf(base, 12), "Ride hailing")
		post(ctx, txSvc, wsID, userID, bankID, &util, "EXPENSE", "1200000.00", dayOf(base, 14), "Electricity + internet")
		post(ctx, txSvc, wsID, userID, cashID, &food, "EXPENSE", "120000.00", dayOf(base, 8), "Coffee and snacks")
	}

	_, _ = budgets.NewService(db).Create(ctx, wsID, userID, &budgets.CreateInput{
		CategoryID: food, AmountMinor: 6000000,
		PeriodStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
	})

	rec := recurring.NewService(db)
	_, _ = rec.Create(ctx, wsID, userID, &recurring.CreateInput{
		AccountID: cashID, Type: "EXPENSE", AmountMinor: 200000, Frequency: "MONTHLY",
		StartDate: time.Now().UTC(), Description: "Streaming subscriptions",
	})
	_, _ = rec.Create(ctx, wsID, userID, &recurring.CreateInput{
		AccountID: bankID, Type: "INCOME", AmountMinor: 150000000, Frequency: "MONTH_END",
		StartDate: time.Now().UTC(), Description: "Monthly salary",
	})
	target := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	_, _ = goals.NewService(db).Create(ctx, wsID, userID, &goals.CreateInput{
		Name: "Emergency Fund", TargetAmount: 1200000000, CurrentAmount: 250000000,
		TargetDate: &target, Priority: "HIGH",
	})

	fmt.Printf("demo user ready -> email: %s password: %s\n", demoEmail, demoPassword)
	return nil
}

// dayOf returns day-in-month at UTC midnight for the given month.
func dayOf(t time.Time, day int) time.Time {
	return time.Date(t.Year(), t.Month(), day, 0, 0, 0, 0, time.UTC)
}

func mustAcct(ctx context.Context, svc *accounts.Service, wsID, userID uuid.UUID, name, kind string, opening int64) uuid.UUID {
	v, err := svc.Create(ctx, wsID, userID, &accounts.CreateInput{
		Name: name, Type: kind, Currency: "IDR", OpeningBalance: opening,
	})
	if err != nil {
		panic(fmt.Sprintf("seed account %s: %v", name, err))
	}
	return v.ID
}

func mustCatType(ctx context.Context, db *gorm.DB, name, typ string) uuid.UUID {
	var ids []string
	if err := db.WithContext(ctx).Table("categories").
		Where("name = ? AND type = ? AND is_system = TRUE", name, typ).
		Pluck("id", &ids).Error; err != nil || len(ids) == 0 {
		if err == nil {
			err = fmt.Errorf("category %s/%s not found", name, typ)
		}
		panic(fmt.Sprintf("seed category %s: %v", name, err))
	}
	id, err := uuid.Parse(ids[0])
	if err != nil {
		panic(fmt.Sprintf("seed category id: %v", err))
	}
	return id
}

func post(ctx context.Context, svc *transactions.Service, wsID, userID, account uuid.UUID, cat *uuid.UUID, typ, amount string, d time.Time, desc string) {
	minor, _ := parseMoney(amount)
	_, err := svc.Create(ctx, wsID, userID, &transactions.CreateInput{
		AccountID: account, CategoryID: cat, Type: typ, AmountMinor: minor,
		TransactionDate: d, Status: "POSTED", Description: desc, Source: "MANUAL",
	})
	if err != nil {
		panic(fmt.Sprintf("seed transaction %s: %v", desc, err))
	}
}

func parseMoney(s string) (int64, error) {
	return money.ParseMinorUnits(s)
}
