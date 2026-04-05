package prompter

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var (
	ErrMaxRetries = errors.New("maximum retry attempts exceeded")
	ErrMismatch   = errors.New("values do not match")
	ErrCancelled  = errors.New("input cancelled")
)

type ErrValidation struct{ Msg string }

func (e ErrValidation) Error() string { return e.Msg }

type Context struct {
	Ctx        context.Context
	Prompt     string
	Attempt    int
	MaxRetries int
	IsRetry    bool
	LastError  error
	IsConfirm  bool
}

type Formatter func(ctx Context) string
type Validator func(value []byte) error
type EmptyHandler func() (proceedToConfirm bool)

type Options struct {
	Required      bool
	MinLen        int
	MaxLen        int
	MaxRetries    int
	Validator     Validator
	Formatter     Formatter
	ErrorCallback func(error)
	Input         io.Reader
}

// DefaultFormatter formats a prompt with attempt count on retry.
func DefaultFormatter(ctx Context) string {
	if ctx.IsRetry && ctx.LastError != nil {
		if ctx.MaxRetries > 0 {
			return fmt.Sprintf("[attempt %d/%d] %s: ", ctx.Attempt, ctx.MaxRetries, ctx.Prompt)
		}
		return fmt.Sprintf("[attempt %d] %s: ", ctx.Attempt, ctx.Prompt)
	}
	return fmt.Sprintf("%s: ", ctx.Prompt)
}
