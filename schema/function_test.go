package schema_test

import (
	"fmt"
	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
	"reflect"
	"testing"
)

// Simple, no input or output.
func TestCallableFunctionSchema_Simple(t *testing.T) {
	called := false
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		nil,
		func() {
			called = true
		},
	)
	assert.NoError(t, err)
	assert.Equals(t, called, false)
	result, err := simpleFunc.Call([]any{})
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.Equals(t, called, true)
}

// Simple, with one input.
func TestCallableFunctionSchema_Simple1Param(t *testing.T) {
	passedInVal := ""
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{schema.NewStringSchema(nil, nil, nil)},
		nil,
		nil,
		func(test string) {
			passedInVal = test
		},
	)
	assert.NoError(t, err)
	result, err := simpleFunc.Call([]any{"a"})
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.Equals(t, passedInVal, "a")
}

// Simple, with no input, and only error out.
func TestCallableFunctionSchema_SimpleWithNilErr(t *testing.T) {
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		nil,
		func() error {
			return nil
		},
	)
	assert.NoError(t, err)
	result, err := simpleFunc.Call([]any{})
	// No result or error
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// One input, nil error output.
func TestCallableFunctionSchema_1ParamWithNilErr(t *testing.T) {
	passedInVal := ""
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{schema.NewStringSchema(nil, nil, nil)},
		nil,
		nil,
		func(test string) error {
			passedInVal = test
			return nil
		},
	)
	assert.NoError(t, err)
	result, err := simpleFunc.Call([]any{"a"})
	assert.NoError(t, err)
	assert.Nil(t, result)
	// Validate that the schema.Type got passed in properly.
	assert.Equals(t, passedInVal, "a")
}

// Simple, with error out.
func TestCallableFunctionSchema_SimpleWithErr(t *testing.T) {
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		nil,
		func() error {
			return fmt.Errorf("this is an error")
		},
	)
	assert.NoError(t, err)
	_, err = simpleFunc.Call([]any{})
	// There should be an error from the function. Validate that it's the correct one.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "this is an error")
}

// Simple, one output.
func TestCallableFunctionSchema_OneReturn(t *testing.T) {
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		func() string {
			return "a"
		},
	)
	assert.NoError(t, err)
	result, err := simpleFunc.Call([]any{})
	// There should be an error. Validate that it's the correct one.
	assert.NoError(t, err)
	assert.InstanceOf[string](t, result)
	assert.Equals(t, result.(string), "a")
}

// Multi-param, with int output.
func TestCallableFunctionSchema_MultiParam(t *testing.T) {
	simpleIntSchema := schema.NewIntSchema(nil, nil, nil)
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{simpleIntSchema, simpleIntSchema, simpleIntSchema, simpleIntSchema, simpleIntSchema, simpleIntSchema},
		simpleIntSchema,
		nil,
		func(a int64, b int64, c int64, d int64, e int64, f int64) int64 {
			return a * b * c * d * e * f
		},
	)
	assert.NoError(t, err)
	result, err := simpleFunc.Call([]any{int64(1), int64(2), int64(3), int64(4), int64(5), int64(6)})
	assert.NoError(t, err)
	// Validate that the schema.Type got passed in properly.
	assert.Equals[int64](t, result.(int64), int64(720))
}

// Multi-param, with int output.
func TestCallableFunctionSchema_MultiParamWithNilErr(t *testing.T) {
	simpleIntSchema := schema.NewIntSchema(nil, nil, nil)
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{simpleIntSchema, simpleIntSchema, simpleIntSchema},
		simpleIntSchema,
		nil,
		func(a int64, b int64, c int64) (int64, error) {
			return a * b * c, nil
		},
	)
	assert.NoError(t, err)
	result, err := simpleFunc.Call([]any{int64(1), int64(2), int64(3)})
	assert.NoError(t, err)
	// Validate that the schema.Type got passed in properly.
	assert.Equals[int64](t, result.(int64), int64(6))
}

