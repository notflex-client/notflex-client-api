package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func AdminCreateTag(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateTag"}

	var body struct {
		Name string `json:"name" validate:"required,max=50"`
		Slug string `json:"slug" validate:"required,max=80"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	body.Slug = strings.ToLower(strings.TrimSpace(body.Slug))

	var count int64
	err = database.DB.WithContext(r.Context()).Model(&models.Tag{}).
		Where("slug = ?", body.Slug).Count(&count).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("checking tag uniqueness", err, logParams...))
		return
	}
	if count > 0 {
		HandleResponseError(w, r, NewValidationError(map[string]string{"slug": helpers.Translate(r.Context(), "TagSlugExisted")}))
		return
	}

	tag := models.Tag{Name: body.Name, Slug: body.Slug}
	err = database.DB.WithContext(r.Context()).Create(&tag).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating tag", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

func AdminUpdateTag(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminUpdateTag"}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidID", logParams...))
		return
	}

	var body struct {
		Name string `json:"name" validate:"required,max=50"`
		Slug string `json:"slug" validate:"required,max=80"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	var tag models.Tag
	err = database.DB.WithContext(r.Context()).Where("id = ?", id).First(&tag).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("TagNotFound", logParams...))
		return
	}

	body.Slug = strings.ToLower(strings.TrimSpace(body.Slug))
	var count int64
	err = database.DB.WithContext(r.Context()).Model(&models.Tag{}).
		Where("slug = ? AND id <> ?", body.Slug, id).Count(&count).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("checking tag uniqueness", err, logParams...))
		return
	}
	if count > 0 {
		HandleResponseError(w, r, NewValidationError(map[string]string{"slug": helpers.Translate(r.Context(), "TagSlugExisted")}))
		return
	}

	tag.Name = body.Name
	tag.Slug = body.Slug
	err = database.DB.WithContext(r.Context()).Save(&tag).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating tag", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(tag)
}

func AdminDeleteTag(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteTag"}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidID", logParams...))
		return
	}

	database.DB.WithContext(r.Context()).Exec("DELETE FROM movie_tags WHERE tag_id = ?", id)

	err = database.DB.WithContext(r.Context()).Where("id = ?", id).Delete(&models.Tag{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting tag", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
