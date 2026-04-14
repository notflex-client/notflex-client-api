package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"notflex_client_api/api"
	"notflex_client_api/common/database"
)

func main() {
	// 1. Tải cấu hình từ .env
	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: No .env file found or error loading it")
	}

	// 2. Khởi tạo Database (PostgreSQL)
	database.InitDB()

	// 3. Khởi tạo router Chi
	r := chi.NewRouter()

	// 4. Cài đặt middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	// 5. Route cơ bản để test server
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to Notflex Client API!"}`))
	})

	// Mount các router con
	r.Mount("/api/v1/movies", api.MovieRoutes())
	// r.Mount("/api/v1/auth", api.AuthRoutes())

	// 6. Chạy Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server is running on port %s...\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Cannot start server: %v", err)
	}
}
