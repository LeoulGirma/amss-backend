package middleware

import (
	"net/http"

	"github.com/aeromaintain/amss/internal/api/rest/response"
)

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	response.WriteError(w, status, code, message, RequestIDFromContext(r.Context()))
}
