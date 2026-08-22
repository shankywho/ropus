package compliance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompliance_EvidencePackageGeneration(t *testing.T) {
	engine := NewComplianceOperationsEngine()

	pkg := engine.GenerateEvidencePackage("SOC2_TYPE_II", "2026-Q3")
	assert.NotEmpty(t, pkg.PackageID)
	assert.NotEmpty(t, pkg.IntegrityHash)
	assert.Equal(t, 5, len(pkg.EvidenceItems))

	for _, item := range pkg.EvidenceItems {
		assert.Equal(t, "COMPLIANT", item.Status)
	}
}
