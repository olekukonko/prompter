package prompter

import (
	"bytes"
	"testing"
)

func TestResult_BytesAndString(t *testing.T) {
	data := []byte("secret")
	r := NewResult(data)

	if !bytes.Equal(r.Bytes(), data) {
		t.Error("bytes don't match")
	}
	if r.String() != "secret" {
		t.Error("string doesn't match")
	}
	if r.Len() != 6 {
		t.Error("wrong length")
	}
}

func TestResult_Zero(t *testing.T) {
	data := []byte("secret")
	r := NewResult(data)
	r.Zero()

	if r.Bytes() != nil {
		t.Error("should be nil after zero")
	}
	if r.Len() != 0 {
		t.Error("length should be 0")
	}
	// Safe to call twice
	r.Zero()
}

func TestResult_NilInput(t *testing.T) {
	r := NewResult(nil)
	if r.Bytes() != nil {
		t.Error("should handle nil")
	}
	if r.String() != "" {
		t.Error("string should be empty")
	}
	r.Zero() // Should not panic
}

func TestResult_CopyOnCreate(t *testing.T) {
	original := []byte("test")
	r := NewResult(original)

	// Modify original
	original[0] = 'X'

	// Result should be unchanged
	if r.String() == "Xest" {
		t.Error("Result should copy bytes, not reference them")
	}
}
