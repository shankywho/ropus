package rules

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Combinator defines logical aggregation for child rules.
type Combinator string

const (
	CombinatorAnd Combinator = "AND"
	CombinatorOr  Combinator = "OR"
	CombinatorNot Combinator = "NOT"
)

// RuleDefinition represents the full rule specification including condition AST and outcome metadata.
type RuleDefinition struct {
	Condition  ASTNode `json:"condition"`
	Action     string  `json:"action"`      // e.g., DECLINE_RECOMMENDATION, MANUAL_REVIEW, STEP_UP_RECOMMENDATION
	ReasonCode string  `json:"reason_code"` // e.g., HIGH_VELOCITY_IP
	Priority   int     `json:"priority"`    // Evaluation priority (lower = evaluated first)
}

// ASTNode represents either a composite logical node (with Combinator and Rules) or a leaf condition (Field, Operator, Value).
type ASTNode struct {
	Combinator Combinator  `json:"combinator,omitempty"` // "AND", "OR", "NOT"
	Rules      []ASTNode   `json:"rules,omitempty"`      // Child nodes
	Field      string      `json:"field,omitempty"`      // Context field name (e.g., "amount", "velocity.ip.1hr")
	Operator   string      `json:"operator,omitempty"`   // "==", "!=", ">", "<", ">=", "<=", "IN", "NOT_IN", "CONTAINS"
	Value      interface{} `json:"value,omitempty"`      // Expected value or list of values
}

// ParseRuleDefinition parses raw JSON bytes into a RuleDefinition.
func ParseRuleDefinition(raw []byte) (*RuleDefinition, error) {
	var ruleDef RuleDefinition
	if err := json.Unmarshal(raw, &ruleDef); err != nil {
		return nil, fmt.Errorf("failed to parse rule definition JSON: %w", err)
	}

	// If top-level AST node has no condition wrapper, attempt unmarshaling as direct ASTNode
	if ruleDef.Condition.Combinator == "" && ruleDef.Condition.Field == "" && len(ruleDef.Condition.Rules) == 0 {
		var directNode ASTNode
		if err := json.Unmarshal(raw, &directNode); err == nil && (directNode.Combinator != "" || directNode.Field != "") {
			ruleDef.Condition = directNode
		}
	}

	return &ruleDef, nil
}

// Evaluate evaluates an ASTNode against a key-value context map.
func (node *ASTNode) Evaluate(ctx map[string]interface{}) (bool, error) {
	// 1. Composite Logical Node
	if node.Combinator != "" || len(node.Rules) > 0 {
		combinator := Combinator(strings.ToUpper(string(node.Combinator)))
		if combinator == "" {
			combinator = CombinatorAnd
		}

		switch combinator {
		case CombinatorAnd:
			if len(node.Rules) == 0 {
				return true, nil
			}
			for _, child := range node.Rules {
				matched, err := child.Evaluate(ctx)
				if err != nil {
					return false, err
				}
				if !matched {
					return false, nil
				}
			}
			return true, nil

		case CombinatorOr:
			if len(node.Rules) == 0 {
				return false, nil
			}
			for _, child := range node.Rules {
				matched, err := child.Evaluate(ctx)
				if err != nil {
					return false, err
				}
				if matched {
					return true, nil
				}
			}
			return false, nil

		case CombinatorNot:
			if len(node.Rules) == 0 {
				return true, nil
			}
			matched, err := node.Rules[0].Evaluate(ctx)
			if err != nil {
				return false, err
			}
			return !matched, nil

		default:
			return false, fmt.Errorf("unsupported logical combinator: %s", node.Combinator)
		}
	}

	// 2. Leaf Condition Node
	if node.Field == "" {
		return false, fmt.Errorf("leaf AST node missing 'field' specification")
	}

	actualVal, exists := resolveField(ctx, node.Field)
	if !exists {
		// Field not present in context: evaluate if condition expects null or return false
		if strings.ToUpper(node.Operator) == "==" && node.Value == nil {
			return true, nil
		}
		if strings.ToUpper(node.Operator) == "!=" && node.Value != nil {
			return true, nil
		}
		return false, nil
	}

	return evaluateOperator(actualVal, node.Operator, node.Value)
}

