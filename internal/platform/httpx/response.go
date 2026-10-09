package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if payload == nil {
		w.WriteHeader(status)
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		status = http.StatusInternalServerError
		body, _ = json.Marshal(ErrorResponse{
			Error: ErrorBody{
				Code:    "internal_error",
				Message: "Internal server error",
			},
		})
	}

	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func MethodNotAllowed(w http.ResponseWriter) {
	WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
}
