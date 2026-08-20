package rules_test

import (
	"encoding/json"
	"testing"

	"github.com/shankywho/ropus/backend/internal/rules"
)

func TestAST_FieldOperators(t *testing.T) {
	tests := []struct {
		name     string
		node     rules.ASTNode
		ctx      map[string]interface{}
		expected bool
	}{
		{
			name: "Numeric Greater Than (Int) - True",
			node: rules.ASTNode{Field: "amount", Operator: ">", Value: 5000},
			ctx:  map[string]interface{}{"amount": 7500},
			expected: true,
		},
		{
			name: "Numeric Greater Than - False",
			node: rules.ASTNode{Field: "amount", Operator: ">", Value: 10000},
			ctx:  map[string]interface{}{"amount": 5000},
			expected: false,
		},
		{
			name: "Numeric Greater Than or Equal (GTE) - True",
			node: rules.ASTNode{Field: "velocity.ip.1hr", Operator: ">=", Value: 5},
			ctx:  map[string]interface{}{"velocity.ip.1hr": 5},
			expected: true,
		},
		{
			name: "Numeric Less Than (LT) - True",
			node: rules.ASTNode{Field: "risk_score", Operator: "<", Value: 30},
			ctx:  map[string]interface{}{"risk_score": 15},
			expected: true,
		},
		{
			name: "Numeric Less Than or Equal (LTE) - True",
			node: rules.ASTNode{Field: "risk_score", Operator: "<=", Value: 30},
			ctx:  map[string]interface{}{"risk_score": 30},
			expected: true,
		},
		{
			name: "String Equality - True",
			node: rules.ASTNode{Field: "currency", Operator: "==", Value: "INR"},
			ctx:  map[string]interface{}{"currency": "INR"},
			expected: true,
		},
		{
			name: "String Inequality - True",
			node: rules.ASTNode{Field: "currency", Operator: "!=", Value: "USD"},
			ctx:  map[string]interface{}{"currency": "INR"},
			expected: true,
		},
		{
			name: "IN Operator (Slice) - True",
			node: rules.ASTNode{Field: "payment_method.type", Operator: "IN", Value: []string{"card", "upi"}},
			ctx:  map[string]interface{}{"payment_method": map[string]interface{}{"type": "upi"}},
			expected: true,
		},
		{
			name: "IN Operator (Slice) - False",
			node: rules.ASTNode{Field: "country", Operator: "IN", Value: []string{"US", "GB"}},
			ctx:  map[string]interface{}{"country": "IN"},
			expected: false,
		},
		{
			name: "NOT_IN Operator - True",
			node: rules.ASTNode{Field: "country", Operator: "NOT_IN", Value: []string{"US", "GB"}},
			ctx:  map[string]interface{}{"country": "IN"},
			expected: true,
		},
		{
			name: "CONTAINS Operator (Substring) - True",
			node: rules.ASTNode{Field: "email", Operator: "CONTAINS", Value: "@fraudster.com"},
			ctx:  map[string]interface{}{"email": "attacker@fraudster.com"},
			expected: true,
		},
		{
			name: "STARTS_WITH Operator - True",
			node: rules.ASTNode{Field: "bin", Operator: "STARTS_WITH", Value: "4111"},
			ctx:  map[string]interface{}{"bin": "411199998888"},
			expected: true,
		},
		{
			name: "ENDS_WITH Operator - True",
			node: rules.ASTNode{Field: "email", Operator: "ENDS_WITH", Value: ".ru"},
			ctx:  map[string]interface{}{"email": "suspicious@domain.ru"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := tt.node.Evaluate(tt.ctx)
			if err != nil {
				t.Fatalf("unexpected error during evaluation: %v", err)
			}
			if matched != tt.expected {
				t.Errorf("expected matched=%v, got %v", tt.expected, matched)
			}
		})
	}
}

func TestAST_NestedLogicalCombinators(t *testing.T) {
	// Rule: (amount > 5000 AND velocity.ip.1hr >= 5) OR (payment_method.type == 'crypto' AND amount > 1000)
	ruleJSON := `{
		"combinator": "OR",
		"rules": [
			{
				"combinator": "AND",
				"rules": [
					{ "field": "amount", "operator": ">", "value": 5000 },
					{ "field": "features.ipTxnCount1h", "operator": ">=", "value": 5 }
				]
			},
			{
				"combinator": "AND",
				"rules": [
					{ "field": "payment_method.type", "operator": "==", "value": "crypto" },
					{ "field": "amount", "operator": ">", "value": 1000 }
				]
			}
		]
	}`

	var rootNode rules.ASTNode
	if err := json.Unmarshal([]byte(ruleJSON), &rootNode); err != nil {
		t.Fatalf("failed to unmarshal test JSON AST: %v", err)
	}

	// Case 1: First branch matches
	ctx1 := map[string]interface{}{
		"amount": 10000,
		"features": map[string]interface{}{
			"ipTxnCount1h": 6,
		},
		"payment_method": map[string]interface{}{
			"type": "card",
		},
	}
	matched1, err := rootNode.Evaluate(ctx1)
	if err != nil || !matched1 {
		t.Errorf("expected Case 1 to match, got %v (err: %v)", matched1, err)
	}

	// Case 2: Second branch matches
	ctx2 := map[string]interface{}{
		"amount": 2000,
		"features": map[string]interface{}{
			"ipTxnCount1h": 1,
		},
		"payment_method": map[string]interface{}{
			"type": "crypto",
		},
	}
	matched2, err := rootNode.Evaluate(ctx2)
	if err != nil || !matched2 {
		t.Errorf("expected Case 2 to match, got %v (err: %v)", matched2, err)
	}

	// Case 3: Neither branch matches
	ctx3 := map[string]interface{}{
		"amount": 500,
		"features": map[string]interface{}{
			"ipTxnCount1h": 1,
		},
		"payment_method": map[string]interface{}{
			"type": "card",
		},
	}
	matched3, err := rootNode.Evaluate(ctx3)
	if err != nil || matched3 {
		t.Errorf("expected Case 3 to NOT match, got %v (err: %v)", matched3, err)
	}
}

func TestAST_ParseRuleDefinition(t *testing.T) {
	raw := []byte(`{
		"condition": {
			"combinator": "AND",
			"rules": [
				{ "field": "amount", "operator": ">", "value": 100000 },
				{ "field": "currency", "operator": "==", "value": "INR" }
			]
		},
		"action": "DECLINE_RECOMMENDATION",
		"reason_code": "EXCESSIVE_INR_AMOUNT",
		"priority": 10
	}`)

	ruleDef, err := rules.ParseRuleDefinition(raw)
	if err != nil {
		t.Fatalf("failed to parse rule definition: %v", err)
	}

	if ruleDef.Action != "DECLINE_RECOMMENDATION" {
		t.Errorf("expected action DECLINE_RECOMMENDATION, got %s", ruleDef.Action)
	}
	if ruleDef.ReasonCode != "EXCESSIVE_INR_AMOUNT" {
		t.Errorf("expected reason_code EXCESSIVE_INR_AMOUNT, got %s", ruleDef.ReasonCode)
	}

	ctxMatch := map[string]interface{}{"amount": 250000, "currency": "INR"}
	matched, err := ruleDef.Condition.Evaluate(ctxMatch)
	if err != nil || !matched {
		t.Errorf("expected condition to match for ctxMatch, got %v (err: %v)", matched, err)
	}
}
