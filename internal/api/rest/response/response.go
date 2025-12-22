package response

import (
	"encoding/json"
	"net/http"

	"github.com/aeromaintain/amss/internal/api/rest/errcode"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, code, message, requestID string) {
	payload := ErrorResponse{
		Error:     message,
		Code:      errcode.Normalize(code),
		RequestID: requestID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
