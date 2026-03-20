package main

import (
	"log"
	"net/http"
	"os"

	"github.com/slikk66/newspin/internal/auth"
	"github.com/slikk66/newspin/internal/db"
	"github.com/slikk66/newspin/internal/handlers"
	"github.com/slikk66/newspin/internal/news"
)

func main() {
	// 1. Init dependencies
	dbClient, err := db.NewClient(os.Getenv("USERS_TABLE"), os.Getenv("PINS_TABLE"))
	if err != nil {
		log.Fatal("failed to init db:", err)
	}

	newsClient := news.NewClient(os.Getenv("NEWS_API_KEY"), "https://newsapi.org/v2")
	h := handlers.NewHandler(dbClient, newsClient)

	// 2. Create router
	mux := http.NewServeMux()

	// 3. Public routes (no auth)
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/login", h.Login)
	mux.HandleFunc("GET /api/news", h.SearchNews)
	mux.HandleFunc("GET /api/news/featured", h.GetFeatured)

	// 4. Protected routes (auth middleware)
	mux.Handle("GET /api/pins", auth.AuthMiddleware(http.HandlerFunc(h.GetPins)))
	mux.Handle("POST /api/pins", auth.AuthMiddleware(http.HandlerFunc(h.CreatePin)))
	mux.Handle("DELETE /api/pins", auth.AuthMiddleware(http.HandlerFunc(h.DeletePin)))

	// 5. Static files (Astro build)
	mux.Handle("/", http.FileServer(http.Dir("./web/dist")))

	// 6. Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
