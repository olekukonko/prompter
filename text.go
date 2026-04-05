package prompter

import (
	"bufio"
	"fmt"
	"os"
)

// Input reads normal text (with echo) from stdin.
type Input struct {
	opts   Options
	prompt string
}

// NewTextInput creates a text input prompter with shared options.
func NewTextInput(prompt string, opts ...Option) *Input {
	t := &Input{
		prompt: prompt,
		opts: Options{
			Formatter: DefaultFormatter,
		},
	}
	for _, opt := range opts {
		opt(&t.opts)
	}
	return t
}

// Run executes the prompt.
func (t *Input) Run() (*Result, error) {
	reader := bufio.NewReader(os.Stdin)

	for attempt := 1; ; attempt++ {
		ctx := Context{
			Prompt:     t.prompt,
			Attempt:    attempt,
			MaxRetries: t.opts.MaxRetries,
			IsRetry:    attempt > 1,
		}
		if ctx.IsRetry && t.opts.LastError != nil {
			ctx.LastError = t.opts.LastError
		}

		fmt.Fprintf(os.Stderr, t.opts.Formatter(ctx), t.prompt)

		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		// Trim newlines
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
		}

		// Validation
		var valErr error
		switch {
		case t.opts.Required && len(line) == 0:
			valErr = ErrValidation{Msg: "input is required"}
		case t.opts.MinLen > 0 && len(line) < t.opts.MinLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("minimum length is %d", t.opts.MinLen)}
		case t.opts.MaxLen > 0 && len(line) > t.opts.MaxLen:
			valErr = ErrValidation{Msg: fmt.Sprintf("maximum length is %d", t.opts.MaxLen)}
		case t.opts.Validator != nil:
			valErr = t.opts.Validator(line)
		}

		if valErr == nil {
			return NewResult(line), nil
		}

		t.opts.LastError = valErr
		if t.opts.ErrorCallback != nil {
			t.opts.ErrorCallback(valErr)
		}

		if t.opts.MaxRetries > 0 && attempt >= t.opts.MaxRetries {
			return nil, fmt.Errorf("%w: %v", ErrMaxRetries, valErr)
		}
	}
}
