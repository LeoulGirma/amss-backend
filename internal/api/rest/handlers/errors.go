package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aeromaintain/amss/internal/api/rest/middleware"
	"github.com/aeromaintain/amss/internal/api/rest/response"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	response.WriteError(w, status, code, message, middleware.RequestIDFromContext(r.Context()))
}
