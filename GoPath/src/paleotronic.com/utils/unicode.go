package utils

import (
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
)

func Hex4(ch rune) string {
	s := strings.ToUpper(strconv.FormatInt(int64(ch), 16))
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

// Escape escapes high characters to \U+0000 notation
func Escape(s string) string {
	out := ""
	for _, ch := range s {
		if ch > 255 {
			out = out + "\\u" + Hex4(ch)
		} else {
			out = out + string(ch)
		}
	}
	return out
}

func Unescape(s string) string {

	out := ""
	for i := 0; i < len(s); i++ {
		ch := rune(s[i])

		if ch == '\\' {

			e := i + 6
			if e > len(s) {
				out = out + string(ch)
				continue
			} else {
				chunk := string(s[i:e])
				h := string(chunk[2:])
				ii, err := strconv.ParseInt(h, 16, 64)
				if err != nil {
					out = out + string(ch)
					continue
				}
				out = out + string(rune(ii))
				i = i + 5
			}

		} else {
			out = out + string(ch)
		}
	}

	return out

}

var tform = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func FlattenAccent(s string) string {

	out, _, _ := transform.String(tform, s)
	return out

}
