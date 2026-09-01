package runtime

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
)

// adminCorsPreflightMiddleware immediately responds to CORS OPTIONS preflight requests (Task 3.5).
// CRITICAL: This MUST run before Auth enforcement, because browsers do not send Auth headers on OPTIONS.
func (s *HTTPServer) adminCorsPreflightMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict CORS Configuration (Dynamic Origin Echo)
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback (Browsers reject * when Authorization is present, but ok for non-credentialed)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Version")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204
			return
		}
		next.ServeHTTP(w, r)
	})
}

// adminSecurityMiddleware enforces Localhost binding, Rate Limiting, and Authorization
func (s *HTTPServer) adminSecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Localhost Binding (Strict Check via X-Forwarded-For)
		host := r.RemoteAddr
		// Check standard RemoteAddr
		if !strings.HasPrefix(host, "127.0.0.1:") && !strings.HasPrefix(host, "[::1]:") {
			s.sendAdminError(w, "FORBIDDEN", "Admin API is restricted to localhost", http.StatusForbidden)
			return
		}

		// Prevent Proxy Bypass
		forwardedFor := r.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			ips := strings.Split(forwardedFor, ",")
			for _, ip := range ips {
				cleanIP := strings.TrimSpace(ip)
				if cleanIP != "127.0.0.1" && cleanIP != "::1" {
					s.sendAdminError(w, "FORBIDDEN", "Admin API proxy access is restricted to localhost", http.StatusForbidden)
					return
				}
			}
		}

		realIP := r.Header.Get("X-Real-IP")
		if realIP != "" && realIP != "127.0.0.1" && realIP != "::1" {
			s.sendAdminError(w, "FORBIDDEN", "Admin API proxy access is restricted to localhost", http.StatusForbidden)
			return
		}

		// 2. Rate Limiting (EDL-054)
		if !s.rateLimiter.Allow() {
			s.sendAdminError(w, "RATE_LIMIT_EXCEEDED", "Admin API rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// 3. Authorization Bearer Token (Auth Bypass check for Server-Sent Events Fallback)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// EventSource JS API cannot send custom headers. Fallback to query param.
			if r.URL.Path == "/__mockctl/events" && r.Method == http.MethodGet {
				authHeader = "Bearer " + r.URL.Query().Get("token")
			}
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			s.sendAdminError(w, "UNAUTHORIZED", "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		token = strings.Trim(token, `"'`)
		expectedToken, err := s.systemStore.GetAuthToken(r.Context())
		if err != nil || expectedToken == "" {
			s.sendAdminError(w, "UNAUTHORIZED", "Admin token not configured", http.StatusUnauthorized)
			return
		}

		// Security Rule (PKS-028): Constant-Time Comparison
		if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			s.sendAdminError(w, "UNAUTHORIZED", fmt.Sprintf("Invalid admin token. Received: '%s', Expected: '%s'", token, expectedToken), http.StatusUnauthorized)
			return
		}

		// 4. Content-Type Validation for Mutations
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") && !strings.HasPrefix(contentType, "multipart/form-data") {
				s.sendAdminError(w, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json or multipart/form-data", http.StatusUnsupportedMediaType)
				return
			}
		}

		// 5. Accept-Version Check
		acceptVersion := r.Header.Get("Accept-Version")
		if acceptVersion == "" && r.URL.Path == "/__mockctl/events" && r.Method == http.MethodGet {
			// EventSource fallback
			acceptVersion = r.URL.Query().Get("version")
		}

		if acceptVersion != "v1" {
			s.sendAdminError(w, "BAD_REQUEST", "Accept-Version header must be v1", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
