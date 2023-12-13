package schema_test

import (
	"errors"
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
		false,
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
		false,
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
		true,
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
		true,
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
		true,
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
	var asFunctionErr *schema.FunctionCallError
	ok := errors.As(err, &asFunctionErr)
	if !ok {
		t.Fatalf("Returned error is not a FunctionCallError")
	}
	assert.Equals(t, asFunctionErr.IsFunctionReportedError, true)
}

// Simple, one output.
func TestCallableFunctionSchema_OneReturn(t *testing.T) {
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil),
		false,
		nil,
		func() string {
			return "a"
		},
	)
	assert.NoError(t, err)
	result, err := simpleFunc.Call([]any{})
	// Valid result. Validate that it's "a", since that's what was returned in the handler.
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
		false,
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
		true,
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
	// In this test, the error is reported from the handler to the caller.
	simpleIntSchema := schema.NewIntSchema(nil, nil, nil)
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{simpleIntSchema, simpleIntSchema, simpleIntSchema},
		simpleIntSchema,
		true,
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
	var asFunctionErr *schema.FunctionCallError
	ok := errors.As(err, &asFunctionErr)
	if !ok {
		t.Fatalf("Returned error is not a FunctionCallError")
	}
	assert.Equals(t, asFunctionErr.IsFunctionReportedError, true)
}

func TestCallableFunctionSchema_Err_TestIncorrectNumArgs(t *testing.T) {
	// In this test, the arg count mismatches on input to Call.
	simpleIntSchema := schema.NewIntSchema(nil, nil, nil)
	simpleFunc, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{simpleIntSchema, simpleIntSchema, simpleIntSchema}, // Three params
		simpleIntSchema,
		true,
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
	var asFunctionErr *schema.FunctionCallError
	ok := errors.As(err, &asFunctionErr)
	if !ok {
		t.Fatalf("Returned error is not a FunctionCallError")
	}
	assert.Equals(t, asFunctionErr.IsFunctionReportedError, false)
}

