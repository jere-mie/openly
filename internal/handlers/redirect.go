package handlers

import (
	"net/http"

	"github.com/jere-mie/openly/internal/models"
)

// Redirect looks up a short code and redirects to the original URL, recording a click.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	var url models.URL
	if err := h.db.Get(&url, "SELECT * FROM urls WHERE short_code = ?", shortCode); err != nil {
		http.NotFound(w, r)
		return
	}

	// Record the click
	h.db.Exec(
		"INSERT INTO clicks (url_id, referrer, user_agent, ip_address) VALUES (?, ?, ?, ?)",
		url.ID, r.Referer(), r.UserAgent(), r.RemoteAddr,
	)
	h.db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE id = ?", url.ID)

	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}
