package graphsage

import (
	"fmt"
	"math"
	"time"
)

// EmployeeBaselineProfile captures personal historical habits.
type EmployeeBaselineProfile struct {
	HasHistoricalBaseline bool      `json:"has_historical_baseline"`
	PriorEventCount       int       `json:"prior_event_count"`
	MeanDailyVolume       float64   `json:"mean_daily_volume"`
	FrequentHours         []int     `json:"frequent_hours"`
	TrustedDevices        []string  `json:"trusted_devices"`
	TrustedIPs            []string  `json:"trusted_ips"`
	VolumeZScore          float64   `json:"volume_z_score"`
	IsUnusualHour         bool      `json:"is_unusual_hour"`
	IsNewDevice           bool      `json:"is_new_device"`
	PersonalAnomalyDelta  float64   `json:"personal_anomaly_delta"`
	Notes                 string    `json:"notes"`
}

// EmployeeBaselineEngine computes point-in-time personal baseline profiles in Go.
type EmployeeBaselineEngine struct{}

// ComputeBaselineAndDeviation evaluates current activity against historical profile.
func (e *EmployeeBaselineEngine) ComputeBaselineAndDeviation(
	employeeID string,
	currentEvents []*HeteroEdge,
	historicalEdges []*HeteroEdge,
	asOf time.Time,
) *EmployeeBaselineProfile {
	var prior []*HeteroEdge
	for _, edge := range historicalEdges {
		if (edge.SourceID == employeeID || edge.TargetID == employeeID) && edge.Timestamp.Before(asOf) {
			prior = append(prior, edge)
		}
	}

	if len(prior) == 0 {
		return &EmployeeBaselineProfile{
			HasHistoricalBaseline: false,
			FrequentHours:         []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
			PersonalAnomalyDelta:  0.10,
			Notes:                 "Cold-start employee; default standard baseline applied.",
		}
	}

	hourCounts := make(map[int]int)
	daysMap := make(map[string]int)
	trustedDevs := make(map[string]bool)
	trustedIPs := make(map[string]bool)

	for _, edge := range prior {
		h := edge.Timestamp.Hour()
		hourCounts[h]++
		dKey := edge.Timestamp.Format("2006-01-02")
		daysMap[dKey]++

		if edge.Type == EdgeEmployeeUsesDevice || edge.Type == EdgeUsesDevice {
			target := edge.TargetID
			if target == employeeID {
				target = edge.SourceID
			}
			trustedDevs[target] = true
		}
		if edge.Type == EdgeEmployeeUsesIP || edge.Type == EdgeUsesIP {
			target := edge.TargetID
			if target == employeeID {
				target = edge.SourceID
			}
			trustedIPs[target] = true
		}
	}

	var frequentHours []int
	for h, cnt := range hourCounts {
		if float64(cnt)/float64(len(prior)) >= 0.05 {
			frequentHours = append(frequentHours, h)
		}
	}

	meanVol := float64(len(prior)) / math.Max(float64(len(daysMap)), 1.0)
	currVol := float64(len(currentEvents))
	volZ := (currVol - meanVol) / math.Max(meanVol*0.5, 1.0)

	freqHourMap := make(map[int]bool)
	for _, h := range frequentHours {
		freqHourMap[h] = true
	}

	unusualHourCount := 0
	isNewDev := false
	for _, edge := range currentEvents {
		if !freqHourMap[edge.Timestamp.Hour()] {
			unusualHourCount++
		}
		if edge.Type == EdgeEmployeeUsesDevice {
			target := edge.TargetID
			if target == employeeID {
				target = edge.SourceID
			}
			if len(trustedDevs) > 0 && !trustedDevs[target] {
				isNewDev = true
			}
		}
	}

	isUnusual := false
	if len(currentEvents) > 0 && float64(unusualHourCount)/float64(len(currentEvents)) >= 0.50 {
		isUnusual = true
	}

	anomalyDelta := 0.0
	if isUnusual {
		anomalyDelta += 0.35
	}
	if volZ > 2.0 {
		anomalyDelta += 0.30
	}
	if isNewDev {
		anomalyDelta += 0.25
	}
	anomalyDelta = math.Min(1.0, anomalyDelta)

	var devList []string
	for d := range trustedDevs {
		devList = append(devList, d)
	}
	var ipList []string
	for ip := range trustedIPs {
		ipList = append(ipList, ip)
	}

	return &EmployeeBaselineProfile{
		HasHistoricalBaseline: true,
		PriorEventCount:       len(prior),
		MeanDailyVolume:       math.Round(meanVol*100) / 100,
		FrequentHours:         frequentHours,
		TrustedDevices:        devList,
		TrustedIPs:            ipList,
		VolumeZScore:          math.Round(volZ*100) / 100,
		IsUnusualHour:         isUnusual,
		IsNewDevice:           isNewDev,
		PersonalAnomalyDelta:  math.Round(anomalyDelta*100) / 100,
		Notes:                 fmt.Sprintf("Profile computed from %d historical events across %d days.", len(prior), len(daysMap)),
	}
}
