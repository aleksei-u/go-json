package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

// Test case for the panic reported in GitHub issue
// This test reproduces the segmentation fault that occurs when marshaling
// nested structs with specific field configurations.

type testBody struct {
	Payload *testPanicDetail `json:"p,omitempty"`
}

type testPanicDetail struct {
	I testPanicItem `json:"i"`
}

type testPanicItem struct {
	A string `json:"a"`
	B string `json:"b,omitempty"`
}

func TestMarshalNestedStructPanic(t *testing.T) {
	// This is the exact case that causes the panic
	t.Run("basic_nested_struct_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with: %v", r)
			}
		}()

		body := testBody{
			Payload: &testPanicDetail{
				I: testPanicItem{
					A: "a",
					B: "b",
				},
			},
		}

		b, err := json.Marshal(body)
		assertErr(t, err)

		expected := `{"p":{"i":{"a":"a","b":"b"}}}`
		assertEq(t, "nested struct marshal", expected, string(b))
	})

	// Test with empty optional field
	t.Run("empty_optional_field", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with: %v", r)
			}
		}()

		body := testBody{
			Payload: &testPanicDetail{
				I: testPanicItem{
					A: "a",
					B: "", // Empty optional field
				},
			},
		}

		b, err := json.Marshal(body)
		assertErr(t, err)

		expected := `{"p":{"i":{"a":"a"}}}`
		assertEq(t, "empty optional field", expected, string(b))
	})

	// Test with nil payload
	t.Run("nil_payload", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with: %v", r)
			}
		}()

		body := testBody{
			Payload: nil,
		}

		b, err := json.Marshal(body)
		assertErr(t, err)

		expected := `{}`
		assertEq(t, "nil payload", expected, string(b))
	})

	// Test with special characters in strings
	t.Run("special_characters", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with: %v", r)
			}
		}()

		body := testBody{
			Payload: &testPanicDetail{
				I: testPanicItem{
					A: "test<>&\"'",
					B: "unicode: \u2603 ❄",
				},
			},
		}

		b, err := json.Marshal(body)
		assertErr(t, err)

		// The exact expected output depends on HTML escaping behavior
		// but the important thing is that it doesn't panic
		if len(b) == 0 {
			t.Error("Expected non-empty JSON output")
		}
	})

	// Test with very long strings that might trigger buffer issues
	t.Run("long_strings", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with: %v", r)
			}
		}()

		longString := string(make([]byte, 10000))
		for i := range longString {
			longString = longString[:i] + "x" + longString[i+1:]
		}

		body := testBody{
			Payload: &testPanicDetail{
				I: testPanicItem{
					A: longString,
					B: "short",
				},
			},
		}

		b, err := json.Marshal(body)
		assertErr(t, err)

		if len(b) == 0 {
			t.Error("Expected non-empty JSON output for long strings")
		}
	})
}

// Additional nested struct variations to test edge cases
type DeepNesting struct {
	Level1 *Level1 `json:"l1,omitempty"`
}

type Level1 struct {
	Level2 *Level2 `json:"l2,omitempty"`
	Value  string  `json:"value"`
}

type Level2 struct {
	Level3 *Level3 `json:"l3,omitempty"`
	Value  string  `json:"value"`
}

type Level3 struct {
	Value string `json:"value"`
	Data  string `json:"data,omitempty"`
}

func TestDeepNestedStructs(t *testing.T) {
	t.Run("deep_nesting", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with deep nesting: %v", r)
			}
		}()

		deep := DeepNesting{
			Level1: &Level1{
				Level2: &Level2{
					Level3: &Level3{
						Value: "deep",
						Data:  "nested",
					},
					Value: "level2",
				},
				Value: "level1",
			},
		}

		b, err := json.Marshal(deep)
		assertErr(t, err)

		expected := `{"l1":{"l2":{"l3":{"value":"deep","data":"nested"},"value":"level2"},"value":"level1"}}`
		assertEq(t, "deep nesting", expected, string(b))
	})

	t.Run("partial_deep_nesting", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with partial deep nesting: %v", r)
			}
		}()

		deep := DeepNesting{
			Level1: &Level1{
				Level2: &Level2{
					Level3: nil, // Missing deep level
					Value:  "level2",
				},
				Value: "level1",
			},
		}

		b, err := json.Marshal(deep)
		assertErr(t, err)

		expected := `{"l1":{"l2":{"value":"level2"},"value":"level1"}}`
		assertEq(t, "partial deep nesting", expected, string(b))
	})
}

// Test struct with various field types that might cause issues
type ComplexStruct struct {
	StringPtr    *string        `json:"string_ptr,omitempty"`
	IntPtr       *int           `json:"int_ptr,omitempty"`
	BoolPtr      *bool          `json:"bool_ptr,omitempty"`
	NestedStruct *testPanicItem `json:"nested,omitempty"`
	SlicePtr     *[]string      `json:"slice_ptr,omitempty"`
}

func TestComplexStructMarshal(t *testing.T) {
	t.Run("complex_struct_all_fields", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with complex struct: %v", r)
			}
		}()

		str := "test"
		num := 42
		flag := true
		slice := []string{"a", "b", "c"}

		complex := ComplexStruct{
			StringPtr:    &str,
			IntPtr:       &num,
			BoolPtr:      &flag,
			NestedStruct: &testPanicItem{A: "nested_a", B: "nested_b"},
			SlicePtr:     &slice,
		}

		b, err := json.Marshal(complex)
		assertErr(t, err)

		if len(b) == 0 {
			t.Error("Expected non-empty JSON output for complex struct")
		}
	})

	t.Run("complex_struct_nil_fields", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Marshal panicked with nil fields: %v", r)
			}
		}()

		complex := ComplexStruct{
			StringPtr:    nil,
			IntPtr:       nil,
			BoolPtr:      nil,
			NestedStruct: nil,
			SlicePtr:     nil,
		}

		b, err := json.Marshal(complex)
		assertErr(t, err)

		expected := `{}`
		assertEq(t, "all nil fields", expected, string(b))
	})
}
