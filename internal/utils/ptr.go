// Package utils provides shared utilities for the Token.io SDK.
package utils

// Ptr returns a pointer to the given value. Useful for optional struct fields.
func Ptr[T any](v T) *T { return &v }

// Deref safely dereferences a pointer, returning the zero value when nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
