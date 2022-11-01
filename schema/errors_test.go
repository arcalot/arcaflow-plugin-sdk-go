package schema_test

import (
	"errors"
	"fmt"
	"testing"

	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestConstraintErrorAddPath(t *testing.T) {
	e := &schema.ConstraintError{
		Message: "Test message",
	}
	assert.Equals(t, len(e.Path), 0)

	_ = e.AddPathSegment("test1")
	assert.Equals(t, len(e.Path), 1)
	assert.Equals(t, e.Path[0], "test1")

	e2 := schema.ConstraintErrorAddPathSegment(e, "test2")
	if !errors.As(e2, &e) {
		assert.Equals(t, len(e.Path), 2)
		assert.Equals(t, e.Path[0], "test2")
		assert.Equals(t, e.Path[1], "test1")
	}
}

func TestConstraintErrorUnwrap(t *testing.T) {
	testErr := fmt.Errorf("test error")

	e := &schema.ConstraintError{
		Message: "Test message",
		Cause:   testErr,
	}

	if !errors.Is(e, testErr) {
		t.Fatal("Unwrap doesn't work properly.")
	}
}

func TestConstraintErrorAddPathSegment(t *testing.T) {
	e := fmt.Errorf("test error")
	e2 := schema.ConstraintErrorAddPathSegment(e, "test2")
	if !errors.Is(e2, e) {
		t.Fatal("error mismatch")
	}
}

func TestConstraintErrorMessage(t *testing.T) {
	assert.Equals(t, (&schema.ConstraintError{Message: "Test"}).Error(), "Validation failed: Test")
	assert.Equals(
		t,
		(&schema.ConstraintError{Message: "Test", Cause: fmt.Errorf("test2")}).Error(),
		"Validation failed: Test (test2)",
	)
	assert.Equals(
		t,
		(&schema.ConstraintError{Message: "Test", Path: []string{"test1"}}).Error(),
		"Validation failed for 'test1': Test",
	)
	assert.Equals(
		t,
		(&schema.ConstraintError{Message: "Test", Path: []string{"test1", "test2"}}).Error(),
		"Validation failed for 'test1' -> 'test2': Test",
	)
	assert.Equals(
		t,
		(&schema.ConstraintError{
			Message: "Test",
			Path:    []string{"test1", "test2"},
			Cause:   fmt.Errorf("test"),
		}).Error(),
		"Validation failed for 'test1' -> 'test2': Test (test)",
	)
}

func TestNoSuchStepErrorMessage(t *testing.T) {
	assert.Equals(t, (&schema.NoSuchStepError{Step: "test"}).Error(), "No such step: test")
}

func TestBadArgumentErrorMessage(t *testing.T) {
	assert.Equals(t, (&schema.BadArgumentError{Message: "test"}).Error(), "test")
	assert.Equals(
		t,
		(&schema.BadArgumentError{Message: "test", Cause: fmt.Errorf("test2")}).Error(),
		"test (test2)",
	)
}

func TestBadArgumentErrorUnwrap(t *testing.T) {
	testErr := fmt.Errorf("test error")

	e := &schema.BadArgumentError{
		Message: "Test message",
		Cause:   testErr,
	}

	if !errors.Is(e, testErr) {
		t.Fatal("Unwrap doesn't work properly.")
	}
}

func TestUnitParseErrorMessage(t *testing.T) {
	assert.Equals(t, (&schema.UnitParseError{Message: "test"}).Error(), "test")
	assert.Equals(
		t,
		(&schema.UnitParseError{Message: "test", Cause: fmt.Errorf("test2")}).Error(),
		"test (test2)",
	)
}

func TestUnitParseErrorUnwrap(t *testing.T) {
	testErr := fmt.Errorf("test error")

	e := &schema.UnitParseError{
		Message: "Test message",
		Cause:   testErr,
	}

	if !errors.Is(e, testErr) {
		t.Fatal("Unwrap doesn't work properly.")
	}
}
