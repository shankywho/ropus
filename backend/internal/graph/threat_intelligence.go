package graph

import (
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"
)

// GeoPoint represents geographical coordinates.
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// IPLocationInfo holds resolved geographic and network metadata for an IP.
type IPLocationInfo struct {
	Point        GeoPoint `json:"point"`
	CountryCode  string   `json:"country_code"`
	City         string   `json:"city"`
	ASN          string   `json:"asn"`
	IsProxy      bool     `json:"is_proxy"`
	IsDatacenter bool     `json:"is_datacenter"`
}

// ThreatReport captures the complete calculated threat evaluation for an ingress transaction.
type ThreatReport struct {
	RiskScore         float64         `json:"risk_score"`
	IsMaliciousIP     bool            `json:"is_malicious_ip"`
	IsCompromisedDev  bool            `json:"is_compromised_dev"`
	IsProxyDatacenter bool            `json:"is_proxy_datacenter"`
	IsImpossibleTrip  bool            `json:"is_impossible_travel"`
	GeoDistanceKm     float64         `json:"geo_distance_km"`
	ImpliedSpeedKmh   float64         `json:"implied_speed_kmh"`
	OriginCountry     string          `json:"origin_country"`
	OriginCity        string          `json:"origin_city"`
	ASN               string          `json:"asn"`
	Location          *IPLocationInfo `json:"location,omitempty"`
	Matches           []string        `json:"matches"`
}

// ThreatIntelligenceEngine maintains dynamic feeds of malicious network IOCs, CIDRs, and geo resolvers.
type ThreatIntelligenceEngine struct {
	mu                 sync.RWMutex
	maliciousIPs       map[string]float64
	maliciousCIDRs     []*net.IPNet
	compromisedDevices map[string]float64
	riskyDomains       map[string]float64
	knownGeoIPs        map[string]IPLocationInfo
	knownGeoSubnets    []struct {
		Subnet *net.IPNet
		Info   IPLocationInfo
	}
}

// NewThreatIntelligenceEngine initializes the threat intelligence engine with baseline IOC feeds and geo resolver.
func NewThreatIntelligenceEngine() *ThreatIntelligenceEngine {
	t := &ThreatIntelligenceEngine{
		maliciousIPs:       make(map[string]float64),
		maliciousCIDRs:     make([]*net.IPNet, 0),
		compromisedDevices: make(map[string]float64),
		riskyDomains:       make(map[string]float64),
		knownGeoIPs:        make(map[string]IPLocationInfo),
	}
	t.initializeBaselineFeeds()
	return t
}

func (t *ThreatIntelligenceEngine) initializeBaselineFeeds() {
	// Exact malicious IPs
	t.maliciousIPs["198.51.100.44"] = 0.95 // Known Bulletproof proxy
	t.maliciousIPs["203.0.113.88"] = 0.99  // Tor exit node cluster
	t.maliciousIPs["203.0.113.195"] = 0.92 // Commercial VPN scraper subnet

	// CIDR blocks
	for _, cidr := range []string{"198.51.100.0/24", "203.0.113.0/24", "185.220.101.0/24"} {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
			t.maliciousCIDRs = append(t.maliciousCIDRs, ipNet)
		}
	}

	// Compromised devices & root emulator signatures
	t.compromisedDevices["dev_emul_root_89a"] = 0.90
	t.compromisedDevices["dev_emulator_compromised"] = 0.95
	t.compromisedDevices["dev_mule_cluster_99"] = 0.92
	t.compromisedDevices["dev_emulator_linux_9f8a"] = 0.90

	// Disposable domains
	t.riskyDomains["temp-mail.org"] = 0.85
	t.riskyDomains["disposable-inbox.com"] = 0.90
	t.riskyDomains["guerrillamail.com"] = 0.88
	t.riskyDomains["10minutemail.com"] = 0.85

	// Baseline GeoIP Subnet and IP directory
	t.registerGeoSubnet("198.51.100.0/24", IPLocationInfo{
		Point:        GeoPoint{Lat: 34.6841, Lon: 33.0379}, // Limassol, Cyprus
		CountryCode:  "CY",
		City:         "Limassol",
		ASN:          "AS13335 (Datacenter / Proxy)",
		IsProxy:      true,
		IsDatacenter: true,
	})
	t.registerGeoSubnet("203.0.113.0/24", IPLocationInfo{
		Point:        GeoPoint{Lat: 51.5074, Lon: -0.1278}, // London, UK
		CountryCode:  "GB",
		City:         "London",
		ASN:          "AS9009 (Bulletproof)",
		IsProxy:      true,
		IsDatacenter: true,
	})
	t.registerGeoSubnet("106.51.0.0/16", IPLocationInfo{
		Point:        GeoPoint{Lat: 12.9716, Lon: 77.5946}, // Bengaluru, KA, India
		CountryCode:  "IN",
		City:         "Bengaluru",
		ASN:          "AS55836 (Jio Fiber Residential)",
		IsProxy:      false,
		IsDatacenter: false,
	})
	t.registerGeoSubnet("122.160.0.0/16", IPLocationInfo{
		Point:        GeoPoint{Lat: 28.6139, Lon: 77.2090}, // New Delhi, India
		CountryCode:  "IN",
		City:         "New Delhi",
		ASN:          "AS24560 (Airtel Broadband)",
		IsProxy:      false,
		IsDatacenter: false,
	})
	t.registerGeoSubnet("12.0.0.0/8", IPLocationInfo{
		Point:        GeoPoint{Lat: 37.7749, Lon: -122.4194}, // San Francisco, CA, USA
		CountryCode:  "US",
		City:         "San Francisco",
		ASN:          "AS7018 (AT&T Services)",
		IsProxy:      false,
		IsDatacenter: false,
	})
}

