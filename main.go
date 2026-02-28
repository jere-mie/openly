package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jere-mie/openly/internal/config"
	"github.com/jere-mie/openly/internal/database"
	"github.com/jere-mie/openly/internal/handlers"
	"github.com/jere-mie/openly/internal/middleware"
)

//go:embed templates
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed migrations
var migrationFS embed.FS

//go:embed version.txt
var version string

func main() {
	cfg := config.Load()

	// CLI commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			db, err := database.Connect(cfg.DatabasePath)
			if err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}
			defer db.Close()
			if err := database.RunMigrations(db, migrationFS); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
			log.Println("All migrations applied successfully.")
			return
		case "version":
			fmt.Print(strings.TrimSpace(version))
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			fmt.Fprintln(os.Stderr, "Usage: openly [migrate|version]")
			os.Exit(1)
		}
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Auto-run pending migrations on startup
	if err := database.RunMigrations(db, migrationFS); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}

	// Initialize handlers
	h := handlers.New(db, cfg, templateFS)

	// Set up routes
	mux := http.NewServeMux()

	// Static files (embedded)
	mux.Handle("GET /static/", http.FileServerFS(staticFS))

	// Auth routes
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.LoginSubmit)
	mux.HandleFunc("POST /logout", h.Logout)

	// Dashboard (protected)
	mux.Handle("GET /dashboard", middleware.RequireAuth(cfg, http.HandlerFunc(h.Dashboard)))

	// API routes (protected)
	mux.Handle("POST /api/urls", middleware.RequireAuthAPI(cfg, http.HandlerFunc(h.CreateURL)))
	mux.Handle("DELETE /api/urls/{id}", middleware.RequireAuthAPI(cfg, http.HandlerFunc(h.DeleteURL)))
	mux.Handle("PATCH /api/urls/{id}", middleware.RequireAuthAPI(cfg, http.HandlerFunc(h.UpdateURL)))
	mux.Handle("GET /api/urls/{id}/stats", middleware.RequireAuthAPI(cfg, http.HandlerFunc(h.GetURLStats)))

	// Public landing page (exact root match)
	mux.HandleFunc("GET /{$}", h.Index)

	// Short URL redirect (single path segment catch-all)
	mux.HandleFunc("GET /{shortCode}", h.Redirect)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("openly %s listening on http://%s", strings.TrimSpace(version), addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
