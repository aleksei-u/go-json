package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

// This test file contains the minimal reproduction case for the reported panic.
// It can be used to verify that the issue has been fixed.

func TestMinimalReproduction(t *testing.T) {
	// Exact reproduction from the issue report

	type testPanic519Item struct {
		A string `json:"a"`
	}

	type testPanic519Detail struct {
		I testPanic519Item `json:"i"`
	}

	// The issue presented only when the root structure has a pointer field on another structure and there is only one field
	// and the inner structure has a field with to another structure as a first field
	// and internal structure has a string field
	type testPanic519Body struct {
		Payload *testPanic519Detail `json:"p,omitempty"`
	}

	body := testPanic519Body{
		Payload: &testPanic519Detail{
			I: testPanic519Item{
				A: "a_field",
			},
		},
	}

	// The test should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Expected no panic, but got: %v", r)
		}
	}()

	var b []byte
	var err error

	b, err = json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	expected := `{"p":{"i":{"a":"a_field"}}}`
	if string(b) != expected {
		t.Errorf("Expected %s, got %s", expected, string(b))
	}

	t.Logf("Successfully marshaled: %s", string(b))
}
