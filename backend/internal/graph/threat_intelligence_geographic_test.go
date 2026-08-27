package graph

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// TestHaversineKnownGeographicPairs verifies great-circle spherical distance calculations
// against internationally accepted geodesic ground-truth coordinates.
func TestHaversineKnownGeographicPairs(t *testing.T) {
	testCases := []struct {
		name        string
		p1          GeoPoint
		p2          GeoPoint
		expectedKm  float64
		toleranceKm float64
	}{
		{
			name:        "London to New York",
			p1:          GeoPoint{Lat: 51.5074, Lon: -0.1278},
			p2:          GeoPoint{Lat: 40.7128, Lon: -74.0060},
			expectedKm:  5570.0,
			toleranceKm: 60.0,
		},
		{
			name:        "Tokyo to Sydney",
			p1:          GeoPoint{Lat: 35.6762, Lon: 139.6503},
			p2:          GeoPoint{Lat: -33.8688, Lon: 151.2093},
			expectedKm:  7820.0,
			toleranceKm: 70.0,
		},
		{
			name:        "Singapore to London",
			p1:          GeoPoint{Lat: 1.3521, Lon: 103.8198},
			p2:          GeoPoint{Lat: 51.5074, Lon: -0.1278},
			expectedKm:  10860.0,
			toleranceKm: 80.0,
		},
		{
			name:        "Bengaluru to San Francisco",
			p1:          GeoPoint{Lat: 12.9716, Lon: 77.5946},
			p2:          GeoPoint{Lat: 37.7749, Lon: -122.4194},
			expectedKm:  14000.0,
			toleranceKm: 100.0,
		},
		{
			name:        "Identical Coords (Zero Distance)",
			p1:          GeoPoint{Lat: 12.9716, Lon: 77.5946},
			p2:          GeoPoint{Lat: 12.9716, Lon: 77.5946},
			expectedKm:  0.0,
			toleranceKm: 0.001,
		},
		{
			name:        "North Pole to South Pole (Antipodal)",
			p1:          GeoPoint{Lat: 90.0, Lon: 0.0},
			p2:          GeoPoint{Lat: -90.0, Lon: 0.0},
			expectedKm:  20015.0, // pi * 6371.0 = 20015.08 km
			toleranceKm: 50.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualKm := CalculateHaversineDistanceKm(tc.p1, tc.p2)
			diff := math.Abs(actualKm - tc.expectedKm)
			if diff > tc.toleranceKm {
				t.Errorf("%s: expected %.2f km (+/- %.2f km), got %.2f km (delta: %.2f km)",
					tc.name, tc.expectedKm, tc.toleranceKm, actualKm, diff)
			}
		})
	}
}

// TestImpossibleTravelVelocityBoundaries tests the physical kinematics velocity threshold (>900 km/h).
func TestImpossibleTravelVelocityBoundaries(t *testing.T) {
	engine := NewThreatIntelligenceEngine()

	// Location A: Bengaluru
	locA := GeoPoint{Lat: 12.9716, Lon: 77.5946}

	// 1. Physically Possible Travel (100 km in 2 hours = 50 km/h)
	// Mysore (12.2958, 76.6394) is ~128 km from Bengaluru
	locMysore := GeoPoint{Lat: 12.2958, Lon: 76.6394}
	dist1, speed1 := CalculateImpliedSpeedKmh(locA, locMysore, 2*time.Hour)
	if speed1 > 900.0 {
		t.Errorf("Expected normal speed <= 900 km/h, got %.2f km/h", speed1)
	}
	if dist1 < 100.0 || dist1 > 160.0 {
		t.Errorf("Expected distance ~130km, got %.2f km", dist1)
	}

	// 2. Impossible Travel (Bengaluru to London in 15 minutes = 32,000 km/h)
	locLondon := GeoPoint{Lat: 51.5074, Lon: -0.1278}
	_, speed2 := CalculateImpliedSpeedKmh(locA, locLondon, 15*time.Minute)
	if speed2 < 900.0 {
		t.Errorf("Expected impossible speed > 900 km/h, got %.2f km/h", speed2)
	}

	// 3. Zero elapsed duration safety check (returns 0.0 safely without panic)
	dist3, speed3 := CalculateImpliedSpeedKmh(locA, locLondon, 0*time.Second)
	if dist3 <= 0 {
		t.Errorf("Expected non-zero distance")
	}
	if speed3 != 0.0 {
		t.Errorf("Expected safe 0.0 speed for zero elapsed time, got %.2f", speed3)
	}

	// 4. Negative elapsed duration safety check
	_, speed4 := CalculateImpliedSpeedKmh(locA, locLondon, -10*time.Minute)
	if speed4 != 0.0 {
		t.Errorf("Expected 0.0 speed for negative elapsed duration, got %.2f", speed4)
	}

	_ = engine
}

