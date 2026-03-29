package context

import (
	"context"
	"testing"
)

// Test types for generic context.
type TestValidationConfig struct {
	FieldA string
	FieldB int
}

type TestErrorConfig struct {
	Code    int
	Message string
}

type TestSanitizationConfig struct {
	Policy string
	Rules  []string
}

func TestNewContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}

	c := NewContext(ctx, value)

	if c.Context != ctx {
		t.Errorf("Expected context %v, got %v", ctx, c.Context)
	}

	if c.ValueType.FieldA != "test" || c.ValueType.FieldB != 42 {
		t.Errorf("Expected value %v, got %v", value, c.ValueType)
	}

	if c.Metadata == nil {
		t.Error("Metadata should be initialized")
	}

	if c.Permissions == nil {
		t.Error("Permissions should be initialized")
	}
}

func TestWithMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	c.WithMetadata("trace_id", "12345").WithMetadata("span_id", "67890")

	if len(c.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(c.Metadata))
	}

	if val, ok := c.Metadata["trace_id"]; !ok || val != "12345" {
		t.Errorf("Expected trace_id=12345, got %v", c.Metadata)
	}

	if val, ok := c.Metadata["span_id"]; !ok || val != "67890" {
		t.Errorf("Expected span_id=67890, got %v", c.Metadata)
	}
}

func TestWithPermissions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	c.WithPermissions("read", "write", "execute")

	if len(c.Permissions) != 3 {
		t.Errorf("Expected 3 permissions, got %d", len(c.Permissions))
	}

	if !c.HasPermission("read") {
		t.Error("Expected to have read permission")
	}

	if !c.HasPermission("write") {
		t.Error("Expected to have write permission")
	}

	if !c.HasPermission("execute") {
		t.Error("Expected to have execute permission")
	}

	if c.HasPermission("admin") {
		t.Error("Should not have admin permission")
	}
}

func TestGetMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	c.Metadata["trace_id"] = "12345"
	c.Metadata["span_id"] = "67890"

	val, ok := c.GetMetadata("trace_id")
	if !ok || val != "12345" {
		t.Errorf("Expected to get trace_id=12345, got %v, %v", val, ok)
	}

	_, ok = c.GetMetadata("nonexistent")
	if ok {
		t.Error("Expected to not get nonexistent key")
	}
}

func TestClone(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	c.WithMetadata("trace_id", "12345").WithPermissions("read", "write")

	cloned := c.Clone()

	if cloned == c {
		t.Error("Clone should create a new instance")
	}

	if cloned.Context != c.Context {
		t.Error("Clone should have same context")
	}

	if len(cloned.Metadata) != len(c.Metadata) {
		t.Error("Clone should have same metadata length")
	}

	if len(cloned.Permissions) != len(c.Permissions) {
		t.Error("Clone should have same permissions length")
	}

	// Modify original to ensure deep copy
	c.WithMetadata("new_key", "new_value")

	if _, ok := cloned.Metadata["new_key"]; ok {
		t.Error("Clone should not be affected by original modifications")
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create first context
	value1 := TestValidationConfig{FieldA: "test1", FieldB: 42}
	c1 := NewContext(ctx, value1)
	c1.WithMetadata("trace_id", "12345").WithPermissions("read", "write")

	// Create second context
	value2 := TestValidationConfig{FieldA: "test2", FieldB: 43}
	c2 := NewContext(ctx, value2)
	c2.WithMetadata("span_id", "67890").WithPermissions("execute")

	// Merge
	merged := c1.Merge(c2)

	if merged == c1 || merged == c2 {
		t.Error("Merge should create a new instance")
	}

	// Check metadata merge (c2 should have both sets)
	if len(merged.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries after merge, got %d", len(merged.Metadata))
	}

	if val, ok := merged.Metadata["trace_id"]; !ok || val != "12345" {
		t.Error("Merged context should have trace_id from c1")
	}

	if val, ok := merged.Metadata["span_id"]; !ok || val != "67890" {
		t.Error("Merged context should have span_id from c2")
	}

	// Check permissions merge (c2 should have all permissions)
	if len(merged.Permissions) != 3 {
		t.Errorf("Expected 3 permissions after merge, got %d", len(merged.Permissions))
	}

	if !merged.HasPermission("read") || !merged.HasPermission("write") ||
		!merged.HasPermission("execute") {
		t.Error("Merged context should have all permissions from both contexts")
	}

	// Verify originals are not modified
	if len(c1.Metadata) != 1 || len(c2.Metadata) != 1 {
		t.Error("Original contexts should not be modified by merge")
	}
}

func TestSetValueType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value1 := TestValidationConfig{FieldA: "test1", FieldB: 42}
	c := NewContext(ctx, value1)

	if c.GetValueType().FieldA != "test1" {
		t.Error("Expected original value in GetValueType()")
	}

	value2 := TestValidationConfig{FieldA: "test2", FieldB: 43}
	c.SetValueType(value2)

	if c.GetValueType().FieldA != "test2" {
		t.Error("Expected updated value in GetValueType()")
	}
}

func TestGetContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	if c.GetContext() != ctx {
		t.Error("Expected GetContext() to return the set context")
	}

	// Test with nil context
	c.Context = nil
	if c.GetContext() != context.Background() {
		t.Error("Expected GetContext() to return context.Background() when context is nil")
	}
}

func TestDifferentTypes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create different types of contexts
	validationCtx := NewContext(ctx, TestValidationConfig{FieldA: "validation"})
	errorCtx := NewContext(ctx, TestErrorConfig{Code: 500, Message: "error"})
	sanitizationCtx := NewContext(
		ctx,
		TestSanitizationConfig{Policy: "strict", Rules: []string{"rule1"}},
	)

	// Verify each holds its type correctly
	if validationCtx.GetValueType().FieldA != "validation" {
		t.Error("Validation context should hold validation config")
	}

	if errorCtx.GetValueType().Code != 500 {
		t.Error("Error context should hold error config")
	}

	if sanitizationCtx.GetValueType().Policy != "strict" {
		t.Error("Sanitization context should hold sanitization config")
	}
}

func TestMetadataOperations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	// Add metadata
	c.WithMetadata("key1", "value1")
	c.WithMetadata("key2", "value2")

	if len(c.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(c.Metadata))
	}

	// Test GetMetadata
	val, ok := c.GetMetadata("key1")
	if !ok || val != "value1" {
		t.Error("Failed to get metadata key1")
	}

	// Test non-existent key
	_, ok = c.GetMetadata("key3")
	if ok {
		t.Error("Should not get non-existent metadata key")
	}

	// Test WithMetadata with empty string
	c.WithMetadata("key4", "")

	val, ok = c.GetMetadata("key4")
	if !ok || val != "" {
		t.Error("Should allow empty string values")
	}
}

func TestPermissionsEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	// Test with empty permissions
	c.WithPermissions()

	if len(c.Permissions) != 0 {
		t.Error("Expected 0 permissions")
	}

	// Test HasPermission with empty permissions
	if c.HasPermission("anything") {
		t.Error("Should not have permissions when none are set")
	}

	// Test with duplicate permissions
	c.WithPermissions("read", "read", "read")

	if len(c.Permissions) != 3 {
		t.Error("Should allow duplicate permissions")
	}

	// Test HasPermission with duplicate
	if !c.HasPermission("read") {
		t.Error("Should have read permission")
	}

	// Test with empty string permission
	c.WithPermissions("")

	if !c.HasPermission("") {
		t.Error("Should allow empty string permissions")
	}
}

// Branching method tests

func TestBranch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	// Test branch with true condition
	result := c.Branch(true, func(c *Context[TestValidationConfig]) *Context[TestValidationConfig] {
		return c.WithMetadata("branched", "true")
	})

	if result == c {
		t.Error("Branch should return a new context when condition is true")
	}

	if val, ok := result.Metadata["branched"]; !ok || val != "true" {
		t.Error("Branch should add metadata when condition is true")
	}

	// Test branch with false condition
	c2 := NewContext(ctx, TestValidationConfig{FieldA: "test2", FieldB: 43})
	result2 := c2.Branch(false, func(c *Context[TestValidationConfig]) *Context[TestValidationConfig] {
		return c.WithMetadata("branched", "true")
	})

	if result2 != c2 {
		t.Error("Branch should return original context when condition is false")
	}
}

func TestBranchWithValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	// Test when predicate is true
	result := c.BranchWithValue(
		func(v TestValidationConfig) bool { return v.FieldB > 10 },
		func(c *Context[TestValidationConfig]) *Context[TestValidationConfig] {
			return c.WithMetadata("value_branched", "true")
		},
	)

	if result == c {
		t.Error("BranchWithValue should return new context when predicate is true")
	}

	// Test when predicate is false
	c2 := NewContext(ctx, TestValidationConfig{FieldA: "test2", FieldB: 5})
	result2 := c2.BranchWithValue(
		func(v TestValidationConfig) bool { return v.FieldB > 10 },
		func(c *Context[TestValidationConfig]) *Context[TestValidationConfig] {
			return c.WithMetadata("value_branched", "true")
		},
	)

	if result2 == c2 {
		t.Error("BranchWithValue should return cloned context when predicate is false")
	}
}

