package riskengine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestKubernetesManifests_SyntacticAndStructuralValidation(t *testing.T) {
	manifestDir := "../../../deploy/kubernetes"
	files, err := os.ReadDir(manifestDir)
	require.NoError(t, err, "Kubernetes deploy directory must exist")

	expectedManifests := map[string]bool{
		"namespace.yaml":       false,
		"service-account.yaml": false,
		"configmap.yaml":       false,
		"secret-template.yaml": false,
		"deployment.yaml":      false,
		"service.yaml":         false,
		"ingress.yaml":         false,
		"hpa.yaml":             false,
		"pdb.yaml":             false,
		"network-policy.yaml":  false,
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			filePath := filepath.Join(manifestDir, file.Name())
			content, readErr := os.ReadFile(filePath)
			require.NoError(t, readErr)

			var parsed map[string]interface{}
			err = yaml.Unmarshal(content, &parsed)
			assert.NoError(t, err, "Manifest %s must be valid YAML", file.Name())
			assert.NotEmpty(t, parsed["apiVersion"], "Manifest %s must have apiVersion", file.Name())
			assert.NotEmpty(t, parsed["kind"], "Manifest %s must have kind", file.Name())

			expectedManifests[file.Name()] = true
		}
	}

	for name, found := range expectedManifests {
		assert.True(t, found, "Expected manifest %s was not found in /deploy/kubernetes", name)
	}
}

func TestHelmValues_ValidYAML(t *testing.T) {
	valuesFile := "../../../deploy/helm/risk-manager/values.yaml"
	content, err := os.ReadFile(valuesFile)
	require.NoError(t, err)

	var values map[string]interface{}
	err = yaml.Unmarshal(content, &values)
	require.NoError(t, err, "Helm values.yaml must be valid YAML")

	assert.Equal(t, 3, values["replicaCount"])
	assert.NotNil(t, values["resources"])
	assert.NotNil(t, values["autoscaling"])
	assert.NotNil(t, values["podSecurityContext"])
}
