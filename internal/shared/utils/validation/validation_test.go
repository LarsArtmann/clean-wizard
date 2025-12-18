package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidator implements Validator interface for testing
type TestValidator struct {
	Name  string
	Valid bool
}

func (tv TestValidator) Validate() error {
	if !tv.Valid {
		return errors.New("invalid test validator")
	}
	if tv.Name == "" {
		return errors.New("name cannot be empty")
	}
	return nil
}

func TestValidateAndWrap(t *testing.T) {
	tests := []struct {
		name     string
		item     TestValidator
		itemType string
		wantErr  bool
	}{
		{
			name:     "valid item",
			item:     TestValidator{Name: "test", Valid: true},
			itemType: "TestValidator",
			wantErr:  false,
		},
		{
			name:     "invalid item",
			item:     TestValidator{Name: "test", Valid: false},
			itemType: "TestValidator",
			wantErr:  true,
		},
		{
			name:     "empty name",
			item:     TestValidator{Name: "", Valid: true},
			itemType: "TestValidator",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAndWrap(tt.item, tt.itemType)

			if tt.wantErr {
				assert.True(t, result.IsErr(), "Expected error but got success")
				err := result.Error()
				assert.Contains(t, err.Error(), "invalid "+tt.itemType)
			} else {
				assert.True(t, result.IsOk(), "Expected success but got error")
				value := result.Value()
				assert.Equal(t, tt.item, value)
			}
		})
	}
}

func TestValidateWithCustomError(t *testing.T) {
	validItem := TestValidator{Name: "test", Valid: true}
	invalidItem := TestValidator{Name: "test", Valid: false}

	t.Run("valid item", func(t *testing.T) {
		result := ValidateWithCustomError(validItem, "Custom validation failed")
		assert.True(t, result.IsOk())
		assert.Equal(t, validItem, result.Value())
	})

	t.Run("invalid item", func(t *testing.T) {
		result := ValidateWithCustomError(invalidItem, "Custom validation failed")
		assert.True(t, result.IsErr())
		err := result.Error()
		assert.Contains(t, err.Error(), "Custom validation failed")
	})
}

func TestValidateAndConvert(t *testing.T) {
	type ConvertedType struct {
		UppercaseName string
	}

	converter := func(tv TestValidator) ConvertedType {
		return ConvertedType{UppercaseName: tv.Name}
	}

	validItem := TestValidator{Name: "test", Valid: true}
	invalidItem := TestValidator{Name: "test", Valid: false}

	t.Run("valid conversion", func(t *testing.T) {
		result := ValidateAndConvert(validItem, converter, "TestValidator")
		require.True(t, result.IsOk())
		converted := result.Value()
		assert.Equal(t, "test", converted.UppercaseName)
	})

	t.Run("invalid conversion", func(t *testing.T) {
		result := ValidateAndConvert(invalidItem, converter, "TestValidator")
		assert.True(t, result.IsErr())
		assert.Contains(t, result.Error().Error(), "invalid TestValidator")
	})
}