// The following test requires bypassing the validation provided in NewCallableFunction.
func TestCallableFunctionSchema_Err_CallWrongErrorReturn(t *testing.T) {
	// In this test case, we bypass the protections in NewCallableFunction, and test the
	// late-validation by having a non-error return value specified that mismatches the schema.
	callable := schema.CallableFunctionSchema{
		IDValue:           "test",
		InputsValue:       []schema.Type{},
		StaticOutputValue: nil, // No return specified here, so only zero returns, or an error return is allowed.
		DisplayValue:      nil,
		Handler: reflect.ValueOf(func() any { // Non-error return schema.Type here
			return 5
		}),
	}
	_, err := callable.Call(make([]any, 0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error return val isn't an error")
	var asFunctionErr *schema.FunctionCallError
	ok := errors.As(err, &asFunctionErr)
	if !ok {
		t.Fatalf("Returned error is not a FunctionCallError")
	}
	assert.Equals(t, asFunctionErr.IsFunctionReportedError, false)
}

// The following test requires bypassing the validation provided in NewCallableFunction.
func TestCallableFunctionSchema_Err_CallWrongReturnCount(t *testing.T) {
	// In this test case, we bypass the protections in NewCallableFunction, and test the
	// late-validation by having an invalid return count that mismatches the schema.
	callable := schema.CallableFunctionSchema{
		IDValue:           "test",
		InputsValue:       []schema.Type{},
		StaticOutputValue: nil, // No returns specified here
		DisplayValue:      nil,
		Handler: reflect.ValueOf(func() (any, any, any) { // Three returns specified here
			return 0, 0, 0
		}),
	}
	_, err := callable.Call(make([]any, 0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected return count")
	var asFunctionErr *schema.FunctionCallError
	ok := errors.As(err, &asFunctionErr)
	if !ok {
		t.Fatalf("Returned error is not a FunctionCallError")
	}
	assert.Equals(t, asFunctionErr.IsFunctionReportedError, false)
}

// Test the input validation.
func TestNewCallableFunction_Err_MismatchedParamCount(t *testing.T) {
	// In this test case, the input schema has a different param count as the handler.
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0), // No params specified here
		nil,
		false,
		nil,
		func(int, int) {}, // Two params specified here
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parameter input counts do not match handler inputs")
}

func TestNewCallableFunction_Err_MismatchedParamType(t *testing.T) {
	// The handler will have a different output type (int) than the schema specifies (string)
	_, err := schema.NewCallableFunction(
		"test",
		[]schema.Type{schema.NewStringSchema(nil, nil, nil)}, // String specified here
		nil,
		false,
		nil,
		func(int) {}, // Int specified here, mismatched
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type mismatch for parameter")
}

func TestNewCallableFunction_Err_NilReturnMismatchedReturnCount(t *testing.T) {
	// In this case, the schema will specify zero returns, when the handler will have both a return type an en error return
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		true,
		nil,
		func() (int, error) { // Zero returns specified. That's wrong.
			return 0, nil
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect return count")
}

func TestNewCallableFunction_Err_NilReturnNotError(t *testing.T) {
	// In this case the handler will have a return type, when the schema specifies none.
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		false,
		nil,
		func() int { // Should be error or void.
			return 0
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect return count")
}

func TestNewCallableFunction_Err_NotEnoughReturns(t *testing.T) {
	// The schema specifies a string return, when the handler has no return. They need to match.
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil), // String specified here.
		false,
		nil,
		func() {}, // No returns here. We specified string earlier.
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect return count")
}

func TestNewCallableFunction_Err_MismatchedReturnType(t *testing.T) {
	// Handler specifies int, when the schema specifies string. They need to match.
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil), // String here
		false,
		nil,
		func() int { // Int here, mismatched
			return 0
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched return type")
}

func TestNewCallableFunction_Err_TooManyReturns(t *testing.T) {
	// In this case, the handler will have too many returns. We allow max the schema specified return,
	// plus an error return.
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil),
		true,
		nil,
		func() (string, string, error) { // Too may returns here
			return "", "", nil
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect return count")
}

func TestNewCallableFunction_Err_ReturnNotError(t *testing.T) {
	// In this test, the return value of the handler will include two non-error returns, when we
	// are expecting one string return, and one error return.
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil),
		true,
		nil,
		func() (string, string) { // Incorrect return schema.Type here
			return "", ""
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected last return type from handler to be error")
}

func TestNewCallableFunction_Err_TooManyReturnNoErr(t *testing.T) {
	// In this test, the return value of the handler will include two non-error returns, when we expect
	// one non-error return
	_, err := schema.NewCallableFunction(
		"test",
		make([]schema.Type, 0),
		schema.NewStringSchema(nil, nil, nil),
		false,
		nil,
		func() (string, string) { // Incorrect return schema.Type here
			return "", ""
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect return count")
}

func TestNewDynamicFunction_Simple(t *testing.T) {
	simpleFunc, err := schema.NewDynamicCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		func() (any, error) {
			return 5, nil
		},
		func(inputType []schema.Type) (schema.Type, error) {
			return schema.NewIntSchema(nil, nil, nil), nil
		},
	)
	assert.NoError(t, err)
	// Validate type
	typeOutput, errOut, err := simpleFunc.Output(make([]schema.Type, 0))
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewIntSchema(nil, nil, nil))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err := simpleFunc.Call([]any{})
	assert.NoError(t, err)
	assert.Equals(t, result, 5)
}

func TestNewDynamicFunction_WrongReturnType(t *testing.T) {
	_, err := schema.NewDynamicCallableFunction(
		"test",
		make([]schema.Type, 0),
		nil,
		func() (int, error) {
			return 5, nil
		},
		func(inputType []schema.Type) (schema.Type, error) {
			return schema.NewIntSchema(nil, nil, nil), nil
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 'any' return type for handler, but got int")
}

func TestNewDynamicFunction_1Param(t *testing.T) {
	simpleFunc, err := schema.NewDynamicCallableFunction(
		"test",
		[]schema.Type{schema.NewAnySchema()},
		nil,
		func(any) (any, error) {
			return 5, nil
		},
		func(inputType []schema.Type) (schema.Type, error) {
			return schema.NewIntSchema(nil, nil, nil), nil
		},
	)
	assert.NoError(t, err)
	// Validate type
	typeOutput, errOut, err := simpleFunc.Output(make([]schema.Type, 0))
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewIntSchema(nil, nil, nil))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err := simpleFunc.Call([]any{"test"})
	assert.NoError(t, err)
	assert.Equals(t, result, 5)
}

func TestNewDynamicFunction_SameType(t *testing.T) {
	simpleFunc, err := schema.NewDynamicCallableFunction(
		"test",
		[]schema.Type{schema.NewAnySchema()},
		nil,
		func(a any) (any, error) {
			return a, nil
		},
		func(inputTypes []schema.Type) (schema.Type, error) {
			assert.Equals(t, len(inputTypes), 1)
			return inputTypes[0], nil
		},
	)
	assert.NoError(t, err)

	// Validate type
	typeOutput, errOut, err := simpleFunc.Output([]schema.Type{schema.NewStringSchema(nil, nil, nil)})
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewStringSchema(nil, nil, nil))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err := simpleFunc.Call([]any{"test"})
	assert.NoError(t, err)
	assert.Equals(t, result, "test")

	// Validate type
	typeOutput, errOut, err = simpleFunc.Output([]schema.Type{schema.NewIntSchema(nil, nil, nil)})
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewIntSchema(nil, nil, nil))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err = simpleFunc.Call([]any{1})
	assert.NoError(t, err)
	assert.Equals(t, result, 1)

	// Validate type
	typeOutput, errOut, err = simpleFunc.Output([]schema.Type{schema.NewFloatSchema(nil, nil, nil)})
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewFloatSchema(nil, nil, nil))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err = simpleFunc.Call([]any{5.0})
	assert.NoError(t, err)
	assert.Equals(t, result, 5.0)

	// Validate type
	typeOutput, errOut, err = simpleFunc.Output([]schema.Type{
		schema.NewListSchema(schema.NewIntSchema(nil, nil, nil), nil, nil),
	})
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewListSchema(schema.NewIntSchema(nil, nil, nil), nil, nil))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err = simpleFunc.Call([]any{make([]int, 0)})
	assert.NoError(t, err)
	assert.Equals(t, result.([]int), make([]int, 0))
}

func TestNewDynamicFunction_SliceOfSameType(t *testing.T) {
	simpleFunc, err := schema.NewDynamicCallableFunction(
		"test",
		[]schema.Type{schema.NewAnySchema()},
		nil,
		func(a any) (any, error) {
			aVal := reflect.ValueOf(a)
			result := reflect.MakeSlice(reflect.SliceOf(aVal.Type()), 2, 2)
			result.Index(0).Set(aVal)
			result.Index(1).Set(aVal)
			return result.Interface(), nil
		},
		func(inputTypes []schema.Type) (schema.Type, error) {
			assert.Equals(t, len(inputTypes), 1)
			return schema.NewListSchema(inputTypes[0], nil, nil), nil
		},
	)
	assert.NoError(t, err)

	// Validate type
	typeOutput, errOut, err := simpleFunc.Output([]schema.Type{schema.NewStringSchema(nil, nil, nil)})
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewListSchema(
		schema.NewStringSchema(nil, nil, nil),
		nil, nil,
	))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err := simpleFunc.Call([]any{"a"})
	assert.NoError(t, err)
	assert.Equals(t, result.([]string), []string{"a", "a"})

	// Validate type
	typeOutput, errOut, err = simpleFunc.Output([]schema.Type{schema.NewIntSchema(nil, nil, nil)})
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewListSchema(
		schema.NewIntSchema(nil, nil, nil),
		nil, nil,
	))
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err = simpleFunc.Call([]any{1})
	assert.NoError(t, err)
	assert.Equals(t, result.([]int), []int{1, 1})

	// Validate type
	typeOutput, errOut, err = simpleFunc.Output([]schema.Type{
		schema.NewListSchema(schema.NewIntSchema(nil, nil, nil), nil, nil),
	})
	assert.NoError(t, err)
	assert.Equals[schema.Type](t, typeOutput, schema.NewListSchema(
		schema.NewListSchema(
			schema.NewIntSchema(nil, nil, nil), nil, nil),
		nil, nil),
	)
	assert.Equals(t, errOut, true)
	// Validate that the function call works as intended
	result, err = simpleFunc.Call([]any{[]int{1}})
	assert.NoError(t, err)
	assert.Equals(t, result.([][]int), [][]int{{1}, {1}})
}