// Multi-param, with int output.
func TestCallableFunctionSchema_Err_MultiParamWithErr(t *testing.T) {
	simpleIntSchema := schema.NewIntSchema(nil, nil, nil)
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{simpleIntSchema, simpleIntSchema, simpleIntSchema},
		simpleIntSchema,
		nil,
		func(a int64, b int64, c int64) (int64, error) {
			return a * b * c, fmt.Errorf("this is an error")
		},
	)
	assert.NoError(t, err)
	_, err = simpleFunc.Call([]any{int64(1), int64(2), int64(3)})
	// There should be an error. Validate that it's the correct one.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "this is an error")
}

func TestCallableFunctionSchema_Err_TestIncorrectNumArgs(t *testing.T) {
	simpleIntSchema := schema.NewIntSchema(nil, nil, nil)
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{simpleIntSchema, simpleIntSchema, simpleIntSchema}, // Three params
		simpleIntSchema,
		nil,
		func(a int64, b int64, c int64) (int64, error) { // Three params
			return a * b * c, nil
		},
	)
	assert.NoError(t, err)
	_, err = simpleFunc.Call([]any{int64(1), int64(2)}) // Two args specified here
	// There should be an error. Validate that it's the correct one.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect number of args")
}

// The following test requires bypassing the validation provided in NewCallableFunction.
func TestCallableFunctionSchema_Err_CallWrongErrorReturn(t *testing.T) {
	callable := schema.CallableFunctionSchema{
		IDValue:      "test",
		InputsValue:  []schema.Type{},
		OutputValue:  nil, // No return specified here, so only zero returns, or an error return is allowed.
		DisplayValue: nil,
		Handler: reflect.ValueOf(func() any { // Non-error return schema.Type here
			return 5
		}),
	}
	_, err := callable.Call(make([]any, 0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error return val isn't an error")
}

// The following test requires bypassing the validation provided in NewCallableFunction.
func TestCallableFunctionSchema_Err_CallWrongReturnCount(t *testing.T) {
	callable := schema.CallableFunctionSchema{
		IDValue:      "test",
		InputsValue:  []schema.Type{},
		OutputValue:  nil, // No returns specified here
		DisplayValue: nil,
		Handler: reflect.ValueOf(func() (any, any, any) { // Three returns specified here
			return 0, 0, 0
		}),
	}
	_, err := callable.Call(make([]any, 0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected return count")
}

// Test the input validation.
func TestNewCallableFunction_Err_MismatchedParamCount(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0), // No params specified here
		nil,
		nil,
		func(int, int) {}, // Two params specified here
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parameter inputs do not match handler inputs")
}

func TestNewCallableFunction_Err_MismatchedParamType(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{schema.NewStringSchema(nil, nil, nil)}, // String specified here
		nil,
		nil,
		func(int) {}, // Int specified here, mismatched
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema.Type mismatch for parameter")
}

func TestNewCallableFunction_Err_NilReturnMismatchedReturnCount(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		nil,
		func() (int, error) { // Zero returns specified. That's wrong.
			return 0, nil
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parameter output is nil, meaning it's a void function")
}

func TestNewCallableFunction_Err_NilReturnNotError(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		nil,
		func() int { // This isn't a valid return given the schema. Should be error or void.
			return 0
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected void or error return")
}

func TestNewCallableFunction_Err_NotEnoughReturns(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil), // String
		nil,
		func() {}, // No returns here. We specified string earlier.
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected handler to have one return")
}

func TestNewCallableFunction_Err_MismatchedReturnType(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil), // String here
		nil,
		func() int { // Int here, mismatched
			return 0
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched return schema.Type")
}

func TestNewCallableFunction_Err_TooManyReturns(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		func() (string, string, error) { // Too may returns here
			return "", "", nil
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected handler to have one return")
}

func TestNewCallableFunction_Err_ReturnNotError(t *testing.T) {
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		func() (string, string) { // Incorrect return schema.Type here
			return "", ""
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected additional return schema.Type to be an error return")
}
