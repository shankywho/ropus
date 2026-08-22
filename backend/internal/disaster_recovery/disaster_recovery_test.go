package disaster_recovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisasterRecovery_SnapshotsAndSLA(t *testing.T) {
	dr := NewDisasterRecoveryEngine()

	// 1. Create Snapshot
	bkp := dr.CreateSnapshot("FULL", 524288000)
	assert.NotEmpty(t, bkp.BackupID)
	assert.NotEmpty(t, bkp.ChecksumSHA256)
	assert.True(t, bkp.IsReplicated)

	// 2. Verify SLA
	isCompliant, rpo, rto := dr.VerifyRecoverySLA()
	assert.True(t, isCompliant)
	assert.Less(t, rpo.Seconds(), 300.0)
	assert.Less(t, rto.Minutes(), 30.0)

	// 3. List Backups
	backups := dr.ListBackups()
	assert.Equal(t, 1, len(backups))
}
