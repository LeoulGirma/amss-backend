package errcode

import "strings"

const (
	CodeAuth        = "auth"
	CodeForbidden   = "forbidden"
	CodeValidation  = "validation"
	CodeConflict    = "conflict"
	CodeNotFound    = "not_found"
	CodeRateLimited = "rate_limited"
	CodeInternal    = "internal"
	CodeUnavailable = "unavailable"
)

func Normalize(code string) string {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return CodeInternal
	}
	switch strings.ToUpper(trimmed) {
	case "AUTH", "AUTH_ERROR", "AUTH_INVALID", "UNAUTHORIZED":
		return CodeAuth
	case "FORBIDDEN":
		return CodeForbidden
	case "VALIDATION", "VALIDATION_ERROR", "BAD_REQUEST":
		return CodeValidation
	case "CONFLICT", "IDEMPOTENCY_CONFLICT":
		return CodeConflict
	case "NOT_FOUND":
		return CodeNotFound
	case "RATE_LIMITED", "RATE_LIMIT":
		return CodeRateLimited
	case "RATE_LIMIT_ERROR":
		return CodeInternal
	case "INTERNAL", "INTERNAL_ERROR":
		return CodeInternal
	case "UNAVAILABLE", "SERVICE_UNAVAILABLE":
		return CodeUnavailable
	default:
		return strings.ToLower(trimmed)
	}
}
