package billing

import (
	"fmt"
	"sync"
	"time"

	"github.com/shankywho/ropus/backend/internal/saas"
)

// PlanDetails defines pricing and quotas for each tier.
type PlanDetails struct {
	Tier              saas.PlanTier `json:"tier"`
	Name              string        `json:"name"`
	MonthlyBasePrice  float64       `json:"monthly_base_price"`
	IncludedChecks    uint64        `json:"included_checks"`
	OveragePerCheck   float64       `json:"overage_per_check"`
	DedicatedCluster  bool          `json:"dedicated_cluster"`
	CustomMLModels    bool          `json:"custom_ml_models"`
}

// InvoiceItem represents a billed line item.
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    uint64  `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// MonthlyInvoice represents a generated billing statement.
type MonthlyInvoice struct {
	InvoiceID     string        `json:"invoice_id"`
	OrgID         string        `json:"org_id"`
	Period        string        `json:"period"`
	Plan          saas.PlanTier `json:"plan"`
	BaseFee       float64       `json:"base_fee"`
	OverageFee    float64       `json:"overage_fee"`
	TotalAmount   float64       `json:"total_amount"`
	Status        string        `json:"status"` // "PAID", "PENDING", "DRAFT"
	Items         []InvoiceItem `json:"items"`
	GeneratedAt   time.Time     `json:"generated_at"`
}

// BillingEngine manages enterprise subscription plans and invoice generation.
type BillingEngine struct {
	mu       sync.RWMutex
	plans    map[saas.PlanTier]PlanDetails
	invoices map[string][]*MonthlyInvoice // orgID -> Invoices
}

// NewBillingEngine initializes the commercial billing engine.
func NewBillingEngine() *BillingEngine {
	b := &BillingEngine{
		plans: map[saas.PlanTier]PlanDetails{
			saas.PlanStarter: {
				Tier:             saas.PlanStarter,
				Name:             "Starter Tier",
				MonthlyBasePrice: 499.0,
				IncludedChecks:   100000,
				OveragePerCheck:  0.005,
				DedicatedCluster: false,
				CustomMLModels:   false,
			},
			saas.PlanGrowth: {
				Tier:             saas.PlanGrowth,
				Name:             "Growth Tier",
				MonthlyBasePrice: 4999.0,
				IncludedChecks:   5000000,
				OveragePerCheck:  0.002,
				DedicatedCluster: false,
				CustomMLModels:   true,
			},
			saas.PlanEnterprise: {
				Tier:             saas.PlanEnterprise,
				Name:             "Enterprise Dedicated",
				MonthlyBasePrice: 24999.0,
				IncludedChecks:   50000000,
				OveragePerCheck:  0.0008,
				DedicatedCluster: true,
				CustomMLModels:   true,
			},
		},
		invoices: make(map[string][]*MonthlyInvoice),
	}
	return b
}

// GenerateInvoice computes the total cost for a tenant based on usage.
func (b *BillingEngine) GenerateInvoice(orgID string, planTier saas.PlanTier, usage *saas.TenantUsageSnapshot) *MonthlyInvoice {
	b.mu.Lock()
	defer b.mu.Unlock()

	plan, exists := b.plans[planTier]
	if !exists {
		plan = b.plans[saas.PlanStarter]
	}

	var items []InvoiceItem
	baseFee := plan.MonthlyBasePrice
	items = append(items, InvoiceItem{
		Description: fmt.Sprintf("%s Subscription Base Fee", plan.Name),
		Quantity:    1,
		UnitPrice:   baseFee,
		Total:       baseFee,
	})

	overageFee := 0.0
	if usage != nil && usage.RiskChecksTotal > plan.IncludedChecks {
		overageQty := usage.RiskChecksTotal - plan.IncludedChecks
		overageFee = float64(overageQty) * plan.OveragePerCheck
		items = append(items, InvoiceItem{
			Description: fmt.Sprintf("Risk Evaluation Overages (%d reqs beyond quota)", overageQty),
			Quantity:    overageQty,
			UnitPrice:   plan.OveragePerCheck,
			Total:       overageFee,
		})
	}

	now := time.Now().UTC()
	invoice := &MonthlyInvoice{
		InvoiceID:   fmt.Sprintf("inv_%s_%s", orgID, now.Format("200601021504")),
		OrgID:       orgID,
		Period:      now.Format("2006-01"),
		Plan:        planTier,
		BaseFee:     baseFee,
		OverageFee:  overageFee,
		TotalAmount: baseFee + overageFee,
		Status:      "PAID",
		Items:       items,
		GeneratedAt: now,
	}

	b.invoices[orgID] = append(b.invoices[orgID], invoice)
	return invoice
}

// GetInvoices retrieves invoice history for an organization.
func (b *BillingEngine) GetInvoices(orgID string) []*MonthlyInvoice {
	b.mu.RLock()
	defer b.mu.RUnlock()

	invs, exists := b.invoices[orgID]
	if !exists {
		return []*MonthlyInvoice{}
	}
	return invs
}
