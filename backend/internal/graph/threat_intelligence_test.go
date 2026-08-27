package graph

import (
	"testing"
	"time"
)

func TestCalculateHaversineDistance(t *testing.T) {
	// Bengaluru (12.9716, 77.5946) to Limassol (34.6841, 33.0379) ~5,300 km
	blr := GeoPoint{Lat: 12.9716, Lon: 77.5946}
	lms := GeoPoint{Lat: 34.6841, Lon: 33.0379}

	dist := CalculateHaversineDistanceKm(blr, lms)
	if dist < 5000 || dist > 5600 {
		t.Errorf("Expected distance between Bengaluru and Limassol ~5300km, got %.2f", dist)
	}

	// Zero / identical distance
	if zeroDist := CalculateHaversineDistanceKm(blr, blr); zeroDist != 0.0 {
		t.Errorf("Expected identical points distance 0, got %.2f", zeroDist)
	}
}

func TestCalculateImpliedSpeed(t *testing.T) {
	blr := GeoPoint{Lat: 12.9716, Lon: 77.5946}
	lms := GeoPoint{Lat: 34.6841, Lon: 33.0379}

	// 12 minutes elapsed
	dist, speed := CalculateImpliedSpeedKmh(blr, lms, 12*time.Minute)
	if dist < 5000 {
		t.Errorf("Expected distance > 5000km, got %.2f", dist)
	}
	if speed < 20000 { // 5300 km / 0.2h = 26,500 km/h
		t.Errorf("Expected speed > 20,000 km/h, got %.2f", speed)
	}
}

func TestThreatIntelligenceEngine_EvaluateThreatContext(t *testing.T) {
	engine := NewThreatIntelligenceEngine()

	blrLocation := GeoPoint{Lat: 12.9716, Lon: 77.5946}
	blrTime := time.Now().Add(-12 * time.Minute)
	now := time.Now()

	// Scenario 1: Clean Domestic Residential Ingress
	cleanReport := engine.EvaluateThreatContext("106.51.12.34", "dev_safari_ios_clean", "gmail.com", &blrLocation, blrTime, now)
	if cleanReport.IsMaliciousIP || cleanReport.IsCompromisedDev || cleanReport.IsImpossibleTrip {
		t.Errorf("Expected clean report for residential Jio IP, got %+v", cleanReport)
	}
	if cleanReport.OriginCountry != "IN" {
		t.Errorf("Expected country IN, got %s", cleanReport.OriginCountry)
	}

	// Scenario 2: Datacenter Proxy + Impossible Travel
	attackReport := engine.EvaluateThreatContext("198.51.100.44", "dev_emulator_compromised", "temp-mail.org", &blrLocation, blrTime, now)
	if !attackReport.IsMaliciousIP {
		t.Errorf("Expected IsMaliciousIP = true")
	}
	if !attackReport.IsCompromisedDev {
		t.Errorf("Expected IsCompromisedDev = true")
	}
	if !attackReport.IsProxyDatacenter {
		t.Errorf("Expected IsProxyDatacenter = true")
	}
	if !attackReport.IsImpossibleTrip {
		t.Errorf("Expected IsImpossibleTrip = true for 5300km in 12min")
	}
	if attackReport.RiskScore < 0.85 {
		t.Errorf("Expected high risk score >= 0.85, got %.2f", attackReport.RiskScore)
	}
}
