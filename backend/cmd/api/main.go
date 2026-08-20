package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/shankywho/ropus/backend/internal/audit"
	"github.com/shankywho/ropus/backend/internal/cases"
	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/ingestion"
	"github.com/shankywho/ropus/backend/internal/riskengine"
	"github.com/shankywho/ropus/backend/internal/rules"
	"github.com/shankywho/ropus/backend/internal/utils"
)

type Config struct {
	Port               string
	DatabaseURL        string
	RedisURL           string
	MLServiceURL       string
	WebhookSecret      string
	KafkaBrokers       []string
	KafkaTopic         string
	ClickHouseAddr     string
	ClickHouseDB       string
	ClickHouseUser     string
	ClickHousePassword string
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

	mlServiceURL := os.Getenv("ML_SERVICE_URL")
	if mlServiceURL == "" {
		mlServiceURL = "http://localhost:8000"
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if webhookSecret == "" {
		webhookSecret = "whsec_dummy_risk_secret_12345"
	}

	kafkaBrokersRaw := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersRaw == "" {
		kafkaBrokersRaw = getEnvOrDefault("KAFKA_BROKER", "localhost:9092")
	}
	kafkaBrokers := strings.Split(kafkaBrokersRaw, ",")

	kafkaTopic := getEnvOrDefault("KAFKA_TOPIC", "risk.events")

	chAddr := os.Getenv("CLICKHOUSE_ADDR")
	if chAddr == "" {
		chHost := getEnvOrDefault("CLICKHOUSE_HOST", "localhost")
		chPort := getEnvOrDefault("CLICKHOUSE_NATIVE_PORT", "9000")
		chAddr = fmt.Sprintf("%s:%s", chHost, chPort)
	}

	return Config{
		Port:               port,
		DatabaseURL:        dbURL,
		RedisURL:           redisURL,
		MLServiceURL:       mlServiceURL,
		WebhookSecret:      webhookSecret,
		KafkaBrokers:       kafkaBrokers,
		KafkaTopic:         kafkaTopic,
		ClickHouseAddr:     chAddr,
		ClickHouseDB:       getEnvOrDefault("CLICKHOUSE_DB", "default"),
		ClickHouseUser:     getEnvOrDefault("CLICKHOUSE_USER", "default"),
		ClickHousePassword: os.Getenv("CLICKHOUSE_PASSWORD"),
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

	// 3. ClickHouse Client (ClickHouse 24 OLAP)
	chClient, err := audit.NewClickHouseClient(cfg.ClickHouseAddr, cfg.ClickHouseDB, cfg.ClickHouseUser, cfg.ClickHousePassword)
	if err != nil {
		log.Printf("Warning: ClickHouse init error (%v). Continuing...", err)
	} else {
		defer chClient.Close()
		chCtx, chCancel := context.WithTimeout(context.Background(), 2*time.Second)
		if pingErr := chClient.Ping(chCtx); pingErr != nil {
			log.Printf("Warning: ClickHouse ping failed (%v). Continuing...", pingErr)
		} else {
			log.Println("Successfully connected to ClickHouse OLAP server.")
		}
		chCancel()
	}

	// 4. Domain Services, KMS & Handlers
	kms := utils.NewMockKMS()
	velocityStore := features.NewVelocityStore(redisClient)

	rulesService := rules.NewService(dbPool)
	rulesHandler := rules.NewHandler(rulesService)

	casesService := cases.NewService(dbPool)
	casesHandler := cases.NewHandler(casesService)

	mlClient := riskengine.NewMLClient(cfg.MLServiceURL)
	orchestrator := riskengine.NewOrchestrator(dbPool, velocityStore, rulesService, mlClient, kms)
	riskHandler := riskengine.NewHandler(orchestrator)

	webhookHandler := ingestion.NewWebhookHandler(dbPool)

	// 5. Asynchronous Kafka Consumers
	kafkaCaseConsumer := cases.NewKafkaConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, "risk-case-manager-group", casesService)
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go kafkaCaseConsumer.Start(consumerCtx)

	kafkaAuditConsumer := audit.NewKafkaAuditConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, "risk-audit-group", chClient)
	go kafkaAuditConsumer.Start(consumerCtx)

	// 6. HTTP Router Setup
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

		chStatus := "connected"
		if chClient == nil || chClient.Ping(pingCtx) != nil {
			chStatus = "disconnected"
		}

		response := map[string]interface{}{
			"status":        "ok",
			"database":      dbStatus,
			"redis":         redisStatus,
			"clickhouse":    chStatus,
			"kms":           "active",
			"ml_service":    cfg.MLServiceURL,
			"kafka_brokers": cfg.KafkaBrokers,
			"kafka_topic":   cfg.KafkaTopic,
			"time":          time.Now().UTC().Format(time.RFC3339),
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

	// Provider Webhooks Ingestion (HMAC protected)
	r.Post("/webhooks/provider", webhookHandler.HandleProviderWebhook)

	// V1 API Routes
	r.Route("/v1", func(r chi.Router) {
		// Real-time risk evaluation orchestrator
		r.Post("/risk-evaluations", riskHandler.EvaluateRisk)

		// Rules management & Maker-Checker workflow
		r.Route("/rules", func(r chi.Router) {
			r.Post("/", rulesHandler.CreateRule)
			r.Get("/", rulesHandler.ListRules)
			r.Get("/{id}", rulesHandler.GetRule)
			r.Put("/{id}", rulesHandler.UpdateRule)
			r.Put("/{id}/status", rulesHandler.TransitionStatus)
		})

		// Case management & Analyst Manual Review Queue
		r.Route("/cases", func(r chi.Router) {
			r.Get("/", casesHandler.ListCases)
			r.Get("/{id}", casesHandler.GetCase)
			r.Put("/{id}/claim", casesHandler.ClaimCase)
			r.Put("/{id}/resolve", casesHandler.ResolveCase)
		})

		// Webhook alias under /v1
		r.Post("/webhooks/provider", webhookHandler.HandleProviderWebhook)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Graceful Shutdown
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

		consumerCancel()
		_ = kafkaCaseConsumer.Close()
		_ = kafkaAuditConsumer.Close()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			log.Fatalf("Could not stop server gracefully: %v", err)
		}
		log.Println("Server stopped gracefully.")
	}
}
