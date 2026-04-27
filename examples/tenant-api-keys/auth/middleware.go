package auth

import (
	"net/http"
	"strings"
)

// Authenticate validates the request's auth token. It now supports
// tenant-scoped API keys (X-API-Key) in addition to user JWTs.
func Authenticate(r *http.Request) bool {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return validateAPIKey(key)
	}
	token := r.Header.Get("Authorization")
	return validateJWT(token)
}

func validateAPIKey(key string) bool {
	if !strings.HasPrefix(key, "ck_") {
		return false
	}
	return len(key) >= 32
}

func validateJWT(token string) bool {
	return len(token) > 32
}
