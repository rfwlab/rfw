// Package dom provides helpers for DOM manipulation.
// StyleInline builds an inline CSS style string from property map.
package dom

import "strings"

// StyleInline converts a map of CSS properties into an inline style string.
// Keys and values are concatenated as "key:value" pairs separated by semicolons.
func StyleInline(styles map[string]string) string {
	var b strings.Builder
	first := true
	for k, v := range styles {
		if !first {
			b.WriteByte(';')
		}
		first = false
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(v)
	}
	return b.String()
}