// TestCIDRAndSubnetMatchingEdgeCases tests IP resolution across exact matches, CIDR blocks, private IPs, and malformed strings.
func TestCIDRAndSubnetMatchingEdgeCases(t *testing.T) {
	engine := NewThreatIntelligenceEngine()

	// 1. Exact Malicious IP Match
	scoreExact, matchesExact := engine.CheckThreat("198.51.100.44", "clean_dev", "clean@example.com")
	if scoreExact < 0.80 {
		t.Errorf("Expected high threat score for exact malicious IP, got %.2f", scoreExact)
	}
	if len(matchesExact) == 0 {
		t.Errorf("Expected non-empty match reasons for known malicious IP")
	}

	// 2. Subnet CIDR match within malicious range (198.51.100.0/24)
	reportCIDR := engine.EvaluateThreatContext("198.51.100.188", "dev_browser", "user@company.com", nil, time.Time{}, time.Time{})
	if !reportCIDR.IsMaliciousIP {
		t.Errorf("Expected IsMaliciousIP = true for IP inside 198.51.100.0/24 CIDR")
	}

	// 3. Tor Exit Node CIDR match (185.220.101.55 inside 185.220.101.0/24)
	reportTor := engine.EvaluateThreatContext("185.220.101.55", "dev_browser", "user@company.com", nil, time.Time{}, time.Time{})
	if !reportTor.IsMaliciousIP {
		t.Errorf("Expected Tor exit node subnet IP to match malicious CIDR")
	}

	// 4. Private / Loopback IPs (Safe local default)
	locLoopback, okLoop := engine.ResolveLocation("127.0.0.1")
	if !okLoop || locLoopback.CountryCode != "IN" {
		t.Errorf("Expected loopback IP to resolve safely to local default")
	}

	locPrivate, okPriv := engine.ResolveLocation("192.168.1.100")
	if !okPriv || locPrivate.IsProxy {
		t.Errorf("Expected private IP to resolve cleanly without proxy flag")
	}

	// 5. Malformed IP strings
	_, okBad1 := engine.ResolveLocation("999.999.999.999")
	if okBad1 {
		t.Errorf("Expected invalid IP 999.999.999.999 to return ok=false")
	}

	_, okBad2 := engine.ResolveLocation("not_an_ip")
	if okBad2 {
		t.Errorf("Expected text 'not_an_ip' to return ok=false")
	}

	_, okBad3 := engine.ResolveLocation("")
	if okBad3 {
		t.Errorf("Expected empty string to return ok=false")
	}
}

// TestGraphBFSCycleAndExpansionLimits verifies bounded 3-hop traversal and cyclic loop handling.
func TestGraphBFSCycleAndExpansionLimits(t *testing.T) {
	engine := NewGraphEngine(nil)

	// Ingest a cyclic graph: Account A -> Device D1 -> Account B -> IP I1 -> Account A
	_ = engine.IngestTransactionLinks("txn_c1", "usr_A", "acc_A", "tok_A", "dev_D1", "106.51.0.1", "merch_1", 1000, false)
	_ = engine.IngestTransactionLinks("txn_c2", "usr_B", "acc_B", "tok_B", "dev_D1", "106.51.0.2", "merch_1", 1000, false)
	_ = engine.IngestTransactionLinks("txn_c3", "usr_B", "acc_B", "tok_B", "dev_D2", "106.51.0.1", "merch_1", 1000, false)

	// Traversal should terminate deterministically without infinite loop
	evidence := engine.EvaluateEntityGraph("acc_A", "dev_D1", "tok_A", "106.51.0.1")
	if evidence.ConnectedAccountCount != 1 {
		t.Errorf("Expected 1 connected account (acc_B), got %d", evidence.ConnectedAccountCount)
	}

	// Large cluster with 80 connected nodes: should cap at MaxTotalNodesExpansion (50)
	largeEngine := NewGraphEngine(nil)
	sharedMuleDev := "dev_massive_mule_ring"
	for i := 1; i <= 80; i++ {
		_ = largeEngine.IngestTransactionLinks(
			fmt.Sprintf("txn_ring_%d", i),
			fmt.Sprintf("usr_ring_%d", i),
			fmt.Sprintf("acc_ring_%d", i),
			fmt.Sprintf("tok_ring_%d", i),
			sharedMuleDev,
			"198.51.100.44",
			"merch_cashout",
			5000,
			false,
		)
	}

	largeEvidence := largeEngine.EvaluateEntityGraph("acc_ring_1", sharedMuleDev, "tok_ring_1", "198.51.100.44")
	if largeEvidence.VisitedNodesCount > MaxTotalNodesExpansion {
		t.Errorf("Expansion exceeded bound: visited %d nodes (max %d)", largeEvidence.VisitedNodesCount, MaxTotalNodesExpansion)
	}
}
