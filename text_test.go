package prompter

import (
	"bytes"
	"context"
	"testing"
)

func TestText_Required(t *testing.T) {
	input := bytes.NewReader([]byte("\njohn\n"))
	ti := NewTextInput("name", WithRequired(true), WithInput(input))
	r, err := ti.Run()
	if err != nil {
		t.Fatal(err)
	}
	if r.String() != "john" {
		t.Fatal("wrong value")
	}
}

func TestText_Length(t *testing.T) {
	input := bytes.NewReader([]byte("ab\nabcdef\nabcd\n"))
	ti := NewTextInput("code", WithLength(3, 5), WithInput(input))
	r, err := ti.Run()
	if err != nil {
		t.Fatal(err)
	}
	if r.String() != "abcd" {
		t.Fatal("wrong value")
	}
	r.Zero()
}

func TestText_NoValidation_AllowsEmpty(t *testing.T) {
	input := bytes.NewReader([]byte("\n"))
	ti := NewTextInput("optional", WithInput(input))
	r, err := ti.Run()
	if err != nil {
		t.Fatal(err)
	}
	if r.Len() != 0 {
		t.Fatal("should be empty")
	}
}

func TestText_ContextCancellation(t *testing.T) {
	ti := NewTextInput("test")
	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	_, err := ti.RunContext(ctx)
	if err != context.Canceled {
		t.Fatalf("expected canceled, got: %v", err)
	}
}
