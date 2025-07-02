package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

// This test file contains the minimal reproduction case for the reported panic.
// It can be used to verify that the issue has been fixed.

func TestMinimalReproduction(t *testing.T) {
	// Exact reproduction from the issue report
	type testPanicBody struct {
		Payload *testPanicDetail `json:"p,omitempty"`
	}

	type testDetail struct {
		I testPanicItem `json:"i"`
	}

	type testItem struct {
		A string `json:"a"`
		B string `json:"b,omitempty"`
	}

	body := testPanicBody{
		Payload: &testPanicDetail{
			I: testPanicItem{
				A: "a",
				B: "b",
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

	expected := `{"p":{"i":{"a":"a","b":"b"}}}`
	if string(b) != expected {
		t.Errorf("Expected %s, got %s", expected, string(b))
	}

	t.Logf("Successfully marshaled: %s", string(b))
}
