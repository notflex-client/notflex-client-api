package helpers

import (
	"net/http"
	"strconv"
	"strings"
)

func PageParams(r *http.Request) (page, limit, offset int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit = 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	offset = (page - 1) * limit
	return
}

func SearchParam(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("q"))
}
