package prompter

import (
	"testing"
)

// runSecretWithBytes exercises Secret logic directly.
func runSecretWithBytes(s *Secret, pass1, pass2 []byte) (*Result, error) {
	// Simulate the validation logic from Run()
	if s.opts.Required && len(pass1) == 0 {
		return nil, ErrValidation{Msg: "input is required"}
	}
	if s.opts.MinLen > 0 && len(pass1) < s.opts.MinLen {
		return nil, ErrValidation{Msg: "too short"}
	}
	if s.opts.MaxLen > 0 && len(pass1) > s.opts.MaxLen {
		return nil, ErrValidation{Msg: "too long"}
	}
	if s.opts.Validator != nil {
		if err := s.opts.Validator(pass1); err != nil {
			return nil, err
		}
	}

	if !s.confirm {
		return NewResult(pass1), nil
	}

	// Check confirmation
	if !bytesEqual(pass1, pass2) {
		return nil, ErrMismatch
	}
	return NewResult(pass1), nil
}

func TestSecret_Required_Empty(t *testing.T) {
	s := NewSecret("pass", WithRequired(true))
	_, err := runSecretWithBytes(s, []byte{}, nil)
	if err == nil {
		t.Fatal("expected required error")
	}
}

func TestSecret_Required_NonEmpty(t *testing.T) {
	s := NewSecret("pass", WithRequired(true))
	r, err := runSecretWithBytes(s, []byte("secret"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.String() != "secret" {
		t.Fatalf("wrong value: %s", r.String())
	}
	r.Zero()
}

func TestSecret_MinLength_TooShort(t *testing.T) {
	s := NewSecret("pass", WithLength(8, 0))
	_, err := runSecretWithBytes(s, []byte("short"), nil)
	if err == nil {
		t.Fatal("expected minLength error")
	}
}

func TestSecret_MinLength_Exact(t *testing.T) {
	s := NewSecret("pass", WithLength(4, 10))
	r, err := runSecretWithBytes(s, []byte("pass"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r.Zero()
}

func TestSecret_MaxLength_TooLong(t *testing.T) {
	s := NewSecret("pass", WithLength(0, 4))
	_, err := runSecretWithBytes(s, []byte("too long"), nil)
	if err == nil {
		t.Fatal("expected maxLength error")
	}
}

func TestSecret_Confirm_Match(t *testing.T) {
	s := NewSecret("pass").WithConfirmation("confirm")
	r, err := runSecretWithBytes(s, []byte("correct"), []byte("correct"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.String() != "correct" {
		t.Fatalf("wrong value: %s", r.String())
	}
	r.Zero()
}

func TestSecret_Confirm_Mismatch(t *testing.T) {
	s := NewSecret("pass").WithConfirmation("confirm")
	_, err := runSecretWithBytes(s, []byte("pass1"), []byte("pass2"))
	if err != ErrMismatch {
		t.Fatalf("expected mismatch, got: %v", err)
	}
}

func TestSecret_NoConfirm_IgnoresPass2(t *testing.T) {
	s := NewSecret("pass") // confirm=false
	r, err := runSecretWithBytes(s, []byte("abc"), []byte("xyz"))
	if err != nil {
		t.Fatal("should not error when confirm disabled")
	}
	if r.String() != "abc" {
		t.Fatal("wrong value")
	}
}

func TestSecret_Validator(t *testing.T) {
	s := NewSecret("pass", WithValidator(func(b []byte) error {
		if string(b) != "magic" {
			return ErrValidation{Msg: "not magic"}
		}
		return nil
	}))

	_, err := runSecretWithBytes(s, []byte("wrong"), nil)
	if err == nil || err.Error() != "not magic" {
		t.Fatalf("expected validator error, got: %v", err)
	}

	r, err := runSecretWithBytes(s, []byte("magic"), nil)
	if err != nil {
		t.Fatal("should pass with magic")
	}
	r.Zero()
}

func TestSecret_RequiredAndLength(t *testing.T) {
	s := NewSecret("pass", WithRequired(true), WithLength(8, 0))

	// Empty
	_, err := runSecretWithBytes(s, []byte{}, nil)
	if err == nil {
		t.Fatal("expected required error")
	}

	// Too short
	_, err = runSecretWithBytes(s, []byte("short"), nil)
	if err == nil {
		t.Fatal("expected length error")
	}

	// OK
	r, err := runSecretWithBytes(s, []byte("longenough"), nil)
	if err != nil {
		t.Fatal("should pass")
	}
	r.Zero()
}
