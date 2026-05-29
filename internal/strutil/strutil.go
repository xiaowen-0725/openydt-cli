// Package strutil holds small string helpers shared across the CLI and its tooling.
package strutil

import (
	"regexp"
	"strings"
)

// Kebab converts camelCase/PascalCase to kebab-case (getParkFee -> get-park-fee).
func Kebab(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i, r := range rs {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := rs[i-1]
				nextLower := i+1 < len(rs) && rs[i+1] >= 'a' && rs[i+1] <= 'z'
				if (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') || ((prev >= 'A' && prev <= 'Z') && nextLower) {
					b.WriteByte('-')
				}
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Clip truncates s to at most n runes, appending an ellipsis when cut.
func Clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

var enumItemRe = regexp.MustCompile(`(?:^|[,，(（：:、])\s*(-?\d+)\s*([^,，)）、]+)`)

// ParseEnum extracts "0其他,1蓝色,2黄色" style enum options from a description,
// returning ["0 其他","1 蓝色", ...]; nil when the text is not enum-like.
func ParseEnum(desc string) []string {
	if desc == "" || !strings.ContainsAny(desc, "0123456789") {
		return nil
	}
	ms := enumItemRe.FindAllStringSubmatch(desc, -1)
	if len(ms) < 2 { // need ≥2 numbered options to look like an enum
		return nil
	}
	var out []string
	for _, m := range ms {
		label := strings.Trim(strings.TrimSpace(m[2]), "：: 。.")
		if label == "" || len([]rune(label)) > 12 { // long text → prose, not a label
			continue
		}
		out = append(out, m[1]+" "+label)
	}
	if len(out) < 2 {
		return nil
	}
	return out
}
