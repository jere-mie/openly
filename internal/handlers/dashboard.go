package handlers

import (
	"net/http"

	"github.com/jere-mie/openly/internal/models"
)

// Dashboard renders the main dashboard view.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	var urls []models.URL
	if err := h.db.Select(&urls, "SELECT * FROM urls ORDER BY created_at DESC"); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var totalClicks int64
	h.db.Get(&totalClicks, "SELECT COALESCE(SUM(clicks), 0) FROM urls")

	data := h.commonData(r)
	data["Title"] = "Dashboard"
	data["URLs"] = urls
	data["TotalURLs"] = len(urls)
	data["TotalClicks"] = totalClicks

	h.render(w, "dashboard", data)
}

// Index renders the public landing page.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	data := h.commonData(r)
	data["Title"] = "Home"
	h.render(w, "index", data)
}
