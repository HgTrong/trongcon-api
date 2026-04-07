package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const maxLen = 200

// FromTitle tạo slug ASCII, lowercase, bỏ dấu, khoảng trắng -> dấu gạch.
func FromTitle(title string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ := transform.String(t, strings.TrimSpace(strings.ToLower(title)))
	var b strings.Builder
	prevHyphen := true
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			if !prevHyphen && b.Len() > 0 {
				b.WriteRune('-')
				prevHyphen = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > maxLen {
		out = out[:maxLen]
		out = strings.TrimRight(out, "-")
	}
	return out
}
