package money

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Savio money rule (AGENTS.md #18, INV-004):
//   - Authoritative amounts are INTEGER MINOR UNITS (int64) stored as BIGINT.
//   - No floating-point arithmetic for authoritative amounts.
//   - Minor units are 1/100 of the base unit (scale 2).
//   - API exchanges decimal strings.

const Scale = 2

// ParseDecimalToMinorUnits converts an API decimal string like "1500000.00"
// into integer minor units (e.g. 150000000). Rejects negative values (Savio
// stores positive amounts; direction comes from transaction type) and values
// with more than 2 decimal places.
func ParseDecimalToMinorUnits(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("amount is required")
	}
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		return 0, fmt.Errorf("amount must be a positive decimal")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format")
	}
	if d.IsNegative() || d.IsZero() {
		return 0, fmt.Errorf("amount must be greater than zero")
	}
	if d.Exponent() < -Scale {
		return 0, fmt.Errorf("amount supports at most %d decimal places", Scale)
	}
	minor := d.Shift(Scale).IntPart()
	if minor > 9_000_000_000_000_000 { // guard: keep within int64 comfort
		return 0, fmt.Errorf("amount exceeds supported range")
	}
	return minor, nil
}

// ParseDecimalMayZero allows zero (used by opening balances / goal amounts).
func ParseDecimalMayZero(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format")
	}
	if d.IsNegative() {
		return 0, fmt.Errorf("amount must not be negative")
	}
	if d.Exponent() < -Scale {
		return 0, fmt.Errorf("amount supports at most %d decimal places", Scale)
	}
	return d.Shift(Scale).IntPart(), nil
}

// MinorUnitsToDecimalString renders minor units back to the API decimal string.
func MinorUnitsToDecimalString(minor int64) string {
	sign := ""
	if minor < 0 {
		sign = "-"
		minor = -minor
	}
	units := minor / 100
	cents := minor % 100
	return fmt.Sprintf("%s%d.%02d", sign, units, cents)
}

// Add / Sub operate on integer minor units without overflow risk for typical
// personal-finance magnitudes.
func Add(a, b int64) int64 { return a + b }
func Sub(a, b int64) int64 { return a - b }