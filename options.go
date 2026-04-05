package prompter

// Option configures shared input behavior.
type Option func(*Options)

// WithRequired fails on empty input.
func WithRequired(required bool) Option {
	return func(o *Options) { o.Required = required }
}

// WithLength sets min/max byte length (0 = no limit).
func WithLength(min, max int) Option {
	return func(o *Options) { o.MinLen, o.MaxLen = min, max }
}

// WithMaxRetries limits attempts. 0 = unlimited.
func WithMaxRetries(n int) Option {
	return func(o *Options) { o.MaxRetries = n }
}

// WithValidator adds custom validation.
func WithValidator(v Validator) Option {
	return func(o *Options) { o.Validator = v }
}

// WithFormatter sets the prompt formatter.
func WithFormatter(f Formatter) Option {
	return func(o *Options) { o.Formatter = f }
}

// WithErrorCallback is called on validation errors (for logging/styling).
func WithErrorCallback(fn func(error)) Option {
	return func(o *Options) { o.ErrorCallback = fn }
}