// resolveField extracts a value from context supporting both flat keys and dot-path navigation.
func resolveField(ctx map[string]interface{}, path string) (interface{}, bool) {
	// First, check direct key lookup (e.g., "velocity.ip.1hr" as flat key)
	if val, ok := ctx[path]; ok {
		return val, true
	}

	// Second, traverse dot notation
	parts := strings.Split(path, ".")
	var current interface{} = ctx

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		current = val
	}

	return current, true
}

// evaluateOperator evaluates comparison operators between actual and expected values.
func evaluateOperator(actual interface{}, operator string, expected interface{}) (bool, error) {
	op := strings.ToUpper(strings.TrimSpace(operator))

	switch op {
	case "==", "EQ", "EQUALS":
		return areEqual(actual, expected), nil

	case "!=", "NEQ", "NOT_EQUALS":
		return !areEqual(actual, expected), nil

	case ">", "GT":
		actNum, actOk := toFloat64(actual)
		expNum, expOk := toFloat64(expected)
		if actOk && expOk {
			return actNum > expNum, nil
		}
		return false, fmt.Errorf("operator '>' requires numeric operands, got actual=%T, expected=%T", actual, expected)

	case ">=", "GTE":
		actNum, actOk := toFloat64(actual)
		expNum, expOk := toFloat64(expected)
		if actOk && expOk {
			return actNum >= expNum, nil
		}
		return false, fmt.Errorf("operator '>=' requires numeric operands, got actual=%T, expected=%T", actual, expected)

	case "<", "LT":
		actNum, actOk := toFloat64(actual)
		expNum, expOk := toFloat64(expected)
		if actOk && expOk {
			return actNum < expNum, nil
		}
		return false, fmt.Errorf("operator '<' requires numeric operands, got actual=%T, expected=%T", actual, expected)

	case "<=", "LTE":
		actNum, actOk := toFloat64(actual)
		expNum, expOk := toFloat64(expected)
		if actOk && expOk {
			return actNum <= expNum, nil
		}
		return false, fmt.Errorf("operator '<=' requires numeric operands, got actual=%T, expected=%T", actual, expected)

	case "IN":
		return isContainedIn(actual, expected), nil

	case "NOT_IN":
		return !isContainedIn(actual, expected), nil

	case "CONTAINS":
		return containsSubstringOrItem(actual, expected), nil

	case "NOT_CONTAINS":
		return !containsSubstringOrItem(actual, expected), nil

	case "STARTS_WITH":
		actStr := fmt.Sprintf("%v", actual)
		expStr := fmt.Sprintf("%v", expected)
		return strings.HasPrefix(actStr, expStr), nil

	case "ENDS_WITH":
		actStr := fmt.Sprintf("%v", actual)
		expStr := fmt.Sprintf("%v", expected)
		return strings.HasSuffix(actStr, expStr), nil

	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// areEqual checks equality across JSON numbers, strings, bools, etc.
func areEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	numA, okA := toFloat64(a)
	numB, okB := toFloat64(b)
	if okA && okB {
		return numA == numB
	}

	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)
	return strA == strB
}

// toFloat64 converts generic types to float64 for numeric comparisons.
func toFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// isContainedIn checks if needle 'actual' is within haystack 'expected' slice/array.
func isContainedIn(actual interface{}, expected interface{}) bool {
	if expected == nil {
		return false
	}

	val := reflect.ValueOf(expected)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return areEqual(actual, expected)
	}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface()
		if areEqual(actual, elem) {
			return true
		}
	}

	return false
}

// containsSubstringOrItem checks if 'actual' (string or slice) contains 'expected'.
func containsSubstringOrItem(actual interface{}, expected interface{}) bool {
	if actual == nil || expected == nil {
		return false
	}

	// String contains check
	if actStr, ok := actual.(string); ok {
		expStr := fmt.Sprintf("%v", expected)
		return strings.Contains(actStr, expStr)
	}

	// Slice contains check
	val := reflect.ValueOf(actual)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			if areEqual(elem, expected) {
				return true
			}
		}
	}

	return false
}
