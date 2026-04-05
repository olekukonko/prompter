package prompter

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestSecret_Required_Empty(t *testing.T) {
	input := bytes.NewReader([]byte{})
	s := NewSecret("pass", WithRequired(true), WithInput(input))
	_, err := s.Run()
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got: %v", err)
	}
}

func TestSecret_Required_NonEmpty(t *testing.T) {
	input := bytes.NewReader([]byte("secret\n"))
	s := NewSecret("pass", WithRequired(true), WithInput(input))
	r, err := s.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.String() != "secret" {
		t.Fatalf("wrong value: %s", r.String())
	}
	r.Zero()
}

func TestSecret_MinLength_TooShort(t *testing.T) {
	input := bytes.NewReader([]byte("short\nalso\n"))
	s := NewSecret("pass", WithLength(8, 0), WithInput(input))
	_, err := s.Run()
	if err == nil || !strings.Contains(err.Error(), "minimum length") {
		t.Fatalf("expected minLength error, got: %v", err)
	}
}

func TestSecret_MaxLength_TooLong(t *testing.T) {
	input := bytes.NewReader([]byte("this is way too long\nanother long one\n"))
	s := NewSecret("pass", WithLength(0, 5), WithInput(input))
	_, err := s.Run()
	if err == nil || !strings.Contains(err.Error(), "maximum length") {
		t.Fatalf("expected maxLength error, got: %v", err)
	}
}

func TestSecret_Confirm_Match(t *testing.T) {
	input := bytes.NewReader([]byte("correct\ncorrect\n"))
	s := NewSecret("pass", WithInput(input)).WithConfirmation("confirm")
	r, err := s.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.String() != "correct" {
		t.Fatalf("wrong value: %s", r.String())
	}
	r.Zero()
}

func TestSecret_Confirm_Mismatch(t *testing.T) {
	input := bytes.NewReader([]byte("pass1\npass2\n"))
	s := NewSecret("pass", WithInput(input)).WithConfirmation("confirm")
	_, err := s.Run()
	if err != ErrMismatch && !strings.Contains(err.Error(), "do not match") {
		t.Fatalf("expected mismatch, got: %v", err)
	}
}

func TestSecret_ContextCancellation(t *testing.T) {
	s := NewSecret("pass")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond)
	_, err := s.RunContext(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected timeout, got: %v", err)
	}
}

func TestSecret_Validator(t *testing.T) {
	input := bytes.NewReader([]byte("wrong\nmagic\n"))
	s := NewSecret("pass", WithInput(input), WithValidator(func(b []byte) error {
		if string(b) != "magic" {
			return ErrValidation{Msg: "not magic"}
		}
		return nil
	}))

	r, err := s.Run()
	if err != nil {
		t.Fatal(err)
	}
	if r.String() != "magic" {
		t.Fatal("wrong value")
	}
	r.Zero()
}
