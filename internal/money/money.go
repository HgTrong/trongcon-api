package money

import (
	"strconv"
	"strings"
)

// DefaultCurrency is the only supported catalog/payment currency.
const DefaultCurrency = "VND"

// Normalize forces VND (TrongCon no longer offers USD pricing).
func Normalize(currency string) string {
	_ = currency
	return DefaultCurrency
}

// IsVND reports whether the currency is VND (after normalize, always true for new writes).
func IsVND(currency string) bool {
	return strings.EqualFold(strings.TrimSpace(currency), "VND") || strings.TrimSpace(currency) == ""
}

// FormatVND renders an amount the way the frontend does via
// `Intl.NumberFormat("vi-VN", {style:"currency", currency:"VND"})` — dot
// thousands separators, no decimals, trailing đ sign — so emails read the
// same as the UI (e.g. "1.500.000 ₫").
func FormatVND(amount float64) string {
	n := int64(amount + 0.5)
	neg := n < 0
	if neg {
		n = -n
	}
	digits := strconv.FormatInt(n, 10)
	var groups []string
	for len(digits) > 3 {
		groups = append([]string{digits[len(digits)-3:]}, groups...)
		digits = digits[:len(digits)-3]
	}
	groups = append([]string{digits}, groups...)
	out := strings.Join(groups, ".")
	if neg {
		out = "-" + out
	}
	return out + " ₫"
}
