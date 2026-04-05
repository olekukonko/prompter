package prompter

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

// Secret reads passwords/secrets (no echo) with optional confirmation.
type Secret struct {
	opts       Options
	prompt     string
	confirm    bool
	confirmMsg string
	onEmpty    EmptyHandler
}

// NewSecret creates a secret prompter with shared options.
func NewSecret(prompt string, opts ...Option) *Secret {
	s := &Secret{
		prompt: prompt,
		opts: Options{
			Formatter: func(ctx Context) string {
				suffix := ""
				if ctx.IsConfirm {
					suffix = " (confirm)"
				} else {
					suffix = " (hidden)"
				}
				if ctx.IsRetry && ctx.LastError != nil {
					return fmt.Sprintf("[attempt %d] %s%s: ", ctx.Attempt, ctx.Prompt, suffix)
				}
				return fmt.Sprintf("%s%s: ", ctx.Prompt, suffix)
			},
		},
		confirmMsg: "Confirm",
	}
	for _, opt := range opts {
		opt(&s.opts)
	}
	return s
}

// WithConfirmation enables confirmation prompt.
func (s *Secret) WithConfirmation(msg string) *Secret {
	s.confirm = true
	if msg != "" {
		s.confirmMsg = msg
	}
	return s
}

// WithOnEmpty sets the handler for empty input (called before confirmation).
func (s *Secret) WithOnEmpty(h EmptyHandler) *Secret {
	s.onEmpty = h
	return s
}

func (s *Secret) readValue(ctx Context) ([]byte, error) {
	fmt.Fprintf(os.Stderr, s.opts.Formatter(ctx), ctx.Prompt)

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, errors.New("stdin is not a terminal")
	}

	val, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	return val, err
}

func (s *Secret) Run() (*Result, error) {
	for attempt := 1; ; attempt++ {
		// First value (password)
		ctx := Context{
			Prompt:     s.prompt,
			Attempt:    attempt,
			MaxRetries: s.opts.MaxRetries,
			IsRetry:    attempt > 1 && s.opts.LastError != nil,
			LastError:  s.opts.LastError,
		}

		val1, err := s.readValue(ctx)
		if err != nil {
			return nil, err
		}

		// Handle empty before validation/confirmation
		if len(val1) == 0 && s.onEmpty != nil {
			if !s.onEmpty() {
				if s.opts.MaxRetries > 0 && attempt >= s.opts.MaxRetries {
					return nil, fmt.Errorf("%w: empty input refused", ErrMaxRetries)
				}
				// Mark error for retry formatting
				s.opts.LastError = ErrValidation{Msg: "empty input"}
				continue
			}
		}

		// Validate BEFORE confirmation
		var valErr error
		switch {
		case s.opts.Required && len(val1) == 0:
			valErr = ErrValidation{Msg: "input is required"}
		case s.opts.MinLen > 0 && len(val1) < s.opts.MinLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("minimum length is %d", s.opts.MinLen)}
		case s.opts.MaxLen > 0 && len(val1) > s.opts.MaxLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("maximum length is %d", s.opts.MaxLen)}
		case s.opts.Validator != nil:
			valErr = s.opts.Validator(val1)
		}

		if valErr != nil {
			s.opts.LastError = valErr
			if s.opts.ErrorCallback != nil {
				s.opts.ErrorCallback(valErr)
			}
			for i := range val1 {
				val1[i] = 0
			}
			if s.opts.MaxRetries > 0 && attempt >= s.opts.MaxRetries {
				return nil, fmt.Errorf("%w: %v", ErrMaxRetries, valErr)
			}
			continue
		}

		if !s.confirm {
			return NewResult(val1), nil
		}

		// Confirmation value
		confirmAttempt := 1
		for {
			confirmCtx := Context{
				Prompt:     s.confirmMsg,
				Attempt:    confirmAttempt,
				MaxRetries: s.opts.MaxRetries,
				IsConfirm:  true,
				IsRetry:    confirmAttempt > 1 && s.opts.LastError != nil,
				LastError:  s.opts.LastError,
			}

			val2, err := s.readValue(confirmCtx)
			if err != nil {
				for i := range val1 {
					val1[i] = 0
				}
				return nil, err
			}

			if bytesEqual(val1, val2) {
				for i := range val2 {
					val2[i] = 0
				}
				return NewResult(val1), nil
			}

			// Mismatch - zero confirmation value, set error for red formatting
			for i := range val2 {
				val2[i] = 0
			}

			err = ErrMismatch
			s.opts.LastError = err
			if s.opts.ErrorCallback != nil {
				s.opts.ErrorCallback(err)
			}

			// Check max retries for this attempt
			if s.opts.MaxRetries > 0 && attempt >= s.opts.MaxRetries && confirmAttempt >= 3 {
				for i := range val1 {
					val1[i] = 0
				}
				return nil, fmt.Errorf("%w: %v", ErrMaxRetries, err)
			}

			confirmAttempt++
		}
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
