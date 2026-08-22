package saas

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// UserRole defines the permission tier for organization members.
type UserRole string

const (
	RoleOwner   UserRole = "OWNER"
	RoleAdmin   UserRole = "ADMIN"
	RoleAnalyst UserRole = "ANALYST"
	RoleViewer  UserRole = "VIEWER"
)

// PlanTier represents the commercial subscription tier.
type PlanTier string

const (
	PlanStarter    PlanTier = "STARTER"    // 100k requests/month
	PlanGrowth     PlanTier = "GROWTH"     // 5M requests/month
	PlanEnterprise PlanTier = "ENTERPRISE" // Unlimited requests + custom models
)

// Member represents a user within a multi-tenant organization.
type Member struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      UserRole  `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// Organization represents a customer tenant organization.
type Organization struct {
	OrgID         string              `json:"org_id"`
	Name          string              `json:"name"`
	Industry      string              `json:"industry"` // "BANKING", "FINTECH", "ECOMMERCE", "CRYPTO"
	Plan          PlanTier            `json:"plan"`
	Members       map[string]*Member  `json:"members"`
	Configuration TenantConfiguration `json:"configuration"`
	CreatedAt     time.Time           `json:"created_at"`
}

// TenantConfiguration holds per-tenant customizable risk and model settings.
type TenantConfiguration struct {
	ActiveModelVersion   string  `json:"active_model_version"` // e.g. "fraud-xgb-v5"
	BlockRiskThreshold   float64 `json:"block_risk_threshold"` // e.g. 80.0
	ReviewRiskThreshold  float64 `json:"review_risk_threshold"`// e.g. 30.0
	EnableAutonomousAI   bool    `json:"enable_autonomous_ai"`
	WebhookURL           string  `json:"webhook_url,omitempty"`
}

// SaaSManager manages multi-tenant organizations, memberships, and tenant configurations.
type SaaSManager struct {
	mu            sync.RWMutex
	organizations map[string]*Organization
}

// NewSaaSManager initializes the multi-tenant SaaS manager.
func NewSaaSManager() *SaaSManager {
	mgr := &SaaSManager{
		organizations: make(map[string]*Organization),
	}

	// Seed default enterprise demo tenants
	mgr.CreateOrganization("org_acme_bank", "Acme Digital Bank", "BANKING", PlanEnterprise, "admin@acmebank.com", "Acme Admin")
	mgr.CreateOrganization("org_hyper_market", "HyperGlobal Marketplace", "ECOMMERCE", PlanGrowth, "ops@hypermarket.com", "Hyper Ops")
	mgr.CreateOrganization("org_pay_fast", "PayFast Gateway", "FINTECH", PlanEnterprise, "sec@payfast.io", "PayFast Lead")

	return mgr
}

// CreateOrganization provisions a new organization with an initial owner.
func (m *SaaSManager) CreateOrganization(orgID, name, industry string, plan PlanTier, ownerEmail, ownerName string) (*Organization, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if orgID == "" {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		orgID = fmt.Sprintf("org_%s", hex.EncodeToString(b))
	}

	ownerID := fmt.Sprintf("usr_%s", hex.EncodeToString([]byte(ownerEmail))[:8])
	now := time.Now().UTC()

	org := &Organization{
		OrgID:    orgID,
		Name:     name,
		Industry: industry,
		Plan:     plan,
		Members: map[string]*Member{
			ownerID: {
				UserID:   ownerID,
				Email:    ownerEmail,
				Name:     ownerName,
				Role:     RoleOwner,
				JoinedAt: now,
			},
		},
		Configuration: TenantConfiguration{
			ActiveModelVersion:  "fraud-xgb-v5-prod",
			BlockRiskThreshold:  80.0,
			ReviewRiskThreshold: 30.0,
			EnableAutonomousAI:  true,
		},
		CreatedAt: now,
	}

	m.organizations[orgID] = org
	return org, nil
}

// GetOrganization retrieves an organization by ID.
func (m *SaaSManager) GetOrganization(orgID string) (*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	org, exists := m.organizations[orgID]
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", orgID)
	}
	return org, nil
}

// InviteMember adds a new member with a specified role.
func (m *SaaSManager) InviteMember(orgID, email, name string, role UserRole) (*Member, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	org, exists := m.organizations[orgID]
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", orgID)
	}

	sum := sha256.Sum256([]byte(email))
	userID := fmt.Sprintf("usr_%s", hex.EncodeToString(sum[:])[:8])

	member := &Member{
		UserID:   userID,
		Email:    email,
		Name:     name,
		Role:     role,
		JoinedAt: time.Now().UTC(),
	}

	org.Members[userID] = member
	return member, nil
}

// UpdateConfiguration modifies custom threshold and model settings.
func (m *SaaSManager) UpdateConfiguration(orgID string, cfg TenantConfiguration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	org, exists := m.organizations[orgID]
	if !exists {
		return fmt.Errorf("organization '%s' not found", orgID)
	}

	org.Configuration = cfg
	return nil
}
