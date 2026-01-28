package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

// TestIssue519 tests the fix for GitHub issue #519
// Issue: Panic on marshaling nested single-field structs with pointer and embedded struct
//
// Problem: When marshaling a struct with:
//   - A pointer field to another struct
//   - That struct has a single embedded (non-pointer) struct field
//   - The embedded struct has string fields
//
// The encoder would panic with "unexpected fault address" or segmentation violation
// because it was incorrectly applying an extra pointer dereference (PtrNum=1) to
// the embedded struct's opcodes, causing string data to be interpreted as memory addresses.
//
// Fix: Modified PtrCode.ToOpcode() to only skip PtrNum for StructHead->StructPtrHead
// conversions without IndirectFlags, and fixed isIndirect propagation in compiler.
//
// Related: https://github.com/goccy/go-json/issues/519
func TestIssue519(t *testing.T) {
	// Test case 1: Exact reproduction from issue report
	t.Run("original_issue_single_field", func(t *testing.T) {
		type Item struct {
			A string `json:"a"`
		}

		type Detail struct {
			I Item `json:"i"` // Embedded struct (not pointer)
		}

		type Body struct {
			Payload *Detail `json:"p,omitempty"` // Pointer to struct
		}

		body := Body{
			Payload: &Detail{
				I: Item{
					A: "test",
				},
			},
		}

		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{"p":{"i":{"a":"test"}}}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})

	// Test case 2: User-reported variation with multiple fields in Item
	t.Run("original_issue_multiple_fields", func(t *testing.T) {
		type Item struct {
			A string `json:"a"`
			B string `json:"b,omitempty"`
		}

		type Detail struct {
			I Item `json:"i"`
		}

		type Body struct {
			Payload *Detail `json:"p,omitempty"`
		}

		body := Body{
			Payload: &Detail{
				I: Item{
					A: "a",
					B: "b",
				},
			},
		}

		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{"p":{"i":{"a":"a","b":"b"}}}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})

	// Test case 3: Empty optional field
	t.Run("empty_optional_field", func(t *testing.T) {
		type Item struct {
			A string `json:"a"`
			B string `json:"b,omitempty"`
		}

		type Detail struct {
			I Item `json:"i"`
		}

		type Body struct {
			Payload *Detail `json:"p,omitempty"`
		}

		body := Body{
			Payload: &Detail{
				I: Item{
					A: "a",
					B: "", // Empty, should be omitted
				},
			},
		}

		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{"p":{"i":{"a":"a"}}}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})

	// Test case 4: Nil pointer (should omit)
	t.Run("nil_pointer_omitempty", func(t *testing.T) {
		type Item struct {
			A string `json:"a"`
		}

		type Detail struct {
			I Item `json:"i"`
		}

		type Body struct {
			Payload *Detail `json:"p,omitempty"`
		}

		body := Body{
			Payload: nil,
		}

		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})

	// Test case 5: Deeper nesting with same pattern
	t.Run("deeper_nesting", func(t *testing.T) {
		type Level3 struct {
			Value string `json:"value"`
		}

		type Level2 struct {
			Inner Level3 `json:"inner"`
		}

		type Level1 struct {
			Middle Level2 `json:"middle"`
		}

		type Root struct {
			Ptr *Level1 `json:"ptr,omitempty"`
		}

		root := Root{
			Ptr: &Level1{
				Middle: Level2{
					Inner: Level3{
						Value: "deep",
					},
				},
			},
		}

		b, err := json.Marshal(root)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{"ptr":{"middle":{"inner":{"value":"deep"}}}}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})

	// Test case 6: Special characters that might have caused memory interpretation issues
	t.Run("special_characters", func(t *testing.T) {
		type Item struct {
			A string `json:"a"`
		}

		type Detail struct {
			I Item `json:"i"`
		}

		type Body struct {
			Payload *Detail `json:"p,omitempty"`
		}

		body := Body{
			Payload: &Detail{
				I: Item{
					A: "test-0.1", // This pattern showed up in crash reports
				},
			},
		}

		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{"p":{"i":{"a":"test-0.1"}}}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})

	// Test case 7: Multiple nested embedded structs
	t.Run("multiple_embedded_structs", func(t *testing.T) {
		type Inner1 struct {
			Value1 string `json:"value1"`
		}

		type Inner2 struct {
			Value2 string `json:"value2"`
		}

		type Container struct {
			I1 Inner1 `json:"i1"`
			I2 Inner2 `json:"i2"`
		}

		type Root struct {
			Ptr *Container `json:"ptr,omitempty"`
		}

		root := Root{
			Ptr: &Container{
				I1: Inner1{Value1: "first"},
				I2: Inner2{Value2: "second"},
			},
		}

		b, err := json.Marshal(root)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{"ptr":{"i1":{"value1":"first"},"i2":{"value2":"second"}}}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})

	// Test case 8: Verify fix doesn't break pointer-to-pointer fields
	t.Run("pointer_to_pointer_fields", func(t *testing.T) {
		type Item struct {
			A *string `json:"a,omitempty"`
		}

		type Detail struct {
			I Item `json:"i"`
		}

		type Body struct {
			Payload *Detail `json:"p,omitempty"`
		}

		str := "pointer_value"
		body := Body{
			Payload: &Detail{
				I: Item{
					A: &str,
				},
			},
		}

		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		expected := `{"p":{"i":{"a":"pointer_value"}}}`
		if string(b) != expected {
			t.Errorf("Expected %s, got %s", expected, string(b))
		}
	})
}
