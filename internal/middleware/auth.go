package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jere-mie/openly/internal/config"
)

const cookieName = "openly_session"

// RequireAuth wraps a handler and redirects unauthenticated requests to /login.
func RequireAuth(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r, cfg) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuthAPI wraps a handler and returns 401 for unauthenticated API requests.
func RequireAuthAPI(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r, cfg) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// IsAuthenticated checks whether the request contains a valid session cookie.
func IsAuthenticated(r *http.Request, cfg *config.Config) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return false
	}
	return validateToken(cookie.Value, cfg.AdminPassword)
}

// SetAuthCookie sets a signed session cookie on the response.
func SetAuthCookie(w http.ResponseWriter, cfg *config.Config) {
	token := generateToken(cfg.AdminPassword)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7, // 7 days
	})
}

// ClearAuthCookie removes the session cookie.
func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func generateToken(password string) string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(password))
	mac.Write([]byte(timestamp))
	signature := hex.EncodeToString(mac.Sum(nil))
	return timestamp + ":" + signature
}

func validateToken(token, password string) bool {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return false
	}

	timestamp := parts[0]
	signature := parts[1]

	mac := hmac.New(sha256.New, []byte(password))
	mac.Write([]byte(timestamp))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}
