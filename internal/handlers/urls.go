package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jere-mie/openly/internal/models"
)

// reserved paths that cannot be used as short codes
var reservedPaths = map[string]bool{
	"login": true, "logout": true, "dashboard": true,
	"api": true, "static": true, "admin": true,
}

// CreateURL handles POST /api/urls - creates a new shortened URL.
func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	originalURL := strings.TrimSpace(r.FormValue("url"))
	customCode := strings.TrimSpace(r.FormValue("custom_code"))

	if originalURL == "" {
		jsonError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Add scheme if missing
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		originalURL = "https://" + originalURL
	}

	var shortCode string
	if customCode != "" {
		if !isValidShortCode(customCode) {
			jsonError(w, "Invalid custom code. Use only letters, numbers, hyphens, and underscores (1-64 chars).", http.StatusBadRequest)
			return
		}
		if reservedPaths[strings.ToLower(customCode)] {
			jsonError(w, "This short code is reserved.", http.StatusBadRequest)
			return
		}
		var count int
		h.db.Get(&count, "SELECT COUNT(*) FROM urls WHERE short_code = ?", customCode)
		if count > 0 {
			jsonError(w, "This custom code is already taken.", http.StatusConflict)
			return
		}
		shortCode = customCode
	} else {
		var err error
		shortCode, err = h.generateShortCode()
		if err != nil {
			jsonError(w, "Failed to generate short code", http.StatusInternalServerError)
			return
		}
	}

	result, err := h.db.Exec(
		"INSERT INTO urls (short_code, original_url) VALUES (?, ?)",
		shortCode, originalURL,
	)
	if err != nil {
		jsonError(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"short_code":   shortCode,
		"original_url": originalURL,
	})
}

// DeleteURL handles DELETE /api/urls/{id}.
func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	res, err := h.db.Exec("DELETE FROM urls WHERE id = ?", id)
	if err != nil {
		jsonError(w, "Failed to delete URL", http.StatusInternalServerError)
		return
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		jsonError(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// UpdateURL handles PATCH /api/urls/{id} - updates the short code.
func (h *Handler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body struct {
		ShortCode string `json:"short_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newCode := strings.TrimSpace(body.ShortCode)
	if newCode == "" {
		jsonError(w, "Short code is required", http.StatusBadRequest)
		return
	}
	if !isValidShortCode(newCode) {
		jsonError(w, "Invalid short code. Use only letters, numbers, hyphens, and underscores (1-64 chars).", http.StatusBadRequest)
		return
	}
	if reservedPaths[strings.ToLower(newCode)] {
		jsonError(w, "This short code is reserved.", http.StatusBadRequest)
		return
	}

	// Check if code already taken by another URL
	var count int
	h.db.Get(&count, "SELECT COUNT(*) FROM urls WHERE short_code = ? AND id != ?", newCode, id)
	if count > 0 {
		jsonError(w, "This short code is already taken.", http.StatusConflict)
		return
	}

	res, err := h.db.Exec("UPDATE urls SET short_code = ? WHERE id = ?", newCode, id)
	if err != nil {
		jsonError(w, "Failed to update URL", http.StatusInternalServerError)
		return
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		jsonError(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "updated",
		"short_code": newCode,
	})
}

// GetURLStats handles GET /api/urls/{id}/stats.
func (h *Handler) GetURLStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var url models.URL
	if err := h.db.Get(&url, "SELECT * FROM urls WHERE id = ?", id); err != nil {
		jsonError(w, "URL not found", http.StatusNotFound)
		return
	}

	var clicks []models.Click
	h.db.Select(&clicks, "SELECT * FROM clicks WHERE url_id = ? ORDER BY clicked_at DESC LIMIT 50", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url":    url,
		"clicks": clicks,
	})
}

func (h *Handler) generateShortCode() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	for attempts := 0; attempts < 10; attempts++ {
		b := make([]byte, length)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		for i := range b {
			b[i] = charset[int(b[i])%len(charset)]
		}
		code := string(b)

		var count int
		h.db.Get(&count, "SELECT COUNT(*) FROM urls WHERE short_code = ?", code)
		if count == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique short code after 10 attempts")
}

func isValidShortCode(code string) bool {
	if len(code) < 1 || len(code) > 64 {
		return false
	}
	for _, c := range code {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
