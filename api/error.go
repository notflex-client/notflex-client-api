package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"notflex_client_api/helpers"
)

type ResponseError struct {
	Status     int
	Details    interface{}
	MessageTag string
	LogMessage string
	LogLevel   slog.Level
	LogParams  []interface{}
	Error      error
}

func NewUnauthorizedError() ResponseError {
	return ResponseError{Status: 401, MessageTag: "Unauthorized"}
}

func NewForbiddenError() ResponseError {
	return ResponseError{Status: 403, MessageTag: "Forbidden"}
}

func NewValidationError(details interface{}) ResponseError {
	return ResponseError{Status: 422, Details: details}
}

func NewBadRequestError(messageTag string, logParams ...interface{}) ResponseError {
	return ResponseError{
		Status:     400,
		MessageTag: messageTag,
		LogMessage: messageTag,
		LogLevel:   slog.LevelWarn,
		LogParams:  logParams,
	}
}

func NewNotFoundError(messageTag string, logParams ...interface{}) ResponseError {
	return ResponseError{
		Status:     404,
		MessageTag: messageTag,
		LogMessage: messageTag,
		LogLevel:   slog.LevelWarn,
		LogParams:  logParams,
	}
}

func NewInternalServerError(logMessage string, err error, logParams ...interface{}) ResponseError {
	return ResponseError{
		Status:     500,
		LogMessage: logMessage,
		LogLevel:   slog.LevelError,
		Error:      err,
		LogParams:  logParams,
	}
}

func HandleResponseError(w http.ResponseWriter, r *http.Request, respErr ResponseError) {
	response := map[string]any{
		"timestamp": time.Now().Unix(),
		"status":    respErr.Status,
	}
	if respErr.MessageTag != "" {
		response["message"] = helpers.Translate(r.Context(), respErr.MessageTag)
	}
	if respErr.Details != nil {
		response["details"] = respErr.Details
	}
	if respErr.LogMessage != "" {
		params := respErr.LogParams
		if respErr.Error != nil {
			params = append(params, "error", respErr.Error)
		}
		slog.Log(r.Context(), respErr.LogLevel, respErr.LogMessage, params...)
	}
	w.WriteHeader(respErr.Status)
	json.NewEncoder(w).Encode(response)
}
