// Package money provides decimal-safe parsing and formatting of monetary
// values as integer minor units (scale 2) backed by math/big. Floating point
// is never used for authoritative money (INV-004).
package money

import (
	"math/big"
	"strings"
)

const scale = int64(100)

// ParseMinorUnits converts a decimal string like "1500000.00" into minor
// units (150000000). At most two decimal digits are accepted; an empty or
// unparseable input is rejected.
func ParseMinorUnits(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errInvalid()
	}
	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}
	if s == "" {
		return 0, errInvalid()
	}
	intPart := s
	fracPart := ""
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		intPart = s[:dot]
		fracPart = s[dot+1:]
		if len(fracPart) > 2 {
			return 0, errInvalid()
		}
	}
	if intPart == "" {
		intPart = "0"
	}
	whole, ok := new(big.Int).SetString(intPart, 10)
	if !ok {
		return 0, errInvalid()
	}
	whole = whole.Mul(whole, big.NewInt(scale))
	var frac *big.Int
	switch len(fracPart) {
	case 2:
		frac, ok = new(big.Int).SetString(fracPart, 10)
	case 1:
		frac, ok = new(big.Int).SetString(fracPart+"0", 10)
	default:
		frac = big.NewInt(0)
		ok = true
	}
	if !ok {
		return 0, errInvalid()
	}
	sum := new(big.Int).Add(whole, frac)
	if neg {
		sum = sum.Neg(sum)
	}
	if !sum.IsInt64() {
		return 0, errInvalid()
	}
	return sum.Int64(), nil
}

// FormatMinorUnits formats minor units back to a two-decimal string.
func FormatMinorUnits(minor int64) string {
	neg := minor < 0
	if neg {
		minor = -minor
	}
	whole := minor / scale
	frac := minor % scale
	out := ""
	if neg {
		out = "-"
	}
	out += itoa(whole) + "." + pad(frac)
	return out
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	return string(buf)
}

func pad(v int64) string {
	if v < 10 {
		return "0" + itoa(v)
	}
	return itoa(v)
}

type invalidError struct{}

func (invalidError) Error() string { return "invalid money amount" }

func errInvalid() error { return invalidError{} }