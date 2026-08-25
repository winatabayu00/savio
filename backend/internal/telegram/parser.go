package telegram

import (
	"regexp"
	"strconv"
	"strings"
)

var amountToken = regexp.MustCompile(`(\d[\d.,]*)\s*(rb|ribu|jt|juta)?`)

// parseExpense extracts an amount from the end of a free-form recap message and
// returns the leftover text as the description. Examples:
//
//	"chocolate hazelnut dutch 24000"      → 24000.00
//	"grab food 24.500"                    → 24500.00
//	"pulsa 50rb" / "token listrik 45 ribu" → 50.000 / 45.000
//	"kopi 2.5jt"                          → 2500000.00
//
// The amount is the last number-like token, so inline numbers in the description
// ("2 botol susu 50000") never win. A bare number without rb/jt must be at least
// 1000 to be treated as an amount ("2 botol susu" alone stays a description).
func parseExpense(text string) (desc string, amountMinor int64, ok bool) {
	suffix := map[string]int64{"rb": 1_000, "ribu": 1_000, "jt": 1_000_000, "juta": 1_000_000}
	matches := amountToken.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return "", 0, false
	}
	last := matches[len(matches)-1]
	numStr := text[last[2]:last[3]]
	intStr, fracStr := splitNum(numStr)
	suf := ""
	if last[4] > 0 {
		suf = strings.ToLower(text[last[4]:last[5]])
	}
	mult, hasSuffix := suffix[suf]
	if !hasSuffix {
		mult = 1
	}
	iv, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil || (iv == 0 && intStr != "0") {
		return "", 0, false
	}
	if !hasSuffix && iv < 1000 {
		return "", 0, false
	}
	fracMinor := int64(0)
	if fracStr != "" {
		fv, err := strconv.ParseInt(fracStr, 10, 64)
		if err == nil {
			fracMinor = fv * 100 / pow10(len(fracStr))
		}
	}
	amountMinor = iv*100*mult + fracMinor*mult
	desc = strings.TrimSpace(text[:last[0]])
	return desc, amountMinor, true
}

// splitNum splits a number that may use "," or "." as thousand separators
// ("24.500", "1.000.000") and a trailing fraction of at most two digits ("2.5")
// into its integer and fractional digit strings.
func splitNum(s string) (intStr, fracStr string) {
	normalized := strings.ReplaceAll(s, ",", ".")
	parts := strings.Split(normalized, ".")
	for i, p := range parts {
		if i == len(parts)-1 && len(parts) > 1 && len(p) <= 2 {
			fracStr = p
			continue
		}
		intStr += p
	}
	return intStr, fracStr
}

func pow10(n int) int64 {
	r := int64(1)
	for i := 0; i < n; i++ {
		r *= 10
	}
	return r
}