func (t *ThreatIntelligenceEngine) registerGeoSubnet(cidr string, info IPLocationInfo) {
	if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
		t.knownGeoSubnets = append(t.knownGeoSubnets, struct {
			Subnet *net.IPNet
			Info   IPLocationInfo
		}{Subnet: ipNet, Info: info})
	}
}

// ResolveLocation maps an IP address to geographical and network classification metadata.
func (t *ThreatIntelligenceEngine) ResolveLocation(ipStr string) (IPLocationInfo, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	parsedIP := net.ParseIP(strings.TrimSpace(ipStr))
	if parsedIP == nil {
		return IPLocationInfo{}, false
	}

	// Exact IP cache
	if info, found := t.knownGeoIPs[ipStr]; found {
		return info, true
	}

	// Subnet search
	for _, entry := range t.knownGeoSubnets {
		if entry.Subnet.Contains(parsedIP) {
			return entry.Info, true
		}
	}

	// Default fallback: Private or Loopback IP
	if parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		return IPLocationInfo{
			Point:        GeoPoint{Lat: 12.9716, Lon: 77.5946},
			CountryCode:  "IN",
			City:         "Local/Private",
			ASN:          "AS0 (Private Network)",
			IsProxy:      false,
			IsDatacenter: false,
		}, true
	}

	return IPLocationInfo{
		Point:        GeoPoint{Lat: 0, Lon: 0},
		CountryCode:  "UNKNOWN",
		City:         "Unknown",
		ASN:          "AS0 (Unknown)",
		IsProxy:      false,
		IsDatacenter: false,
	}, false
}

