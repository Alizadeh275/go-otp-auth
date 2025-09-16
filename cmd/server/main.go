// @title OTP Auth API
// @version 1.0
// @description This is the OTP authentication service API
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @type apiKey
// @in header
// @name Authorization

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/example/go-otp-auth/docs" // import generated swagger docs
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/example/go-otp-auth/internal/api"
	"github.com/example/go-otp-auth/internal/auth"
	"github.com/example/go-otp-auth/internal/config"
	"github.com/example/go-otp-auth/internal/storage"
)

func main() {
	// load config from env
	cfg, err := config.LoadFromEnv()
	if err != nil {
		panic(err)
	}

	log.Info().Msg("starting server")

	// init postgres
	pg, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect postgres")
	}
	defer pg.Close()

	// init redis
	rd, err := storage.NewRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect redis")
	}
	defer rd.Close()

	// init jwt
	auth.InitJWT(cfg.JWTSecret)

	// build router
	r := chi.NewRouter()

	h := api.NewHandler(pg, rd, cfg)

	// OTP endpoints
	r.Post("/otp/request", h.RequestOTP)
	r.Post("/otp/verify", h.VerifyOTP)

	// User endpoints (protected)
	r.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware)
		r.Get("/users", h.ListUsers)
		r.Get("/users/{id}", h.GetUser)
	})

	// Swagger UI routes
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info().Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown failed")
	}
}
