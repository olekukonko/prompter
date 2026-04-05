package prompter

import (
	"testing"
)

func TestOptions_Chaining(t *testing.T) {
	// Test that all options can be chained and values stick
	formatterCalled := false
	validatorCalled := false
	errCallbackCalled := false

	opts := Options{}

	WithRequired(true)(&opts)
	WithLength(8, 128)(&opts)
	WithMaxRetries(3)(&opts)
	WithValidator(func(b []byte) error {
		validatorCalled = true
		return nil
	})(&opts)
	WithFormatter(func(ctx Context) string {
		formatterCalled = true
		return "test"
	})(&opts)
	WithErrorCallback(func(e error) {
		errCallbackCalled = true
	})(&opts)

	if !opts.Required {
		t.Error("Required not set")
	}
	if opts.MinLen != 8 || opts.MaxLen != 128 {
		t.Error("Length not set")
	}
	if opts.MaxRetries != 3 {
		t.Error("MaxRetries not set")
	}
	if opts.Validator == nil {
		t.Error("Validator not set")
	}

	// Test callbacks work
	opts.Validator([]byte("test"))
	if !validatorCalled {
		t.Error("Validator not called")
	}

	opts.Formatter(Context{})
	if !formatterCalled {
		t.Error("Formatter not called")
	}

	opts.ErrorCallback(ErrValidation{Msg: "test"})
	if !errCallbackCalled {
		t.Error("ErrorCallback not called")
	}
}

func TestOptions_Defaults(t *testing.T) {
	s := NewSecret("test")
	if s.opts.Formatter == nil {
		t.Error("default formatter not set")
	}
	if s.opts.Required {
		t.Error("should not be required by default")
	}

	ti := NewTextInput("test")
	if ti.opts.Formatter == nil {
		t.Error("default formatter not set")
	}
}
