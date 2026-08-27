package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProductionSafetyConfig(t *testing.T) {
	// 1. Valid production shadow soak configuration
	validCfg := Config{
		ShadowEnabled:               true,
		ShadowCandidateModelVersion: "extended_catboost_58f",
		CanaryEnabled:               false,
		CanaryPercentage:            0,
	}
	require.NoError(t, ValidateProductionSafetyConfig(validCfg))

	// 2. Invalid: Canary enabled with percentage > 0 during shadow soak
	invalidCanaryCfg := Config{
		ShadowEnabled:               true,
		ShadowCandidateModelVersion: "extended_catboost_58f",
		CanaryEnabled:               true,
		CanaryPercentage:            10,
	}
	errCanary := ValidateProductionSafetyConfig(invalidCanaryCfg)
	require.Error(t, errCanary)
	assert.Contains(t, errCanary.Error(), "CANARY_ROUTING_FORBIDDEN")

	// 3. Invalid: Shadow enabled but missing candidate model version
	invalidShadowCfg := Config{
		ShadowEnabled:               true,
		ShadowCandidateModelVersion: "",
		CanaryEnabled:               false,
		CanaryPercentage:            0,
	}
	errShadow := ValidateProductionSafetyConfig(invalidShadowCfg)
	require.Error(t, errShadow)
	assert.Contains(t, errShadow.Error(), "SHADOW_CANDIDATE_MODEL_REQUIRED")
}