func TestFork(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := TestValidationConfig{FieldA: "test", FieldB: 42}
	c := NewContext(ctx, value)

	branches := map[bool]func(*Context[TestValidationConfig]) *Context[TestValidationConfig]{
		true: func(c *Context[TestValidationConfig]) *Context[TestValidationConfig] {
			return c.WithMetadata("fork1", "value1")
		},
		false: func(c *Context[TestValidationConfig]) *Context[TestValidationConfig] {
			return c.WithMetadata("fork2", "value2")
		},
	}

	results := c.Fork(branches)

	if len(results) != 1 {
		t.Errorf("Fork should return 1 result when only one condition is true, got %d", len(results))
	}

	if _, ok := results[0].Metadata["fork1"]; !ok {
		t.Error("Fork should apply true branch transform")
	}
}

func TestJoin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c1 := NewContext(ctx, TestValidationConfig{FieldA: "test1", FieldB: 42})
	c1.WithMetadata("key1", "value1")

	c2 := NewContext(ctx, TestValidationConfig{FieldA: "test2", FieldB: 43})
	c2.WithMetadata("key2", "value2")

	merged := Join(c1, c2)

	if len(merged.Metadata) != 2 {
		t.Errorf("Join should merge metadata, got %d", len(merged.Metadata))
	}

	if merged.Metadata["key1"] != "value1" || merged.Metadata["key2"] != "value2" {
		t.Error("Join should have metadata from both contexts")
	}

	// Test empty join
	empty := Join[*Context[TestValidationConfig]]()

	if empty == nil {
		t.Error("Join should return non-nil for empty input")
	}
}

func TestJoinWithValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c1 := NewContext(ctx, 10)
	c1.WithMetadata("key1", "value1")

	c2 := NewContext(ctx, 20)
	c2.WithMetadata("key2", "value2")

	merged := JoinWithValue([]*Context[int]{c1, c2}, func(values []int) int {
		sum := 0
		for _, v := range values {
			sum += v
		}
		return sum
	})

	if merged.ValueType != 30 {
		t.Errorf("JoinWithValue should sum values, got %d", merged.ValueType)
	}

	if len(merged.Metadata) != 2 {
		t.Error("JoinWithValue should merge metadata")
	}
}

func TestTransform(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c := NewContext(ctx, TestValidationConfig{FieldA: "test", FieldB: 42})
	c.WithMetadata("key", "value")

	transformed := Transform(c, func(v TestValidationConfig) string {
		return v.FieldA
	})

	if transformed.ValueType != "test" {
		t.Errorf("Transform should transform value type, got %s", transformed.ValueType)
	}

	if transformed.Metadata["key"] != "value" {
		t.Error("Transform should preserve metadata")
	}
}

func TestPick(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c1 := NewContext(ctx, TestValidationConfig{FieldA: "test1", FieldB: 42})
	c2 := NewContext(ctx, TestValidationConfig{FieldA: "test2", FieldB: 43})
	c3 := NewContext(ctx, TestValidationConfig{FieldA: "test3", FieldB: 44})

	contexts := []*Context[TestValidationConfig]{c1, c2, c3}

	// Pick by value
	picked := Pick(contexts, func(c *Context[TestValidationConfig]) bool {
		return c.ValueType.FieldB == 43
	})

	if picked == nil {
		t.Error("Pick should return context when predicate matches")
	}

	if picked.ValueType.FieldA != "test2" {
		t.Error("Pick should return matching context")
	}

	// Pick when no match
	pickedNone := Pick(contexts, func(c *Context[TestValidationConfig]) bool {
		return c.ValueType.FieldB == 999
	})

	if pickedNone != nil {
		t.Error("Pick should return nil when no match")
	}
}

func TestPickWithValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c1 := NewContext(ctx, TestValidationConfig{FieldA: "test1", FieldB: 42})
	c2 := NewContext(ctx, TestValidationConfig{FieldA: "test2", FieldB: 43})
	c3 := NewContext(ctx, TestValidationConfig{FieldA: "test3", FieldB: 44})

	contexts := []*Context[TestValidationConfig]{c1, c2, c3}

	// Pick by value
	picked := PickWithValue(contexts, func(v TestValidationConfig) bool {
		return v.FieldB == 44
	})

	if picked == nil {
		t.Error("PickWithValue should return context when predicate matches")
	}

	if picked.ValueType.FieldA != "test3" {
		t.Error("PickWithValue should return matching context")
	}
}
