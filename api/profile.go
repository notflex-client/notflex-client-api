package api

import (
	"encoding/json"
	"net/http"

	"notflex_client_api/helpers"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
