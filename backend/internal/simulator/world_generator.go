package simulator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// WorldConfig configures synthetic environment parameters.
type WorldConfig struct {
	Seed              int64   `json:"seed"`
	CustomerCount     int     `json:"customer_count"`
	MerchantCount     int     `json:"merchant_count"`
	TransactionsCount int     `json:"transactions_count"`
	FraudRatio        float64 `json:"fraud_ratio"` // e.g. 0.02 (2% fraud)
}

// SyntheticCustomer represents a generated customer entity.
type SyntheticCustomer struct {
	CustomerID   string    `json:"customer_id"`
	CustomerHash string    `json:"customer_hash"`
	Name         string    `json:"name"`
	IsSynthetic  bool      `json:"is_synthetic"`
	HomeCountry  string    `json:"home_country"`
	DeviceIDs    []string  `json:"device_ids"`
	CreatedAt    time.Time `json:"created_at"`
}

// SyntheticMerchant represents a registered merchant entity.
type SyntheticMerchant struct {
	MerchantID   string `json:"merchant_id"`
	Name         string `json:"name"`
	CategoryCode string `json:"category_code"`
	RiskTier     string `json:"risk_tier"`
}

// SyntheticTransaction represents a generated banking transaction event.
type SyntheticTransaction struct {
	TransactionID string    `json:"transaction_id"`
	CustomerID    string    `json:"customer_id"`
	MerchantID    string    `json:"merchant_id"`
	DeviceID      string    `json:"device_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Location      string    `json:"location"`
	IsFraud       bool      `json:"is_fraud"`
	FraudType     string    `json:"fraud_type,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// SyntheticWorldEngine generates reproducible, deterministic banking-scale environments.
type SyntheticWorldEngine struct {
	mu           sync.RWMutex
	cfg          WorldConfig
	rng          *rand.Rand
	customers    []*SyntheticCustomer
	merchants    []*SyntheticMerchant
	transactions []*SyntheticTransaction
}

// NewSyntheticWorldEngine initializes the world simulator.
func NewSyntheticWorldEngine(cfg WorldConfig) *SyntheticWorldEngine {
	if cfg.Seed == 0 {
		cfg.Seed = 42
	}
	if cfg.CustomerCount == 0 {
		cfg.CustomerCount = 1000
	}
	if cfg.MerchantCount == 0 {
		cfg.MerchantCount = 50
	}
	if cfg.TransactionsCount == 0 {
		cfg.TransactionsCount = 5000
	}
	if cfg.FraudRatio == 0 {
		cfg.FraudRatio = 0.025
	}

	return &SyntheticWorldEngine{
		cfg:          cfg,
		rng:          rand.New(rand.NewSource(cfg.Seed)),
		customers:    make([]*SyntheticCustomer, 0),
		merchants:    make([]*SyntheticMerchant, 0),
		transactions: make([]*SyntheticTransaction, 0),
	}
}

// HashString produces a deterministic SHA-256 hash.
func HashString(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// GenerateWorld synthesizes customers, merchants, and transactions.
func (e *SyntheticWorldEngine) GenerateWorld() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()

	// 1. Generate Merchants
	categories := []string{"RETAIL", "CRYPTO_EXCHANGE", "TRAVEL", "GAMING", "LUXURY_GOODS"}
	for i := 0; i < e.cfg.MerchantCount; i++ {
		mID := fmt.Sprintf("merch_%04d", i)
		cat := categories[e.rng.Intn(len(categories))]
		riskTier := "LOW"
		if cat == "CRYPTO_EXCHANGE" || cat == "LUXURY_GOODS" {
			riskTier = "HIGH"
		}
		e.merchants = append(e.merchants, &SyntheticMerchant{
			MerchantID:   mID,
			Name:         fmt.Sprintf("Merchant_%s_%04d", cat, i),
			CategoryCode: cat,
			RiskTier:     riskTier,
		})
	}

	// 2. Generate Customers
	countries := []string{"US", "CA", "GB", "DE", "FR", "IN", "SG"}
	for i := 0; i < e.cfg.CustomerCount; i++ {
		cID := fmt.Sprintf("cust_%06d", i)
		isSynth := e.rng.Float64() < 0.05 // 5% synthetic identities
		home := countries[e.rng.Intn(len(countries))]
		devID := fmt.Sprintf("dev_%s_%04d", cID, e.rng.Intn(2))

		e.customers = append(e.customers, &SyntheticCustomer{
			CustomerID:   cID,
			CustomerHash: HashString(cID),
			Name:         fmt.Sprintf("User_%06d", i),
			IsSynthetic:  isSynth,
			HomeCountry:  home,
			DeviceIDs:    []string{devID},
			CreatedAt:    now.Add(-time.Duration(e.rng.Intn(180)) * 24 * time.Hour),
		})
	}

	// 3. Generate Transactions
	fraudTypes := []string{"ACCOUNT_TAKEOVER", "SYNTHETIC_IDENTITY", "CARD_TESTING", "MULE_LAUNDERING"}
	for i := 0; i < e.cfg.TransactionsCount; i++ {
		txID := fmt.Sprintf("tx_%08d", i)
		cust := e.customers[e.rng.Intn(len(e.customers))]
		merch := e.merchants[e.rng.Intn(len(e.merchants))]
		isFraud := e.rng.Float64() < e.cfg.FraudRatio || cust.IsSynthetic

		amount := 10.0 + (e.rng.Float64() * 200.0) // normal spending
		fraudType := ""
		loc := cust.HomeCountry
		dev := cust.DeviceIDs[0]

		if isFraud {
			amount = 1500.0 + (e.rng.Float64() * 12000.0)
			fraudType = fraudTypes[e.rng.Intn(len(fraudTypes))]
			if fraudType == "ACCOUNT_TAKEOVER" {
				loc = "NG" // impossible travel
				dev = "dev_rogue_emulator_88"
			} else if fraudType == "SYNTHETIC_IDENTITY" {
				dev = "dev_shared_mule_cluster"
			}
		}

		e.transactions = append(e.transactions, &SyntheticTransaction{
			TransactionID: txID,
			CustomerID:    cust.CustomerID,
			MerchantID:    merch.MerchantID,
			DeviceID:      dev,
			Amount:        amount,
			Currency:      "USD",
			Location:      loc,
			IsFraud:       isFraud,
			FraudType:     fraudType,
			Timestamp:     now.Add(-time.Duration(e.rng.Intn(7200)) * time.Second),
		})
	}
}

// GetSummary returns generation statistics.
func (e *SyntheticWorldEngine) GetSummary() (int, int, int, int) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	fraudCount := 0
	for _, tx := range e.transactions {
		if tx.IsFraud {
			fraudCount++
		}
	}
	return len(e.customers), len(e.merchants), len(e.transactions), fraudCount
}