// CalculateHaversineDistanceKm computes great-circle distance between two geo-points in kilometers.
func CalculateHaversineDistanceKm(p1, p2 GeoPoint) float64 {
	const earthRadiusKm = 6371.0

	// If coordinates are identical or uninitialized
	if p1.Lat == p2.Lat && p1.Lon == p2.Lon {
		return 0.0
	}
	if (p1.Lat == 0 && p1.Lon == 0) || (p2.Lat == 0 && p2.Lon == 0) {
		return 0.0
	}

	dLat := (p2.Lat - p1.Lat) * (math.Pi / 180.0)
	dLon := (p2.Lon - p1.Lon) * (math.Pi / 180.0)

	lat1Rad := p1.Lat * (math.Pi / 180.0)
	lat2Rad := p2.Lat * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// CalculateImpliedSpeedKmh calculates velocity (km/h) across two geo-points and an elapsed time duration.
func CalculateImpliedSpeedKmh(p1, p2 GeoPoint, elapsed time.Duration) (float64, float64) {
	distKm := CalculateHaversineDistanceKm(p1, p2)
	if distKm <= 0 || elapsed <= 0 {
		return distKm, 0.0
	}
	hours := elapsed.Hours()
	if hours <= 0 {
		return distKm, 99999.0
	}
	return distKm, distKm / hours
}

// CheckThreat evaluates whether an incoming transaction intersects with known malicious IOCs (legacy interface).
func (t *ThreatIntelligenceEngine) CheckThreat(ipAddress, deviceFingerprint, emailDomain string) (float64, []string) {
	report := t.EvaluateThreatContext(ipAddress, deviceFingerprint, emailDomain, nil, time.Time{}, time.Time{})
	return report.RiskScore, report.Matches
}

// AddMaliciousIP allows dynamic threat feed ingestion.
func (t *ThreatIntelligenceEngine) AddMaliciousIP(ip string, score float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maliciousIPs[ip] = score
}

// EvaluateThreatContext computes full threat signals including CIDR matching, device compromise, and velocity.
func (t *ThreatIntelligenceEngine) EvaluateThreatContext(
	ipAddress, deviceFingerprint, emailDomain string,
	prevLocation *GeoPoint, prevTimestamp time.Time, currentTimestamp time.Time,
) *ThreatReport {
	t.mu.RLock()
	defer t.mu.RUnlock()

	report := &ThreatReport{
		Matches: make([]string, 0),
	}

	parsedIP := net.ParseIP(strings.TrimSpace(ipAddress))
	var locInfo IPLocationInfo
	var locFound bool

	if parsedIP != nil {
		for _, entry := range t.knownGeoSubnets {
			if entry.Subnet.Contains(parsedIP) {
				locInfo = entry.Info
				locFound = true
				break
			}
		}
	}

	if locFound {
		report.Location = &locInfo
		report.OriginCountry = locInfo.CountryCode
		report.OriginCity = locInfo.City
		report.ASN = locInfo.ASN
		if locInfo.IsProxy || locInfo.IsDatacenter {
			report.IsProxyDatacenter = true
			report.Matches = append(report.Matches, fmt.Sprintf("Source IP %s belongs to commercial proxy/datacenter ASN (%s)", ipAddress, locInfo.ASN))
			report.RiskScore = math.Max(report.RiskScore, 0.35)
		}
	}

	// 1. Exact Malicious IP Match
	if score, found := t.maliciousIPs[ipAddress]; found {
		report.IsMaliciousIP = true
		report.Matches = append(report.Matches, fmt.Sprintf("IP %s flagged on known threat blocklist (score: %.2f)", ipAddress, score))
		report.RiskScore = math.Max(report.RiskScore, score)
	}

	// 2. CIDR Block Threat Match
	if parsedIP != nil {
		for _, cidrNet := range t.maliciousCIDRs {
			if cidrNet.Contains(parsedIP) {
				report.IsMaliciousIP = true
				report.Matches = append(report.Matches, fmt.Sprintf("IP %s matches malicious CIDR block %s", ipAddress, cidrNet.String()))
				report.RiskScore = math.Max(report.RiskScore, 0.85)
				break
			}
		}
	}

	// 3. Compromised Device Match
	devLower := strings.ToLower(deviceFingerprint)
	if score, found := t.compromisedDevices[deviceFingerprint]; found {
		report.IsCompromisedDev = true
		report.Matches = append(report.Matches, fmt.Sprintf("Device fingerprint %s matches known rootkit/emulator IOC (score: %.2f)", deviceFingerprint, score))
		report.RiskScore = math.Max(report.RiskScore, score)
	} else if strings.Contains(devLower, "emulator") || strings.Contains(devLower, "mule") || strings.Contains(devLower, "rootkit") || strings.Contains(devLower, "genymotion") || strings.Contains(devLower, "bluestacks") || strings.Contains(devLower, "nox") {
		report.IsCompromisedDev = true
		report.Matches = append(report.Matches, fmt.Sprintf("Device fingerprint %s contains virtualized emulator IOC", deviceFingerprint))
		report.RiskScore = math.Max(report.RiskScore, 0.75)
	}

	// 4. Risky Email Domain
	if emailDomain != "" {
		if score, found := t.riskyDomains[emailDomain]; found {
			report.Matches = append(report.Matches, fmt.Sprintf("Email domain %s is a temporary disposable provider", emailDomain))
			report.RiskScore = math.Max(report.RiskScore, score*0.5)
		}
	}

	// 5. Geolocation Velocity & Impossible Travel (>900 km/h)
	if prevLocation != nil && locFound && !prevTimestamp.IsZero() && !currentTimestamp.IsZero() {
		elapsed := currentTimestamp.Sub(prevTimestamp)
		if elapsed > 0 {
			distKm, speedKmh := CalculateImpliedSpeedKmh(*prevLocation, locInfo.Point, elapsed)
			report.GeoDistanceKm = math.Round(distKm*100) / 100.0
			report.ImpliedSpeedKmh = math.Round(speedKmh*100) / 100.0

			if distKm > 300.0 && speedKmh > 900.0 {
				report.IsImpossibleTrip = true
				report.Matches = append(report.Matches, fmt.Sprintf("Impossible travel velocity: %.1f km traversed in %.1f minutes (implied speed: %.1f km/h > 900 km/h ceiling)", distKm, elapsed.Minutes(), speedKmh))
				report.RiskScore = math.Max(report.RiskScore, 0.88)
			}
		}
	}

	return report
}
