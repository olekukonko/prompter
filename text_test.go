package prompter

import (
	"testing"
)

// runTextWithBytes exercises Input logic.
func runTextWithBytes(ti *Input, input []byte) (*Result, error) {
	// Simulate validation from Run()
	if ti.opts.Required && len(input) == 0 {
		return nil, ErrValidation{Msg: "input is required"}
	}
	if ti.opts.MinLen > 0 && len(input) < ti.opts.MinLen {
		return nil, ErrValidation{Msg: "too short"}
	}
	if ti.opts.MaxLen > 0 && len(input) > ti.opts.MaxLen {
		return nil, ErrValidation{Msg: "too long"}
	}
	if ti.opts.Validator != nil {
		if err := ti.opts.Validator(input); err != nil {
			return nil, err
		}
	}
	return NewResult(input), nil
}

func TestText_Required(t *testing.T) {
	ti := NewTextInput("name", WithRequired(true))

	_, err := runTextWithBytes(ti, []byte{})
	if err == nil {
		t.Fatal("expected required error")
	}

	r, err := runTextWithBytes(ti, []byte("john"))
	if err != nil {
		t.Fatal(err)
	}
	if r.String() != "john" {
		t.Fatal("wrong value")
	}
}

func TestText_Length(t *testing.T) {
	ti := NewTextInput("code", WithLength(3, 5))

	// Too short
	_, err := runTextWithBytes(ti, []byte("ab"))
	if err == nil {
		t.Fatal("expected error")
	}

	// Too long
	_, err = runTextWithBytes(ti, []byte("abcdef"))
	if err == nil {
		t.Fatal("expected error")
	}

	// Just right
	r, err := runTextWithBytes(ti, []byte("abcd"))
	if err != nil {
		t.Fatal(err)
	}
	r.Zero()
}

func TestText_NoValidation_AllowsEmpty(t *testing.T) {
	ti := NewTextInput("optional")
	r, err := runTextWithBytes(ti, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Len() != 0 {
		t.Fatal("should be empty")
	}
}
