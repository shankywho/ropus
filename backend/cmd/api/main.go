package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/riskengine"
)

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
}

func loadConfig() Config {
	port := getEnvOrDefault("PORT", "8080")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbHost := getEnvOrDefault("POSTGRES_HOST", "localhost")
		dbPort := getEnvOrDefault("POSTGRES_PORT", "5432")
		dbUser := getEnvOrDefault("POSTGRES_USER", "risk_user")
		dbPass := getEnvOrDefault("POSTGRES_PASSWORD", "risk_password")
		dbName := getEnvOrDefault("POSTGRES_DB", "risk_engine")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
		redisPort := getEnvOrDefault("REDIS_PORT", "6379")
		redisURL = fmt.Sprintf("redis://%s:%s/0", redisHost, redisPort)
	}

	return Config{
		Port:        port,
		DatabaseURL: dbURL,
		RedisURL:    redisURL,
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	cfg := loadConfig()

	log.Printf("Starting AI Risk Manager API on port %s...", cfg.Port)

	// 1. Database Connection Pool (Postgres 16)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to parse DATABASE_URL: %v", err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Printf("Warning: Postgres ping failed (%v). Continuing to start server...", err)
	} else {
		log.Println("Successfully connected to PostgreSQL database.")
	}

	// 2. Redis Client (Redis 7)
	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Unable to parse REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpt)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis ping failed (%v). Continuing to start server...", err)
	} else {
		log.Println("Successfully connected to Redis feature store.")
	}

	// 3. Domain Services & Handlers
	velocityStore := features.NewVelocityStore(redisClient)
	riskHandler := riskengine.NewHandler(velocityStore)

	// 4. HTTP Router Setup
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		dbStatus := "connected"
		pingCtx, pingCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer pingCancel()

		if err := dbPool.Ping(pingCtx); err != nil {
			dbStatus = "disconnected"
		}

		redisStatus := "connected"
		if err := redisClient.Ping(pingCtx).Err(); err != nil {
			redisStatus = "disconnected"
		}

		response := map[string]interface{}{
			"status":   "ok",
			"database": dbStatus,
			"redis":    redisStatus,
			"time":     time.Now().UTC().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	})

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service": "AI Risk Manager API",
			"version": "v1.0-mvp",
		})
	})

	// V1 Risk Evaluation API route
	r.Route("/v1", func(r chi.Router) {
		r.Post("/risk-evaluations", riskHandler.EvaluateRisk)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Graceful Shutdown
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("AI Risk Manager API listening on :%s", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)
	case sig := <-shutdown:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			log.Fatalf("Could not stop server gracefully: %v", err)
		}
		log.Println("Server stopped gracefully.")
	}
}
