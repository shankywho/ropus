package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_DatabaseAndRepositories(t *testing.T) {
	ctx := context.Background()
	dbMgr, err := NewDatabaseManager(DatabaseConfig{})
	require.NoError(t, err)
	require.NoError(t, dbMgr.HealthCheck(ctx))

	// 1. Transaction Repository
	txRepo := NewTransactionRepository(dbMgr)
	tx := &StoredTransaction{
		ID:           "tx_store_01",
		TenantID:     "org_default",
		CustomerHash: "cust_hash_123",
		MerchantID:   "merch_456",
		Amount:       2500.0,
		Currency:     "USD",
		RiskScore:    94.5,
		Decision:     "BLOCK",
		ModelVersion: "v3.35-xgb-prod",
	}
	require.NoError(t, txRepo.Save(ctx, tx))

	foundTx, err := txRepo.FindByID(ctx, "tx_store_01")
	require.NoError(t, err)
	assert.Equal(t, "BLOCK", foundTx.Decision)
	assert.Equal(t, 2500.0, foundTx.Amount)

	// 2. Case Repository
	caseRepo := NewCaseRepository(dbMgr)
	fraudCase := &StoredCase{
		ID:            "case_store_01",
		TransactionID: "tx_store_01",
		Priority:      "CRITICAL",
		Status:        "INVESTIGATING",
		AssignedAgent: "agent_llm_investigator",
		Evidence:      `{"graph_degree": 14}`,
		Resolution:    "Pending analyst confirmation",
	}
	require.NoError(t, caseRepo.Save(ctx, fraudCase))

	foundCase, err := caseRepo.FindByID(ctx, "case_store_01")
	require.NoError(t, err)
	assert.Equal(t, "INVESTIGATING", foundCase.Status)

	// 3. Model Registry Repository
	modelRepo := NewModelRepository(dbMgr)
	model := &StoredModel{
		ID:               "model_xgb_v4",
		Version:          "v4.0.0",
		Algorithm:        "XGBoost",
		Metrics:          `{"auc": 0.988, "ks": 0.74}`,
		ApprovalStatus:   "APPROVED",
		ArtifactLocation: "s3://ropus-models/xgb-v4.tar.gz",
		DeployedAt:       time.Now().UTC(),
	}
	require.NoError(t, modelRepo.Save(ctx, model))

	foundModel, err := modelRepo.FindByID(ctx, "model_xgb_v4")
	require.NoError(t, err)
	assert.Equal(t, "APPROVED", foundModel.ApprovalStatus)

	// 4. Audit Log (Cryptographic Hash-Chaining)
	auditRepo := NewAuditRepository(dbMgr)
	e1, err := auditRepo.Append(ctx, "analyst_alice", "UPDATE_THRESHOLD", "policy_carding")
	require.NoError(t, err)
	assert.NotEmpty(t, e1.CurrentHash)

	e2, err := auditRepo.Append(ctx, "system_mlops", "DEPLOY_MODEL", "model_xgb_v4")
	require.NoError(t, err)
	assert.Equal(t, e1.CurrentHash, e2.PreviousHash, "Hash chain must link to previous hash")

	entries, err := auditRepo.ListRecent(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(entries))
}
