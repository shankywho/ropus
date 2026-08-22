package main

import (
	"context"
	"crypto/subtle"
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
	ShadowEnabled      bool
	ShadowWorkerCount  int
	ShadowQueueCapacity int
	ShadowSampleRate   float64
	CanaryEnabled      bool
	CanaryPercentage   int
	CanaryModelVersion string
	CanaryMaxErrorRate float64
	CanaryMaxFallbackRate float64
	CanaryMaxP95LatencyMs float64
	CanaryMaxP99LatencyMs float64
	CanaryMaxDecisionChangeRate float64
	AdminAPIKey                 string
	DriftMonitorInterval        time.Duration
	DriftMaxWindowSize          int
	DriftMinSamples             int
	DriftPSIWarnThreshold       float64
	DriftPSIHighThreshold       float64
	DriftPSICritThreshold       float64
	RetrainingEnabled                    bool
	RetrainingMinSamples                 int
	RetrainingDriftThreshold             float64
	RetrainingRequiredConsecutiveWindows int
	RetrainingCooldownDuration           time.Duration
	RetrainingMaxErrorRate               float64
	RetrainingMaxLatencyRegressionMs     float64
	MLTrainingEnabled                    bool
	MLTrainingCommand                    string
	MLTrainingTimeout                    time.Duration
	MLTrainingDataset                    string
	MLTrainingOutputDir                  string
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

	shadowEnabled := getEnvOrDefault("SHADOW_SCORING_ENABLED", "true") == "true"
	shadowWorkers := 4
	if val := os.Getenv("SHADOW_WORKER_COUNT"); val != "" {
		var w int
		if _, err := fmt.Sscanf(val, "%d", &w); err == nil && w > 0 {
			shadowWorkers = w
		}
	}
	shadowQueue := 1000
	if val := os.Getenv("SHADOW_QUEUE_CAPACITY"); val != "" {
		var q int
		if _, err := fmt.Sscanf(val, "%d", &q); err == nil && q > 0 {
			shadowQueue = q
		}
	}
	shadowSample := 1.0
	if val := os.Getenv("SHADOW_SAMPLE_RATE"); val != "" {
		var s float64
		if _, err := fmt.Sscanf(val, "%f", &s); err == nil && s >= 0.0 && s <= 1.0 {
			shadowSample = s
		}
	}

	canaryEnabled := getEnvOrDefault("RISK_MODEL_CANARY_ENABLED", "false") == "true"
	canaryPercent := 0
	if val := os.Getenv("RISK_MODEL_CANARY_PERCENT"); val != "" {
		var p int
		if _, err := fmt.Sscanf(val, "%d", &p); err == nil && p >= 0 && p <= 100 {
			canaryPercent = p
		}
	}
	canaryModel := getEnvOrDefault("RISK_MODEL_CANARY_VERSION", "fraud-xgb-25f-candidate-v1")

	maxErrRate := 0.01
	if val := os.Getenv("CANARY_MAX_ERROR_RATE"); val != "" {
		var r float64
		if _, err := fmt.Sscanf(val, "%f", &r); err == nil && r > 0.0 {
			maxErrRate = r
		}
	}
	maxFallbackRate := 0.01
	if val := os.Getenv("CANARY_MAX_FALLBACK_RATE"); val != "" {
		var r float64
		if _, err := fmt.Sscanf(val, "%f", &r); err == nil && r > 0.0 {
			maxFallbackRate = r
		}
	}
	maxP95 := 15.0
	if val := os.Getenv("CANARY_MAX_P95_LATENCY_MS"); val != "" {
		var lat float64
		if _, err := fmt.Sscanf(val, "%f", &lat); err == nil && lat > 0.0 {
			maxP95 = lat
		}
	}
	maxP99 := 25.0
	if val := os.Getenv("CANARY_MAX_P99_LATENCY_MS"); val != "" {
		var lat float64
		if _, err := fmt.Sscanf(val, "%f", &lat); err == nil && lat > 0.0 {
			maxP99 = lat
		}
	}
	maxChangeRate := 0.10
	if val := os.Getenv("CANARY_MAX_DECISION_CHANGE_RATE"); val != "" {
		var cr float64
		if _, err := fmt.Sscanf(val, "%f", &cr); err == nil && cr > 0.0 {
			maxChangeRate = cr
		}
	}

	adminKey := getEnvOrDefault("ADMIN_API_KEY", "adm_risk_super_secret_key_98765")

	driftInterval := 5 * time.Minute
	if val := os.Getenv("DRIFT_MONITOR_INTERVAL"); val != "" {
		if d, err := time.ParseDuration(val); err == nil && d > 0 {
			driftInterval = d
		}
	}
	driftWindow := 10000
	if val := os.Getenv("DRIFT_MAX_WINDOW_SIZE"); val != "" {
		var w int
		if _, err := fmt.Sscanf(val, "%d", &w); err == nil && w > 0 {
			driftWindow = w
		}
	}
	driftMinSamples := 50
	if val := os.Getenv("DRIFT_MIN_SAMPLES_FOR_DRIFT"); val != "" {
		var s int
		if _, err := fmt.Sscanf(val, "%d", &s); err == nil && s > 0 {
			driftMinSamples = s
		}
	}
	driftPSIWarn := 0.10
	if val := os.Getenv("DRIFT_PSI_WARN_THRESHOLD"); val != "" {
		var p float64
		if _, err := fmt.Sscanf(val, "%f", &p); err == nil && p > 0 {
			driftPSIWarn = p
		}
	}
	driftPSIHigh := 0.20
	if val := os.Getenv("DRIFT_PSI_HIGH_THRESHOLD"); val != "" {
		var p float64
		if _, err := fmt.Sscanf(val, "%f", &p); err == nil && p > 0 {
			driftPSIHigh = p
		}
	}
	driftPSICrit := 0.30
	if val := os.Getenv("DRIFT_PSI_CRIT_THRESHOLD"); val != "" {
		var p float64
		if _, err := fmt.Sscanf(val, "%f", &p); err == nil && p > 0 {
			driftPSICrit = p
		}
	}

	retrainingEnabled := getEnvOrDefault("RETRAINING_ENABLED", "true") == "true"
	retrainingMinSamples := 200
	if val := os.Getenv("RETRAINING_MIN_SAMPLES"); val != "" {
		var s int
		if _, err := fmt.Sscanf(val, "%d", &s); err == nil && s > 0 {
			retrainingMinSamples = s
		}
	}
	retrainingDriftThreshold := 0.20
	if val := os.Getenv("RETRAINING_DRIFT_THRESHOLD"); val != "" {
		var d float64
		if _, err := fmt.Sscanf(val, "%f", &d); err == nil && d > 0 {
			retrainingDriftThreshold = d
		}
	}
	retrainingConsecutiveWindows := 2
	if val := os.Getenv("RETRAINING_REQUIRED_CONSECUTIVE_WINDOWS"); val != "" {
		var w int
		if _, err := fmt.Sscanf(val, "%d", &w); err == nil && w > 0 {
			retrainingConsecutiveWindows = w
		}
	}
	retrainingCooldown := 30 * time.Minute
	if val := os.Getenv("RETRAINING_COOLDOWN_MINUTES"); val != "" {
		var m int
		if _, err := fmt.Sscanf(val, "%d", &m); err == nil && m > 0 {
			retrainingCooldown = time.Duration(m) * time.Minute
		}
	}
	retrainingMaxErrorRate := 0.01
	if val := os.Getenv("RETRAINING_MAX_ERROR_RATE"); val != "" {
		var r float64
		if _, err := fmt.Sscanf(val, "%f", &r); err == nil && r > 0 {
			retrainingMaxErrorRate = r
		}
	}
	retrainingMaxLatencyRegression := 5.0
	if val := os.Getenv("RETRAINING_MAX_LATENCY_REGRESSION"); val != "" {
		var l float64
		if _, err := fmt.Sscanf(val, "%f", &l); err == nil && l > 0 {
			retrainingMaxLatencyRegression = l
		}
	}

	mlTrainingEnabled := os.Getenv("ML_TRAINING_ENABLED") == "true"
	mlTrainingCommand := getEnvOrDefault("ML_TRAINING_COMMAND", "python")
	mlTrainingTimeout := 5 * time.Minute
	if val := os.Getenv("ML_TRAINING_TIMEOUT_MINUTES"); val != "" {
		var tm int
		if _, err := fmt.Sscanf(val, "%d", &tm); err == nil && tm > 0 {
			mlTrainingTimeout = time.Duration(tm) * time.Minute
		}
	}
	mlTrainingDataset := getEnvOrDefault("ML_TRAINING_DATASET", "ml-service/data")
	mlTrainingOutputDir := getEnvOrDefault("ML_TRAINING_OUTPUT_DIR", "ml-service/model/candidates")

	return Config{
		Port:                                 port,
		DatabaseURL:                          dbURL,
		RedisURL:                             redisURL,
		MLServiceURL:                         mlServiceURL,
		WebhookSecret:                        webhookSecret,
		KafkaBrokers:                         kafkaBrokers,
		KafkaTopic:                           kafkaTopic,
		ClickHouseAddr:                       chAddr,
		ClickHouseDB:                         getEnvOrDefault("CLICKHOUSE_DB", "default"),
		ClickHouseUser:                       getEnvOrDefault("CLICKHOUSE_USER", "default"),
		ClickHousePassword:                   os.Getenv("CLICKHOUSE_PASSWORD"),
		ShadowEnabled:                        shadowEnabled,
		ShadowWorkerCount:                    shadowWorkers,
		ShadowQueueCapacity:                  shadowQueue,
		ShadowSampleRate:                     shadowSample,
		CanaryEnabled:                        canaryEnabled,
		CanaryPercentage:                     canaryPercent,
		CanaryModelVersion:                   canaryModel,
		CanaryMaxErrorRate:                   maxErrRate,
		CanaryMaxFallbackRate:                maxFallbackRate,
		CanaryMaxP95LatencyMs:                maxP95,
		CanaryMaxP99LatencyMs:                maxP99,
		CanaryMaxDecisionChangeRate:          maxChangeRate,
		AdminAPIKey:                          adminKey,
		DriftMonitorInterval:                 driftInterval,
		DriftMaxWindowSize:                   driftWindow,
		DriftMinSamples:                      driftMinSamples,
		DriftPSIWarnThreshold:                driftPSIWarn,
		DriftPSIHighThreshold:                driftPSIHigh,
		DriftPSICritThreshold:                driftPSICrit,
		RetrainingEnabled:                    retrainingEnabled,
		RetrainingMinSamples:                 retrainingMinSamples,
		RetrainingDriftThreshold:             retrainingDriftThreshold,
		RetrainingRequiredConsecutiveWindows: retrainingConsecutiveWindows,
		RetrainingCooldownDuration:           retrainingCooldown,
		RetrainingMaxErrorRate:               retrainingMaxErrorRate,
		RetrainingMaxLatencyRegressionMs:     retrainingMaxLatencyRegression,
		MLTrainingEnabled:                    mlTrainingEnabled,
		MLTrainingCommand:                    mlTrainingCommand,
		MLTrainingTimeout:                    mlTrainingTimeout,
		MLTrainingDataset:                    mlTrainingDataset,
		MLTrainingOutputDir:                  mlTrainingOutputDir,
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// RequireAdminAuth enforces constant-time administrative API key authentication.
func RequireAdminAuth(adminKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Admin-API-Key")
		if key == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				key = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(adminKey)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "unauthorized",
				"message": "Invalid or missing administrative API key",
			})
			return
		}
		next(w, r)
	}
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
	deviceFeatureStore := features.NewDeviceFeatureStore(redisClient)
	graphStore := features.NewAccountDeviceGraphStore(redisClient)
	paymentTokenStore := features.NewPaymentTokenStore(redisClient)
	deviceVelocityStore := features.NewDeviceVelocityStore(redisClient)
	deviceReputationStore := features.NewDeviceReputationStore(redisClient)

	rulesService := rules.NewService(dbPool)
	rulesHandler := rules.NewHandler(rulesService)

	casesService := cases.NewService(dbPool)
	casesHandler := cases.NewHandler(casesService)

	mlClient := riskengine.NewMLClient(cfg.MLServiceURL)
	orchestrator := riskengine.NewOrchestrator(dbPool, velocityStore, rulesService, mlClient, kms)
	orchestrator.SetDeviceFeatureStore(deviceFeatureStore)
	orchestrator.SetAccountDeviceGraphStore(graphStore)
	orchestrator.SetPaymentTokenStore(paymentTokenStore)
	orchestrator.SetDeviceVelocityStore(deviceVelocityStore)
	orchestrator.SetDeviceReputationStore(deviceReputationStore)

	shadowCfg := riskengine.ShadowScorerConfig{
		Enabled:                  cfg.ShadowEnabled,
		WorkerCount:              cfg.ShadowWorkerCount,
		QueueCapacity:            cfg.ShadowQueueCapacity,
		SampleRate:               cfg.ShadowSampleRate,
		ScoreDivergenceThreshold: 0.05,
		CandidateModelVersion:    "fraud-xgb-25f-candidate-v1",
		CandidateFeatureContract: riskengine.MLFeatureContractV25,
	}
	shadowScorer := riskengine.NewShadowScorer(shadowCfg, mlClient, chClient)
	orchestrator.SetShadowScorer(shadowScorer)

	canaryCfg := riskengine.CanaryRouterConfig{
		Enabled:                  cfg.CanaryEnabled,
		Percentage:               cfg.CanaryPercentage,
		CandidateModelVersion:    cfg.CanaryModelVersion,
		CandidateFeatureContract: riskengine.MLFeatureContractV25,
		MaxErrorRate:             cfg.CanaryMaxErrorRate,
		MaxFallbackRate:          cfg.CanaryMaxFallbackRate,
		MaxP95LatencyMs:          cfg.CanaryMaxP95LatencyMs,
		MaxP99LatencyMs:          cfg.CanaryMaxP99LatencyMs,
		MaxDecisionChangeRate:    cfg.CanaryMaxDecisionChangeRate,
	}
	canaryRouter := riskengine.NewCanaryRouter(canaryCfg, chClient)
	orchestrator.SetCanaryRouter(canaryRouter)

	// Initialize Drift Detector for continuous model monitoring
	driftCfg := riskengine.DriftConfig{
		PSIWarnThreshold:    cfg.DriftPSIWarnThreshold,
		PSIHighThreshold:    cfg.DriftPSIHighThreshold,
		PSICritThreshold:    cfg.DriftPSICritThreshold,
		JSDWarnThreshold:    0.05,
		JSDHighThreshold:    0.10,
		JSDCritThreshold:    0.15,
		KLWarnThreshold:     0.10,
		MinSamplesForDrift:  cfg.DriftMinSamples,
		MaxWindowSize:       cfg.DriftMaxWindowSize,
		CalculationInterval: cfg.DriftMonitorInterval,
		Epsilon:             1e-6,
	}
	driftDetector, err := riskengine.NewDriftDetector(driftCfg, chClient)
	if err != nil {
		log.Fatalf("Failed to initialize DriftDetector: %v", err)
	}
	go driftDetector.Start(context.Background())
	orchestrator.SetDriftDetector(driftDetector)

	// Initialize Retraining Coordinator for automated closed-loop model lifecycle
	retrainingCfg := riskengine.DefaultRetrainingConfig()
	retrainingCfg.Enabled = cfg.RetrainingEnabled
	retrainingCfg.MinSamples = uint32(cfg.RetrainingMinSamples)
	retrainingCfg.DriftThreshold = cfg.RetrainingDriftThreshold
	retrainingCfg.RequiredConsecutiveWindows = cfg.RetrainingRequiredConsecutiveWindows
	retrainingCfg.CooldownDuration = cfg.RetrainingCooldownDuration
	retrainingCfg.MaxErrorRate = cfg.RetrainingMaxErrorRate
	retrainingCfg.MaxLatencyRegressionMs = cfg.RetrainingMaxLatencyRegressionMs

	var runner riskengine.TrainingRunner
	if cfg.MLTrainingEnabled && cfg.MLTrainingCommand != "" {
		runner = riskengine.NewLocalProcessTrainingAdapter(riskengine.LocalProcessConfig{
			Command:     cfg.MLTrainingCommand,
			Args:        []string{"ml-service/train_25f.py"},
			DatasetPath: cfg.MLTrainingDataset,
			OutputDir:   cfg.MLTrainingOutputDir,
			Timeout:     cfg.MLTrainingTimeout,
			MaxLogBytes: 65536,
		})
	} else {
		runner = riskengine.NewFixtureTrainingAdapter()
	}

	retrainingCoordinator := riskengine.NewRetrainingCoordinator(
		retrainingCfg,
		runner,
		shadowScorer,
		canaryRouter,
		chClient,
		func(newModelVersion string) {
			log.Printf("[PROMOTION] Production model promoted: %s", newModelVersion)
		},
	)

	// Initialize Phase 3.19 Observability, SLOs, Incidents, and Health Aggregator
	metricsEngine := riskengine.NewMetricsEngine()
	sloEngine := riskengine.NewSLOEngine(5 * time.Minute)
	alertManager := riskengine.NewAlertManager(&riskengine.LogAlertSink{}, riskengine.NewInMemoryAlertSink(200))
	incidentEngine := riskengine.NewIncidentEngine(alertManager, chClient)
	healthAggregator := riskengine.NewHealthAggregator(
		dbPool,
		redisClient,
		chClient,
		mlClient,
		orchestrator,
		retrainingCoordinator,
		canaryRouter,
		sloEngine,
	)

	orchestrator.SetMetricsEngine(metricsEngine)
	orchestrator.SetSLOEngine(sloEngine)

	retrainingCoordinator.SetMetricsEngine(metricsEngine)
	retrainingCoordinator.SetSLOEngine(sloEngine)
	retrainingCoordinator.SetIncidentEngine(incidentEngine)
	retrainingCoordinator.SetHealthAggregator(healthAggregator)

	// Initialize Phase 3.20 Disaster Recovery Manager, Safety Auditor, Artifact Health & Orphan Cleaner
	statePath := getEnvOrDefault("MODEL_REGISTRY_STATE_PATH", "ml-service/model/registry_state.json")
	var drManager *riskengine.DisasterRecoveryManager
	var artifactScanner *riskengine.ArtifactHealthScanner
	var safetyAuditor *riskengine.SafetyAuditor
	var orphanCleaner *riskengine.OrphanCleaner
	var budgetPolicy *riskengine.ErrorBudgetPolicyEngine

	verifier := riskengine.NewArtifactVerifier()
	artifactStore, _ := riskengine.NewLocalFilesystemArtifactStore("ml-service/model/candidates")
	artifactScanner = riskengine.NewArtifactHealthScanner(artifactStore, verifier)

	if fileStore, err := riskengine.NewFileStateStore(statePath); err == nil {
		retrainingCoordinator.SetStateStore(fileStore)
		drManager = riskengine.NewDisasterRecoveryManager(fileStore, verifier, chClient)
		if drReport, err := drManager.ExecuteRecovery(context.Background(), retrainingCoordinator.GetModelRegistry(), retrainingCoordinator, canaryRouter); err == nil && drReport != nil {
			log.Printf("[DISASTER_RECOVERY] Startup state reconciliation: %s (Repairs: %d, State: %s, Prod: %s)",
				drReport.Status, drReport.RepairsMade, drReport.ReconciledState, drReport.ProductionModelVersion)
		}
	}

	safetyAuditor = riskengine.NewSafetyAuditor(
		retrainingCoordinator.GetModelRegistry(),
		retrainingCoordinator,
		canaryRouter,
		sloEngine,
		verifier,
	)

	orphanCleaner = riskengine.NewOrphanCleaner(
		retrainingCoordinator,
		artifactScanner,
		riskengine.DefaultOrphanCleanerConfig(),
	)

	budgetPolicy = riskengine.NewErrorBudgetPolicyEngine(sloEngine, retrainingCoordinator)

	// Set autonomous remediation callback on critical incidents
	incidentEngine.SetAutoRemediationHandler(func(ctx context.Context, inc riskengine.Incident) {
		if inc.Severity == riskengine.IncidentSeverityCritical {
			log.Printf("[AUTO_REMEDIATION] Enacting safety lock due to CRITICAL incident: %s (%s)", inc.IncidentID, inc.Reason)
			_ = retrainingCoordinator.SetModelFrozen(ctx, true, "AUTO_REMEDIATION", inc.Reason)
		}
	})

	orchestrator.SetRetrainingCoordinator(retrainingCoordinator)
	driftDetector.SetOnMeasurementEvaluated(func(m *riskengine.DriftMeasurement) {
		retrainingCoordinator.OnDriftEvaluated(context.Background(), m)
	})

	riskHandler := riskengine.NewHandler(orchestrator)

	webhookHandler := ingestion.NewWebhookHandler(dbPool)

	// 5. Asynchronous Kafka Consumers & Periodic Incident Evaluator
	kafkaCaseConsumer := cases.NewKafkaConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, "risk-case-manager-group", casesService)
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go kafkaCaseConsumer.Start(consumerCtx)

	kafkaAuditConsumer := audit.NewKafkaAuditConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, "risk-audit-group", chClient)
	go kafkaAuditConsumer.Start(consumerCtx)

	// Start background orphan cleaner
	go orphanCleaner.Start(consumerCtx)

	// Periodic Incident & SLO Evaluation Loop
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				report := healthAggregator.GetHealthReport()
				sloSum := sloEngine.Evaluate(time.Now().UTC())
				budgetPolicy.Evaluate(context.Background())
				var cbState riskengine.CircuitBreakerState = riskengine.CircuitStateHealthy
				if canaryRouter != nil && canaryRouter.GetCircuitBreaker() != nil {
					cbState = canaryRouter.GetCircuitBreaker().GetState()
				}
				var driftSt riskengine.DriftStatus = riskengine.DriftStatusHealthy
				if driftDetector != nil {
					dStatus := driftDetector.GetStatus()
					if stVal, ok := dStatus["status"].(riskengine.DriftStatus); ok {
						driftSt = stVal
					}
				}
				retrainSt := riskengine.StateIdle
				if stVal, ok := retrainingCoordinator.GetStatus()["state"].(riskengine.JobState); ok {
					retrainSt = stVal
				}
				prodModel := "fraud-xgb-25f-v3.0"
				if pm, err := retrainingCoordinator.GetModelRegistry().GetProductionModel(); err == nil {
					prodModel = pm.Version
				}
				incidentEngine.Evaluate(context.Background(), report, sloSum, cbState, driftSt, retrainSt, prodModel)
			case <-consumerCtx.Done():
				return
			}
		}
	}()

	// 6. HTTP Router Setup
	r := chi.NewRouter()

	r.Use(utils.CorrelationIDMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Register Operations, Metrics & Health Handlers
	RegisterOperationsHandlers(
		r,
		retrainingCoordinator,
		healthAggregator,
		sloEngine,
		metricsEngine,
		incidentEngine,
		canaryRouter,
		safetyAuditor,
		artifactScanner,
		drManager,
		cfg.AdminAPIKey,
	)

	startTime := time.Now().UTC()

	// 1. Health check endpoint (Lightweight Liveness probe)
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		uptimeSec := int(time.Since(startTime).Seconds())

		response := map[string]interface{}{
			"status":      "ok",
			"service":     "AI Risk Manager API",
			"version":     "v1.0-prod",
			"uptime_sec":  uptimeSec,
			"model": map[string]interface{}{
				"version":             "fraud-xgb-25f-v3.0",
				"fallback_version":    "fraud-xgb-15f-v1.5",
				"feature_contract":    "fraud-risk-25f-v2.5",
				"calibration_version": "beta-calibrated-v2.5",
				"features_count":      25,
				"status":              "healthy",
			},
			"canary": map[string]interface{}{
				"enabled":                canaryRouter.GetStatus()["enabled"],
				"percentage":             canaryRouter.GetStatus()["target_percentage"],
				"safety_status":          canaryRouter.GetStatus()["safety_gate_status"],
				"circuit_breaker_status": canaryRouter.GetStatus()["circuit_breaker"].(map[string]interface{})["state"],
			},
			"time": time.Now().UTC().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
	r.Get("/health", healthHandler)
	r.Get("/healthz", healthHandler)

	// 2. Readiness check endpoint (Dependency Readiness probe)
	readinessHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		pingCtx, pingCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer pingCancel()

		isReady := true
		dependencies := map[string]string{
			"postgres":   "HEALTHY",
			"redis":      "HEALTHY",
			"clickhouse": "HEALTHY",
			"ml_service": "HEALTHY",
		}

		if err := dbPool.Ping(pingCtx); err != nil {
			dependencies["postgres"] = "UNHEALTHY"
			isReady = false
		}
		if err := redisClient.Ping(pingCtx).Err(); err != nil {
			dependencies["redis"] = "UNHEALTHY"
			isReady = false
		}
		if chClient == nil || chClient.Ping(pingCtx) != nil {
			dependencies["clickhouse"] = "UNHEALTHY"
			isReady = false
		}

		// Fast ML sidecar health ping
		mlHealthReq, _ := http.NewRequestWithContext(pingCtx, "GET", cfg.MLServiceURL+"/health", nil)
		if mlResp, mlErr := http.DefaultClient.Do(mlHealthReq); mlErr != nil || mlResp.StatusCode != http.StatusOK {
			dependencies["ml_service"] = "DEGRADED"
		} else {
			_ = mlResp.Body.Close()
		}

		statusStr := "READY"
		statusCode := http.StatusOK
		if !isReady {
			statusStr = "NOT_READY"
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":                 statusStr,
			"dependencies":           dependencies,
			"production_model":       "fraud-xgb-25f-v3.0",
			"fallback_model":         "fraud-xgb-15f-v1.5",
			"circuit_breaker_state":  canaryRouter.GetStatus()["circuit_breaker"].(map[string]interface{})["state"],
			"evaluated_at":           time.Now().UTC().Format(time.RFC3339),
		})
	}
	r.Get("/readiness", readinessHandler)
	r.Get("/readyz", readinessHandler)

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service": "AI Risk Manager API",
			"version": "v1.0-prod",
			"status":  "operational",
		})
	})

	// Provider Webhooks Ingestion (HMAC protected)
	r.Post("/webhooks/provider", webhookHandler.HandleProviderWebhook)

	// V1 API Routes
	r.Route("/v1", func(r chi.Router) {
		// Real-time risk evaluation orchestrator
		r.Post("/risk-evaluations", riskHandler.EvaluateRisk)

		RegisterDriftHandlers(r, driftDetector)
		RegisterRetrainingHandlers(r, retrainingCoordinator, cfg.AdminAPIKey)

		// Canary Rollout status & safety gate diagnostics
		r.Get("/canary/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(canaryRouter.GetStatus())
		})

		// Authenticated dynamic Canary Control endpoint for hot-reloading percentages
		r.Post("/canary/control", RequireAdminAuth(cfg.AdminAPIKey, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			var ctrlReq struct {
				Enabled    *bool  `json:"enabled"`
				Percentage *int   `json:"percentage"`
				Reason     string `json:"reason"`
			}
			if err := json.NewDecoder(r.Body).Decode(&ctrlReq); err != nil || ctrlReq.Percentage == nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "bad_request",
					"message": "Invalid request body; 'percentage' (0-100) is required",
				})
				return
			}

			// OPERATOR SAFETY: Require explicit reason for audit trail
			reason := strings.TrimSpace(ctrlReq.Reason)
			if reason == "" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "bad_request",
					"message": "A non-empty 'reason' string is required for administrative rollout adjustments",
				})
				return
			}

			enabled := true
			if ctrlReq.Enabled != nil {
				enabled = *ctrlReq.Enabled
			}
			pct := *ctrlReq.Percentage
			if pct < 0 || pct > 100 {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "bad_request",
					"message": "Percentage must be between 0 and 100",
				})
				return
			}

			// OPERATOR SAFETY: If enabling canary while circuit breaker is currently in ROLLED_BACK state,
			// require confirmation in the reason string
			cbStatus := canaryRouter.GetStatus()["circuit_breaker"].(map[string]interface{})
			if enabled && pct > 0 && cbStatus["state"] == "ROLLED_BACK" {
				if !strings.Contains(strings.ToLower(reason), "override") && !strings.Contains(strings.ToLower(reason), "reset") {
					w.WriteHeader(http.StatusConflict)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error":   "circuit_breaker_rolled_back",
						"message": "Circuit breaker is currently in ROLLED_BACK state. Include 'override' or 'reset' in reason to clear cooldown.",
					})
					return
				}
			}

			actor := r.Header.Get("X-Admin-User")
			if actor == "" {
				actor = "admin_api_user"
			}

			if err := canaryRouter.UpdateConfig(enabled, pct, actor, reason); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "bad_request",
					"message": err.Error(),
				})
				return
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "ok",
				"message": fmt.Sprintf("Canary rollout percentage successfully updated to %d%%", pct),
				"canary":  canaryRouter.GetStatus(),
			})
		}))

		// Model Information Status Endpoint
		r.Get("/models/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"service": "risk-engine-model-subsystem",
				"primary_production_model": map[string]interface{}{
					"name":                "fraud_model",
					"version":             "fraud-xgb-25f-v3.0",
					"feature_contract":    "fraud-risk-25f-v2.5",
					"feature_count":       25,
					"calibration_version": "beta-calibrated-v2.5",
					"format":              "ONNX",
					"status":              "ACTIVE",
				},
				"emergency_fallback_model": map[string]interface{}{
					"name":                "fraud_model_15f_v1",
					"version":             "fraud-xgb-15f-v1.5",
					"feature_contract":    "fraud-risk-15f-v1.5",
					"feature_count":       15,
					"calibration_version": "beta-calibrated-v1.5",
					"format":              "ONNX",
					"status":              "ACTIVE_FALLBACK",
				},
				"policy_thresholds": map[string]interface{}{
					"allow":         "< 0.05",
					"manual_review": "0.05 - 0.35",
					"decline":       ">= 0.35",
				},
				"sidecar_url": cfg.MLServiceURL,
				"evaluated_at": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Operational Consolidated System Snapshot Endpoint
		r.Get("/system/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			healthReport := healthAggregator.GetHealthReport()
			sloReport := sloEngine.Evaluate(time.Now().UTC())
			metricsSnap := metricsEngine.GetSnapshot()
			incidents := incidentEngine.ListIncidents()
			opControls := retrainingCoordinator.GetOperationalControls()

			prodVer := "fraud-xgb-25f-v3.0"
			fbVer := "fraud-xgb-15f-v1.5"
			if pm, err := retrainingCoordinator.GetModelRegistry().GetProductionModel(); err == nil {
				prodVer = pm.Version
			}
			if fm, err := retrainingCoordinator.GetModelRegistry().GetFallbackModel(); err == nil {
				fbVer = fm.Version
			}

			status := canaryRouter.GetStatus()
			cbStatus := status["circuit_breaker"].(map[string]interface{})

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":            healthReport.OverallStatus,
				"health":            healthReport,
				"slo":               sloReport,
				"metrics":           metricsSnap,
				"incidents":         incidents,
				"maintenance":       opControls,
				"production_model":  prodVer,
				"fallback_model":    fbVer,
				"canary_percentage": status["target_percentage"],
				"safety_gate":       status["safety_gate_status"],
				"circuit_breaker":   cbStatus["state"],
				"drift":             driftDetector.GetSummary(),
				"retraining":        retrainingCoordinator.GetSummary(),
				"dependencies": map[string]string{
					"ml_service": string(healthReport.Components["ml_runtime"].Status),
					"postgres":   string(healthReport.Components["postgres"].Status),
					"redis":      string(healthReport.Components["redis"].Status),
					"clickhouse": string(healthReport.Components["clickhouse"].Status),
				},
				"evaluated_at": time.Now().UTC().Format(time.RFC3339),
			})
		})

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
		_ = shadowScorer.Close(5 * time.Second)
		driftDetector.Stop()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			log.Fatalf("Could not stop server gracefully: %v", err)
		}
		log.Println("Server stopped gracefully.")
	}
}
