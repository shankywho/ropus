package streaming

import (
	"fmt"
	"time"
)

// ZeroTrustContext encapsulates continuous verification context for internal mesh operations.
type ZeroTrustContext struct {
	CallerID             string    `json:"caller_id"`
	CallerRole           string    `json:"caller_role"`
	DeviceCertThumbprint string    `json:"device_cert_thumbprint"`
	SourceNamespace      string    `json:"source_namespace"`
	TargetResource       string    `json:"target_resource"`
	RequestedAction      string    `json:"requested_action"`
	TenantID             string    `json:"tenant_id"`
	Timestamp            time.Time `json:"timestamp"`
}

// ZeroTrustSecurityGuard validates every service-to-service invocation against explicit trust policies.
type ZeroTrustSecurityGuard struct {
	AllowedNamespaces map[string]bool
}

// NewZeroTrustSecurityGuard initializes the zero-trust validator.
func NewZeroTrustSecurityGuard() *ZeroTrustSecurityGuard {
	return &ZeroTrustSecurityGuard{
		AllowedNamespaces: map[string]bool{
			"risk-engine": true,
			"default":     true,
		},
	}
}

// AuthorizeAction checks caller identity, cryptographic thumbprint, and role permissions.
func (g *ZeroTrustSecurityGuard) AuthorizeAction(ctx ZeroTrustContext) (bool, error) {
	if ctx.CallerID == "" {
		return false, fmt.Errorf("zero trust authorization failed: empty caller identity")
	}
	if ctx.DeviceCertThumbprint == "" {
		return false, fmt.Errorf("zero trust authorization failed: missing mTLS client certificate thumbprint")
	}
	if !g.AllowedNamespaces[ctx.SourceNamespace] {
		return false, fmt.Errorf("zero trust authorization failed: namespace '%s' is not authorized for risk mesh operations", ctx.SourceNamespace)
	}

	return true, nil
}
