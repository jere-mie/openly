package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/jere-mie/openly/internal/config"
	"github.com/jere-mie/openly/internal/middleware"
	"github.com/jmoiron/sqlx"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	db        *sqlx.DB
	cfg       *config.Config
	templates map[string]*template.Template
}

// New creates a Handler with parsed templates.
func New(db *sqlx.DB, cfg *config.Config, templateFS embed.FS) *Handler {
	funcMap := template.FuncMap{
		"formatDate": func(s string) string {
			if s == "" {
				return ""
			}
			for _, layout := range []string{
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05Z",
				time.RFC3339,
			} {
				if t, err := time.Parse(layout, s); err == nil {
					return t.Format("Jan 2, 2006 · 3:04 PM")
				}
			}
			return s
		},
		"truncate": func(s string, n int) string {
			runes := []rune(s)
			if len(runes) <= n {
				return s
			}
			return string(runes[:n]) + "…"
		},
		"formatNumber": func(n int64) string {
			s := fmt.Sprintf("%d", n)
			if len(s) <= 3 {
				return s
			}
			var b strings.Builder
			for i, c := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					b.WriteByte(',')
				}
				b.WriteRune(c)
			}
			return b.String()
		},
	}

	parseTemplate := func(files ...string) *template.Template {
		return template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS, files...),
		)
	}

	return &Handler{
		db:  db,
		cfg: cfg,
		templates: map[string]*template.Template{
			"index":     parseTemplate("templates/base.html", "templates/index.html"),
			"login":     parseTemplate("templates/base.html", "templates/login.html"),
			"dashboard": parseTemplate("templates/base.html", "templates/dashboard.html"),
		},
	}
}

// render executes a named template with the given data.
func (h *Handler) render(w http.ResponseWriter, name string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates[name].ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// commonData returns a map with data shared across all pages.
func (h *Handler) commonData(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"Authenticated": middleware.IsAuthenticated(r, h.cfg),
	}
}
