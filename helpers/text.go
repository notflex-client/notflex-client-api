package helpers

import (
	"math/rand"
	"unicode"
	"unicode/utf8"
)

func LowerFirstChar(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

func StringToInt64(str string, defaultVal int64) int64 {
	if str == "" {
		return defaultVal
	}
	var n int64
	for _, c := range str {
		if c < '0' || c > '9' {
			return defaultVal
		}
		n = n*10 + int64(c-'0')
	}
	return n
}

func RandomString(n int, charset string) string {
	runes := []rune(charset)
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}
