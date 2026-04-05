package prompter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

const maxTextLineLength = 4096

type Input struct {
	opts   Options
	prompt string
	reader io.Reader
}

// NewTextInput creates a plain-text prompt. Does NOT zero memory; avoid for secrets.
func NewTextInput(prompt string, opts ...Option) *Input {
	t := &Input{
		prompt: prompt,
		reader: os.Stdin,
		opts: Options{
			Formatter: DefaultFormatter,
		},
	}
	for _, opt := range opts {
		opt(&t.opts)
	}
	if t.opts.Input != nil {
		t.reader = t.opts.Input
	}
	return t
}

// Run executes the text prompt using a background context.
func (t *Input) Run() (*Result, error) {
	return t.RunContext(context.Background())
}

// RunContext executes the text prompt, respecting context cancellation.
// Note: cancellation is best-effort; the underlying read may block until the user acts.
func (t *Input) RunContext(ctx context.Context) (*Result, error) {
	type response struct {
		val []byte
		err error
	}
	ch := make(chan response, 1)
	go func() {
		val, err := t.run(ctx)
		ch <- response{val, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return NewResult(r.val), r.err
	}
}

func (t *Input) run(ctx context.Context) ([]byte, error) {
	var lastErr error
	reader := bufio.NewReader(t.reader)
	for attempt := 1; ; attempt++ {
		pctx := Context{
			Ctx:        ctx,
			Prompt:     t.prompt,
			Attempt:    attempt,
			MaxRetries: t.opts.MaxRetries,
			IsRetry:    attempt > 1,
			LastError:  lastErr,
		}
		fmt.Fprint(os.Stderr, t.opts.Formatter(pctx))
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF && t.opts.Required {
				return nil, ErrValidation{Msg: "input is required"}
			}
			return nil, err
		}
		if len(line) > maxTextLineLength {
			return nil, ErrValidation{Msg: "input too long"}
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
		}
		var valErr error
		switch {
		case t.opts.Required && len(line) == 0:
			valErr = ErrValidation{Msg: "input is required"}
		case t.opts.MinLen > 0 && len(line) < t.opts.MinLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("minimum length is %d", t.opts.MinLen)}
		case t.opts.MaxLen > 0 && len(line) > t.opts.MaxLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("maximum length is %d", t.opts.MaxLen)}
		case t.opts.Validator != nil:
			valErr = safeValidate(t.opts.Validator, line)
		}
		if valErr == nil {
			return line, nil
		}
		lastErr = valErr
		if t.opts.ErrorCallback != nil {
			t.opts.ErrorCallback(valErr)
		}
		if t.opts.MaxRetries > 0 && attempt >= t.opts.MaxRetries {
			return nil, fmt.Errorf("%w: %v", ErrMaxRetries, valErr)
		}
	}
}
