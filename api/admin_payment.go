package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func AdminListPayment(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminListPayment"}

	pageParam := strings.TrimSpace(r.URL.Query().Get("page"))
	pageSizeParam := strings.TrimSpace(r.URL.Query().Get("pageSize"))
	statusParam := strings.TrimSpace(r.URL.Query().Get("status"))

	page := helpers.StringToInt64(pageParam, 1)
	pageSize := helpers.StringToInt64(pageSizeParam, 20)
	if pageSize < 1 || pageSize > 100 {
		HandleResponseError(w, r, NewBadRequestError("InvalidPageSize", logParams...))
		return
	}

	query := database.DB.WithContext(r.Context()).Model(&models.Payment{})
	if statusParam != "" {
		query = query.Where("status = ?", statusParam)
	}

	itemCount := int64(0)
	err := query.Count(&itemCount).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting payments", err, logParams...))
		return
	}

	offset := (page - 1) * pageSize
	payments := make([]models.Payment, 0, pageSize)
	err = query.
		Preload("Subscription.Plan").
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&payments).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing payments", err, logParams...))
		return
	}

	pageCount := (itemCount + pageSize - 1) / pageSize
	json.NewEncoder(w).Encode(map[string]any{
		"items":     payments,
		"page":      page,
		"itemCount": itemCount,
		"pageCount": pageCount,
	})
}
