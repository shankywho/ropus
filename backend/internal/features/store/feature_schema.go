package store

import (
	"fmt"
	"time"
)

// DataType specifies the primitive feature value type.
type DataType string

const (
	TypeFloat64 DataType = "FLOAT64"
	TypeInt64   DataType = "INT64"
	TypeString  DataType = "STRING"
	TypeBool    DataType = "BOOL"
)

// FeatureDefinition defines the metadata, validation constraints, and lineage of a feature.
type FeatureDefinition struct {
	Name         string            `json:"name"`
	Version      int               `json:"version"`
	DataType     DataType          `json:"data_type"`
	Description  string            `json:"description"`
	DefaultValue interface{}       `json:"default_value"`
	SourceEntity string            `json:"source_entity"` // "transaction", "device", "user", "ip"
	CreatedAt    time.Time         `json:"created_at"`
	Tags         []string          `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ValidateValue checks if a given value complies with the feature definition data type.
func (f *FeatureDefinition) ValidateValue(val interface{}) error {
	if val == nil {
		return nil // Missing values allowed
	}
	switch f.DataType {
	case TypeFloat64:
		switch val.(type) {
		case float64, float32, int, int64, int32:
			return nil
		default:
			return fmt.Errorf("feature %s expects float64, got %T", f.Name, val)
		}
	case TypeInt64:
		switch val.(type) {
		case int, int64, int32:
			return nil
		default:
			return fmt.Errorf("feature %s expects int64, got %T", f.Name, val)
		}
	case TypeString:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("feature %s expects string, got %T", f.Name, val)
		}
	case TypeBool:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("feature %s expects bool, got %T", f.Name, val)
		}
	}
	return nil
}
