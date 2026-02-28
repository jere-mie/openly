package handlers

import (
	"net/http"

	"github.com/jere-mie/openly/internal/middleware"
)

// LoginPage renders the login form.
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middleware.IsAuthenticated(r, h.cfg) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	data := h.commonData(r)
	data["Title"] = "Login"
	h.render(w, "login", data)
}

// LoginSubmit processes the login form.
func (h *Handler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	if password != h.cfg.AdminPassword {
		data := h.commonData(r)
		data["Title"] = "Login"
		data["Error"] = "Invalid password. Please try again."
		h.render(w, "login", data)
		return
	}

	middleware.SetAuthCookie(w, h.cfg)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout clears the session and redirects to the landing page.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	middleware.ClearAuthCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
