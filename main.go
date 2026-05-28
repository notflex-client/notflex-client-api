package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"notflex_client_api/api"
	"notflex_client_api/common/database"
	"notflex_client_api/middlewares"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("main: no .env file found")
	}
	if os.Getenv("PORT") == "" {
		panic("main: PORT is not set")
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	slog.Info("starting", "app", "notflex_client_api", "port", os.Getenv("PORT"))

	database.InitDB()
	api.InitStripe()
	r := initRouter()
	setupAPI(r)
	startServer(r)
}

func initRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middlewares.LocaleHeader)
	r.Use(middlewares.ContentTypeHeader)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Cache-Control"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	return r
}

func setupAPI(r *chi.Mux) {
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	type Route struct {
		Method  string
		Path    string
		Handler http.HandlerFunc
	}

	publicRoutes := []Route{
		{"GET", "/health", api.HealthCheck},
		{"POST", "/auth/login", api.Login},
		{"POST", "/auth/login-code", api.CreateLoginCodeRequest},
		{"POST", "/auth/login-code/confirm", api.ConfirmLoginCode},
		{"POST", "/registration/request", api.CreateRegistrationRequest},
		{"POST", "/registration/verify", api.RegistrationVerify},
		{"POST", "/registration/confirm", api.ConfirmRegistrationRequest},
		{"GET", "/proxy/hls", api.ProxyHLS},
		{"GET", "/genres", api.ListGenre},
		{"GET", "/tags", api.ListTag},
		{"GET", "/movies", api.ListMovie},
		{"GET", "/movies/{id}", api.GetMovie},
		{"GET", "/movies/{id}/similar", api.GetSimilarMovies},
		{"GET", "/movies/{movieId}/subtitles", api.ListSubtitleByMovie},
		{"GET", "/movies/{movieId}/audio-tracks", api.ListAudioTrackByMovie},
		{"GET", "/movies/genre/{genreId}", api.GetMoviesByGenre},
		{"GET", "/subscription/plans", api.ListSubscriptionPlans},
		{"POST", "/auth/forgot-password", api.ForgotPassword},
		{"GET", "/auth/forgot-password/token/{token}", api.GetForgotPasswordToken},
		{"POST", "/auth/forgot-password/reset", api.ResetPassword},
		{"POST", "/webhooks/stripe", api.StripeWebhook},
		{"GET", "/banners", api.ListBanner},
	}

	privateRoutes := []Route{
		{"GET", "/auth/me", api.GetProfile},
		{"PUT", "/auth/me", api.UpdateProfile},
		{"POST", "/auth/change-password", api.ChangePassword},
		{"DELETE", "/auth/logout", api.Logout},
		{"GET", "/subscription/me", api.GetMySubscription},
		{"POST", "/subscription/checkout", api.CheckoutSubscription},
		{"POST", "/subscription/stripe-checkout", api.CreateStripeCheckout},
		{"POST", "/subscription/stripe-verify", api.VerifyStripeSession},
		{"GET", "/payments", api.ListPayments},
		{"GET", "/recommendations", api.ListRecommendations},
		{"GET", "/recommendations/me", api.GetMyRecommendations},
		{"POST", "/watch-history", api.CreateWatchHistory},
		{"GET", "/watch-history", api.ListWatchHistory},
		{"POST", "/ratings", api.CreateRating},
		{"GET", "/favorites", api.ListFavorite},
		{"POST", "/favorites/{movieId}", api.CreateFavorite},
		{"DELETE", "/favorites/{movieId}", api.DeleteFavorite},
	}

	adminRoutes := []Route{
		{"POST", "/admin/uploads/video", api.AdminUploadVideo},
		{"POST", "/admin/uploads/image", api.AdminUploadImage},
		{"POST", "/admin/uploads/image", api.AdminUploadImage},
		{"POST", "/admin/movies", api.AdminCreateMovie},
		{"PUT", "/admin/movies/{id}", api.AdminUpdateMovie},
		{"DELETE", "/admin/movies/{id}", api.AdminDeleteMovie},
		{"POST", "/admin/movies/{movieId}/subtitles", api.AdminCreateSubtitle},
		{"DELETE", "/admin/subtitles/{id}", api.AdminDeleteSubtitle},
		{"POST", "/admin/movies/{movieId}/audio-tracks", api.AdminCreateAudioTrack},
		{"DELETE", "/admin/audio-tracks/{id}", api.AdminDeleteAudioTrack},
		{"POST", "/admin/tags", api.AdminCreateTag},
		{"PUT", "/admin/tags/{id}", api.AdminUpdateTag},
		{"DELETE", "/admin/tags/{id}", api.AdminDeleteTag},
		{"POST", "/admin/genres", api.AdminCreateGenre},
		{"PUT", "/admin/genres/{id}", api.AdminUpdateGenre},
		{"DELETE", "/admin/genres/{id}", api.AdminDeleteGenre},
		{"GET", "/admin/subscription-plans", api.AdminListSubscriptionPlan},
		{"POST", "/admin/subscription-plans", api.AdminCreateSubscriptionPlan},
		{"PUT", "/admin/subscription-plans/{id}", api.AdminUpdateSubscriptionPlan},
		{"DELETE", "/admin/subscription-plans/{id}", api.AdminDeleteSubscriptionPlan},
		{"GET", "/admin/movies/{movieId}/episodes", api.AdminListEpisode},
		{"POST", "/admin/movies/{movieId}/episodes", api.AdminCreateEpisode},
		{"PUT", "/admin/episodes/{id}", api.AdminUpdateEpisode},
		{"DELETE", "/admin/episodes/{id}", api.AdminDeleteEpisode},
		{"GET", "/admin/banners", api.AdminListBanner},
		{"POST", "/admin/banners", api.AdminCreateBanner},
		{"PUT", "/admin/banners/{id}", api.AdminUpdateBanner},
		{"DELETE", "/admin/banners/{id}", api.AdminDeleteBanner},
		{"GET", "/admin/payments", api.AdminListPayment},
		{"GET", "/admin/users", api.AdminListUser},
		{"PUT", "/admin/users/{id}", api.AdminUpdateUser},
		{"DELETE", "/admin/users/{id}", api.AdminDeleteUser},
		{"GET", "/admin/stats", api.AdminGetStats},
	}

	r.Group(func(r chi.Router) {
		for _, route := range publicRoutes {
			r.Method(route.Method, route.Path, route.Handler)
		}
	})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.Authentication)
		for _, route := range privateRoutes {
			r.Method(route.Method, route.Path, route.Handler)
		}
	})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.Authentication)
		r.Use(middlewares.RequireRole("admin"))
		for _, route := range adminRoutes {
			r.Method(route.Method, route.Path, route.Handler)
		}
	})
}

func startServer(r *chi.Mux) {
	server := &http.Server{Addr: "0.0.0.0:" + os.Getenv("PORT"), Handler: r}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		database.CloseDB()
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out")
			}
		}()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-serverCtx.Done()
}
