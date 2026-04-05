package prompter

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/zero"
	"golang.org/x/term"
)

const maxSecretLineLength = 4096

type Secret struct {
	opts       Options
	prompt     string
	confirm    bool
	confirmMsg string
	onEmpty    EmptyHandler
	input      io.Reader
}

// NewSecret creates a Secret prompter with the given prompt and options.
// Default formatting adds "(hidden)" or "(confirm)" suffixes and shows attempt counts on retry.
func NewSecret(prompt string, opts ...Option) *Secret {
	s := &Secret{
		prompt: prompt,
		input:  os.Stdin,
		opts: Options{
			Formatter: func(ctx Context) string {
				suffix := " (hidden)"
				if ctx.IsConfirm {
					suffix = " (confirm)"
				}
				if ctx.IsRetry && ctx.LastError != nil {
					if ctx.MaxRetries > 0 {
						return fmt.Sprintf("[attempt %d/%d] %s%s: ", ctx.Attempt, ctx.MaxRetries, ctx.Prompt, suffix)
					}
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
	if s.opts.Input != nil {
		s.input = s.opts.Input
	}
	return s
}

// WithConfirmation enables two-entry confirmation mode.
// The optional msg overrides the default "Confirm" prompt text.
func (s *Secret) WithConfirmation(msg string) *Secret {
	s.confirm = true
	if msg != "" {
		s.confirmMsg = msg
	}
	return s
}

// WithOnEmpty sets a callback invoked when the user submits empty input.
// Return true to accept the empty value, false to retry.
func (s *Secret) WithOnEmpty(h EmptyHandler) *Secret {
	s.onEmpty = h
	return s
}

// Run executes the secret prompt using a background context.
func (s *Secret) Run() (*Result, error) {
	return s.RunContext(context.Background())
}

// RunContext executes the secret prompt, respecting context cancellation.
// Note: cancellation is best-effort; the underlying read may block until the user acts.
func (s *Secret) RunContext(ctx context.Context) (*Result, error) {
	type response struct {
		val []byte
		err error
	}
	ch := make(chan response, 1)
	go func() {
		val, err := s.run()
		ch <- response{val, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return NewResult(r.val), r.err
	}
}

func (s *Secret) run() ([]byte, error) {
	var lastErr error
	for attempt := 1; ; attempt++ {
		ctx := Context{
			Prompt:     s.prompt,
			Attempt:    attempt,
			MaxRetries: s.opts.MaxRetries,
			IsRetry:    attempt > 1 && lastErr != nil,
			LastError:  lastErr,
		}
		val1, err := s.readValue(ctx)
		if err != nil {
			if err == io.EOF {
				if s.opts.Required {
					return nil, ErrValidation{Msg: "input is required"}
				}
				if lastErr != nil {
					return nil, lastErr
				}
				return []byte{}, nil
			}
			return nil, err
		}
		if len(val1) == 0 && s.onEmpty != nil {
			if !s.onEmpty() {
				if s.opts.MaxRetries > 0 && attempt >= s.opts.MaxRetries {
					return nil, fmt.Errorf("%w: empty input refused", ErrMaxRetries)
				}
				lastErr = ErrValidation{Msg: "empty input"}
				continue
			}
		}
		var valErr error
		switch {
		case s.opts.Required && len(val1) == 0:
			valErr = ErrValidation{Msg: "input is required"}
		case s.opts.MinLen > 0 && len(val1) < s.opts.MinLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("minimum length is %d", s.opts.MinLen)}
		case s.opts.MaxLen > 0 && len(val1) > s.opts.MaxLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("maximum length is %d", s.opts.MaxLen)}
		case s.opts.Validator != nil:
			valErr = safeValidate(s.opts.Validator, val1)
		}
		if valErr != nil {
			lastErr = valErr
			if s.opts.ErrorCallback != nil {
				s.opts.ErrorCallback(valErr)
			}
			zero.Bytes(val1)
			if s.opts.MaxRetries > 0 && attempt >= s.opts.MaxRetries {
				return nil, fmt.Errorf("%w: %v", ErrMaxRetries, valErr)
			}
			continue
		}
		if !s.confirm {
			return val1, nil
		}
		val2, err := s.readConfirm(ctx)
		if err != nil {
			zero.Bytes(val1)
			if err == io.EOF {
				return nil, ErrMismatch
			}
			return nil, err
		}
		if subtle.ConstantTimeCompare(val1, val2) == 1 {
			zero.Bytes(val2)
			return val1, nil
		}
		zero.Bytes(val2)
		lastErr = ErrMismatch
		if s.opts.ErrorCallback != nil {
			s.opts.ErrorCallback(lastErr)
		}
		if s.opts.MaxRetries > 0 && attempt >= s.opts.MaxRetries {
			zero.Bytes(val1)
			return nil, fmt.Errorf("%w: %v", ErrMaxRetries, lastErr)
		}
		zero.Bytes(val1)
	}
}

func (s *Secret) readValue(ctx Context) ([]byte, error) {
	if r, ok := s.input.(interface{ Fd() uintptr }); ok && r != os.Stdin {
		fd := int(r.Fd())
		if !term.IsTerminal(fd) {
			return s.readFromReader(ctx)
		}
		fmt.Fprint(os.Stderr, s.opts.Formatter(ctx))
		val, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		return val, err
	}
	if s.input == os.Stdin {
		fd := int(os.Stdin.Fd())
		if term.IsTerminal(fd) {
			fmt.Fprint(os.Stderr, s.opts.Formatter(ctx))
			val, err := term.ReadPassword(fd)
			fmt.Fprintln(os.Stderr)
			return val, err
		}
	}
	return s.readFromReader(ctx)
}

// readFromReader reads one line byte-by-byte to avoid bufio pre-fetching consuming
// bytes that belong to subsequent reads (e.g. the confirmation prompt).
// Each staging byte is zeroed immediately after use.
func (s *Secret) readFromReader(ctx Context) ([]byte, error) {
	fmt.Fprint(os.Stderr, s.opts.Formatter(ctx))

	result := make([]byte, 0, 64)
	oneByte := make([]byte, 1)

	for {
		_, err := s.input.Read(oneByte)
		if err != nil {
			if err == io.EOF {
				if len(result) > 0 {
					break
				}
				zero.Bytes(result)
				return nil, io.EOF
			}
			zero.Bytes(result)
			return nil, err
		}
		b := oneByte[0]
		oneByte[0] = 0

		if b == '\n' {
			break
		}
		if b == '\r' {
			continue
		}
		if len(result) >= maxSecretLineLength {
			zero.Bytes(result)
			return nil, ErrValidation{Msg: "input too long"}
		}
		result = append(result, b)
	}

	return result, nil
}

func (s *Secret) readConfirm(ctx Context) ([]byte, error) {
	confirmCtx := Context{
		Prompt:     s.confirmMsg,
		Attempt:    1,
		MaxRetries: s.opts.MaxRetries,
		IsConfirm:  true,
	}
	return s.readValue(confirmCtx)
}

// safeValidate calls a user-supplied Validator, recovering from panics.
func safeValidate(v Validator, val []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("validator panicked: %v", r)
		}
	}()
	return v(val)
}
