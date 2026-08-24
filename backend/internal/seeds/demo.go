package seeds

import "context"

// SeedDemo seeds synthetic demo finance data for a development user.
// See internal/seeds/demo.go. Intentionally empty until the finance modules
// (accounts/transactions) exist so seeding reuses the real domain services.
func SeedDemo(ctx context.Context, db interface{}) error {
	return nil
}