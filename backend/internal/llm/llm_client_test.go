package llm

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestLLMClient_DeterministicFallbackWhenUnconfigured(t *testing.T) {
	client := NewLLMClient("", "", "")
	if client.IsConfigured() {
		t.Errorf("Expected IsConfigured = false when no API key is passed")
	}

	resp, err := client.GenerateCompletion(context.Background(), []LLMMessage{
		{Role: "user", Content: "Investigate transaction txn_123"},
	})
	if err != nil {
		t.Fatalf("GenerateCompletion returned error: %v", err)
	}

	if !resp.IsDegraded {
		t.Errorf("Expected IsDegraded = true for deterministic fallback")
	}
	if resp.TokensUsed != 0 {
		t.Errorf("Expected 0 tokens used for deterministic fallback, got %d", resp.TokensUsed)
	}
	if resp.Model != "deterministic_forensic_engine_v1" {
		t.Errorf("Expected model 'deterministic_forensic_engine_v1', got %s", resp.Model)
	}
}

func TestLLMClient_RealProviderSuccessWithMock(t *testing.T) {
	client := NewLLMClient("https://api.anthropic.com/v1", "test-key-xyz", "claude-3-7-sonnet-20250219")
	if !client.IsConfigured() {
		t.Errorf("Expected IsConfigured = true")
	}

	mockJSON := `{
		"id": "chatcmpl-991823",
		"choices": [
			{
				"message": {
					"role": "assistant",
					"content": "Forensic analysis: High risk detected."
				}
			}
		],
		"usage": {
			"prompt_tokens": 120,
			"completion_tokens": 45,
			"total_tokens": 165
		}
	}`

	client.SetHTTPClient(&mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") != "Bearer test-key-xyz" {
				t.Errorf("Missing or incorrect Authorization header: %s", req.Header.Get("Authorization"))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(mockJSON))),
			}, nil
		},
	})

	resp, err := client.GenerateCompletion(context.Background(), []LLMMessage{
		{Role: "user", Content: "Analyze txn_99"},
	})
	if err != nil {
		t.Fatalf("GenerateCompletion returned error: %v", err)
	}

	if resp.IsDegraded {
		t.Errorf("Expected IsDegraded = false for successful mock provider call")
	}
	if resp.TokensUsed != 165 {
		t.Errorf("Expected 165 total tokens, got %d", resp.TokensUsed)
	}
	if resp.Content != "Forensic analysis: High risk detected." {
		t.Errorf("Unexpected content: %s", resp.Content)
	}
}
