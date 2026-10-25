package transaction

import (
	"errors"
	"maps"
	"slices"
	"strings"
)

// Sentinel errors let the HTTP layer map domain failures to status codes
// without knowing anything about Postgres.
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// ValidationError reports which input fields are invalid and why.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, k := range slices.Sorted(maps.Keys(e.Fields)) {
		parts = append(parts, k+" "+e.Fields[k])
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

type validator struct {
	fields map[string]string
}

func (v *validator) check(ok bool, field, msg string) {
	if ok {
		return
	}
	if v.fields == nil {
		v.fields = make(map[string]string)
	}
	if _, exists := v.fields[field]; !exists {
		v.fields[field] = msg
	}
}

func (v *validator) err() error {
	if len(v.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: v.fields}
}
