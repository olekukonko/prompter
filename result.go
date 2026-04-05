package prompter

import "github.com/olekukonko/zero"

type Result struct {
	value []byte
}

// NewResult allocates a copy of v so the caller's buffer can be zeroed independently.
func NewResult(v []byte) *Result {
	if v == nil {
		return &Result{value: nil}
	}
	buf := make([]byte, len(v))
	copy(buf, v)
	return &Result{value: buf}
}

// String returns the value as a string.
// SECURITY WARNING: Creates an immutable Go string; prefer Bytes()+Zero() for secrets.
func (r *Result) String() string {
	if r.value == nil {
		return ""
	}
	return string(r.value)
}

// Bytes returns the raw byte slice; do not retain the reference after calling Zero.
func (r *Result) Bytes() []byte { return r.value }

// Len returns the length of the stored value.
func (r *Result) Len() int { return len(r.value) }

// Zero wipes the value from memory and sets the internal slice to nil.
func (r *Result) Zero() {
	if r.value == nil {
		return
	}
	zero.Bytes(r.value)
	r.value = nil
}
