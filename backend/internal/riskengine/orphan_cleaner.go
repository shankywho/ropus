package riskengine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// OrphanCleanerConfig configures TTLs and intervals for autonomous orphan cleanup.
type OrphanCleanerConfig struct {
	ScanInterval           time.Duration `json:"scan_interval"`
	JobMaxDuration         time.Duration `json:"job_max_duration"`
	StaleCandidateTTL      time.Duration `json:"stale_candidate_ttl"`
	QuarantineRetentionTTL time.Duration `json:"quarantine_retention_ttl"`
}

// DefaultOrphanCleanerConfig returns production default timeouts.
func DefaultOrphanCleanerConfig() OrphanCleanerConfig {
	return OrphanCleanerConfig{
		ScanInterval:           1 * time.Minute,
		JobMaxDuration:         30 * time.Minute,
		StaleCandidateTTL:      24 * time.Hour,
		QuarantineRetentionTTL: 7 * 24 * time.Hour,
	}
}

// OrphanCleanupResult summarizes items reconciled and cleaned up during a scan.
type OrphanCleanupResult struct {
	Timestamp          time.Time `json:"timestamp"`
	OrphanedJobsFixed  int       `json:"orphaned_jobs_fixed"`
	StaleCandidatesGC  int       `json:"stale_candidates_gc"`
	QuarantinedFilesGC int       `json:"quarantined_files_gc"`
	Details            []string  `json:"details,omitempty"`
}

// OrphanCleaner autonomously identifies stuck training jobs, orphaned candidate records, and abandoned files.
type OrphanCleaner struct {
	mu          sync.Mutex
	coordinator *RetrainingCoordinator
	scanner     *ArtifactHealthScanner
	config      OrphanCleanerConfig
	stopChan    chan struct{}
}

// NewOrphanCleaner initializes the orphan cleaner worker.
func NewOrphanCleaner(
	coordinator *RetrainingCoordinator,
	scanner *ArtifactHealthScanner,
	config OrphanCleanerConfig,
) *OrphanCleaner {
	if config.ScanInterval <= 0 {
		config.ScanInterval = 1 * time.Minute
	}
	if config.JobMaxDuration <= 0 {
		config.JobMaxDuration = 30 * time.Minute
	}
	if config.StaleCandidateTTL <= 0 {
		config.StaleCandidateTTL = 24 * time.Hour
	}
	if config.QuarantineRetentionTTL <= 0 {
		config.QuarantineRetentionTTL = 7 * 24 * time.Hour
	}

	return &OrphanCleaner{
		coordinator: coordinator,
		scanner:     scanner,
		config:      config,
		stopChan:    make(chan struct{}),
	}
}

// Start launches the periodic background orphan cleanup goroutine.
func (oc *OrphanCleaner) Start(ctx context.Context) {
	ticker := time.NewTicker(oc.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-oc.stopChan:
			return
		case <-ticker.C:
			res, err := oc.CleanOnce(ctx)
			if err != nil {
				log.Printf("[ORPHAN_CLEANER_WARN] Scan failed: %v", err)
			} else if res.OrphanedJobsFixed > 0 || res.StaleCandidatesGC > 0 || res.QuarantinedFilesGC > 0 {
				log.Printf("[ORPHAN_CLEANER] Reconciled %d stuck jobs, %d stale candidates, %d expired quarantine files",
					res.OrphanedJobsFixed, res.StaleCandidatesGC, res.QuarantinedFilesGC)
			}
		}
	}
}

// Stop gracefully signals the background worker to stop.
func (oc *OrphanCleaner) Stop() {
	select {
	case <-oc.stopChan:
	default:
		close(oc.stopChan)
	}
}

// CleanOnce performs a single pass of orphan detection and cleanup.
func (oc *OrphanCleaner) CleanOnce(ctx context.Context) (*OrphanCleanupResult, error) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	result := &OrphanCleanupResult{
		Timestamp: time.Now().UTC(),
		Details:   make([]string, 0),
	}

	now := time.Now().UTC()

	// 1. Check for stuck / orphaned active training jobs
	if oc.coordinator != nil {
		oc.coordinator.mu.Lock()
		activeJob := oc.coordinator.activeJob
		if activeJob != nil {
			isInterruptedState := oc.coordinator.currentState == StateTraining ||
				oc.coordinator.currentState == StateValidating ||
				oc.coordinator.currentState == StateShadowEvaluation

			if isInterruptedState && now.Sub(activeJob.TriggeredAt) > oc.config.JobMaxDuration {
				// Job has exceeded maximum allowed execution duration; fail it safely
				reason := fmt.Sprintf("ORPHAN_CLEANUP_TIMEOUT: Job exceeded max duration %v (triggered %v)",
					oc.config.JobMaxDuration, activeJob.TriggeredAt.Format(time.RFC3339))
				activeJob.State = StateFailed
				activeJob.CompletedAt = now
				activeJob.Error = reason
				oc.coordinator.currentState = StateFailed
				oc.coordinator.jobHistory = append(oc.coordinator.jobHistory, *activeJob)
				oc.coordinator.activeJob = nil
				oc.coordinator.persistStateLocked(ctx)

				result.OrphanedJobsFixed++
				result.Details = append(result.Details, reason)
				log.Printf("[ORPHAN_CLEANER] Reconciled stuck job %s -> FAILED", activeJob.JobID)
			}
		}
		oc.coordinator.mu.Unlock()
	}

	// 2. Clean up expired quarantine files
	if oc.scanner != nil {
		cleaned, err := oc.scanner.CleanupExpiredQuarantine(ctx, oc.config.QuarantineRetentionTTL)
		if err == nil && cleaned > 0 {
			result.QuarantinedFilesGC = cleaned
			result.Details = append(result.Details, fmt.Sprintf("Cleaned up %d expired quarantine files", cleaned))
		}
	}

	return result, nil
}