func TestFunctionToStringOneParamVoid(t *testing.T) {
	oneParamVoidFunction, err := schema.NewCallableFunction(
		"a",
		[]schema.Type{schema.NewStringSchema(nil, nil, nil)},
		nil,
		false,
		nil,
		func(a string) {},
	)
	assert.NoError(t, err)
	funcStr := oneParamVoidFunction.String()
	assert.Equals(t, funcStr, "a(string) void")
}

func TestFunctionToStringToFunctionTwoParam(t *testing.T) {
	callableFunc, err := schema.NewCallableFunction(
		"b",
		[]schema.Type{
			schema.NewStringSchema(nil, nil, nil),
			schema.NewIntSchema(nil, nil, nil),
		},
		schema.NewIntSchema(nil, nil, nil),
		false,
		nil,
		func(a string, b int64) int64 { return 0 },
	)
	assert.NoError(t, err)
	funcStr := callableFunc.String()
	assert.Equals(t, funcStr, "b(string, integer) integer")

	// Now the version in FunctionSchema instead of CallableFunction
	funcSchema, err := callableFunc.ToFunctionSchema()
	assert.NoError(t, err)

	funcStr = funcSchema.String()
	assert.Equals(t, funcStr, "b(string, integer) integer")
}

func TestDynamicFunctionToStringToFunction(t *testing.T) {
	oneParamVoidFunction, err := schema.NewDynamicCallableFunction(
		"c",
		[]schema.Type{schema.NewAnySchema(), schema.NewIntSchema(nil, nil, nil)},
		nil,
		func(a any, b int64) (any, error) {
			return b, nil
		},
		func(inputType []schema.Type) (schema.Type, error) {
			return schema.NewIntSchema(nil, nil, nil), nil
		},
	)
	assert.NoError(t, err)
	funcStr := oneParamVoidFunction.String()
	assert.Equals(t, funcStr, "c(any, integer) (dynamic, error)")

	// Cannot be represented as a FunctionSchema due to the dynamic typing.
	_, err = oneParamVoidFunction.ToFunctionSchema()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dynamic typing")
}
