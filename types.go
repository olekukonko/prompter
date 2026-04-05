package prompter

import "errors"

var (
	ErrMaxRetries = errors.New("maximum retry attempts exceeded")
	ErrMismatch   = errors.New("values do not match")
)

type ErrValidation struct{ Msg string }

func (e ErrValidation) Error() string { return e.Msg }

// Context gives formatters rich information for styling.
type Context struct {
	Prompt     string
	Attempt    int
	MaxRetries int
	IsRetry    bool
	LastError  error
	IsConfirm  bool
}

type Formatter func(ctx Context) string
type Validator func(value []byte) error

// EmptyHandler is called when secret input is empty (before confirmation).
// Return true to proceed to confirmation, false to retry immediately.
type EmptyHandler func() (proceedToConfirm bool)

// Options holds configuration shared by Input and Secret.
type Options struct {
	Required      bool
	MinLen        int
	MaxLen        int
	MaxRetries    int
	Validator     Validator
	Formatter     Formatter
	ErrorCallback func(error)
	LastError     error
}

// DefaultFormatter is a simple plain-text formatter.
func DefaultFormatter(ctx Context) string {
	if ctx.IsRetry && ctx.LastError != nil {
		return "[attempt %d/%d] %s: "
	}
	return "%s: "
}
