package money

import "strings"

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
