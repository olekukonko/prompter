package prompter

import "github.com/olekukonko/zero"

type Result struct {
	value []byte
}

func NewResult(v []byte) *Result {
	if v == nil {
		return &Result{value: nil}
	}
	buf := make([]byte, len(v))
	copy(buf, v)
	return &Result{value: buf}
}

// String returns the value as string.
// SECURITY WARNING: Creates immutable string in memory. Prefer Bytes()+Zero().
func (r *Result) String() string {
	if r.value == nil {
		return ""
	}
	return string(r.value)
}

func (r *Result) Bytes() []byte { return r.value }

func (r *Result) Len() int { return len(r.value) }

func (r *Result) Zero() {
	if r.value == nil {
		return
	}
	zero.Bytes(r.value)
	r.value = nil
}
