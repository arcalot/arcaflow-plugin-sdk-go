package schema

import (
	"errors"
	"fmt"
	"strings"
)

// ConstraintError indicates that the passed data violated one or more constraints defined in the schema.
// The message holds the exact path of the problematic field, as well as a message explaining the error.
// If this error is not easily understood, please open an issue on the Arcaflow plugin SDK.
type ConstraintError struct {
	Message string
	Path    []string
	Cause   error
}

// Error returns the error message.
func (c *ConstraintError) Error() string {
	pathDescriptor := ""
	if len(c.Path) > 0 {
		pathDescriptor = " for '" + strings.Join(c.Path, "' -> '") + "'"
	}
	result := fmt.Sprintf("Validation failed%s: %s", pathDescriptor, c.Message)
	if c.Cause != nil {
		result += " (" + c.Cause.Error() + ")"
	}
	return result
}

// AddPathSegment adds a path segment to the constraint error.
func (c *ConstraintError) AddPathSegment(pathSegment string) error {
	c.Path = append([]string{pathSegment}, c.Path...)
	return c
}

// Unwrap returns the underlying error if any.
func (c *ConstraintError) Unwrap() error {
	return c.Cause
}

// ConstraintErrorAddPathSegment adds a path segment if a ConstraintError is found.
func ConstraintErrorAddPathSegment(err error, pathSegment string) error {
	var c *ConstraintError
	if errors.As(err, &c) {
		return c.AddPathSegment(pathSegment)
	}
	return err
}

// NoSuchStepError indicates that the given step is not supported by the plugin.
type NoSuchStepError struct {
	Step string
}

// Error returns the error message.
func (n NoSuchStepError) Error() string {
	return fmt.Sprintf("No such step: %s", n.Step)
}

// BadArgumentError indicates that an invalid configuration was passed to a schema component. The message will
// explain what exactly the problem is, but may not be able to locate the exact error as the schema may be manually
// built.
type BadArgumentError struct {
	Message string
	Cause   error
}

// Error returns the error message.
func (b BadArgumentError) Error() string {
	result := b.Message
	if b.Cause != nil {
		result += " (" + b.Cause.Error() + ")"
	}
	return result
}

// Unwrap returns the underlying error if any.
func (b BadArgumentError) Unwrap() error {
	return b.Cause
}

// UnitParseError indicates that it failed to parse a unit string.
type UnitParseError struct {
	Message string
	Cause   error
}

func (u UnitParseError) Error() string {
	result := u.Message
	if u.Cause != nil {
		result += " (" + u.Cause.Error() + ")"
	}
	return result
}

// Unwrap returns the underlying error if any.
func (u UnitParseError) Unwrap() error {
	return u.Cause
}
