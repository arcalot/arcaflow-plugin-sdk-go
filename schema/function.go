package schema

import (
	"fmt"
	"reflect"
	"strings"
)

type Function interface {
	ID() string
	Parameters() []Type
	// Output determines the output type. This can be static, or it can depend on the input types.
	// It also returns whether the handler may self-report an error.
	Output([]Type) (Type, bool, error)
	Display() Display
	String() string
}

type CallableFunction interface {
	Function
	ToFunctionSchema() (*FunctionSchema, error)
	Call(arguments []any) (any, error)
}

type FunctionCallError struct {
	// isFunctionReportedError is true when the error originated from the function itself from its return value or a panic
	// It is false when the function did something unexpected, or the call is invalid.
	IsFunctionReportedError bool
	SourceError             error
}

func NewFunctionCallError(err error, isFunctionReportedError bool) *FunctionCallError {
	return &FunctionCallError{
		isFunctionReportedError,
		err,
	}
}

func (e *FunctionCallError) Error() string {
	return e.SourceError.Error()
}

const errorType = "error"

// NewCallableFunction creates a CallableFunction schema type for the strictly typed function.
//
// - The handler types must match the input and output types specified.
// - The return type of the handler is determined by the value specified for output. If output is nil, it's a void
// function that may have no return, or a single error return if outputsError is true.
// - If output is not nil, the return type must match, plus it must have an error return if outputsError is true.
func NewCallableFunction(
	id string,
	inputs []Type,
	output Type,
	outputsError bool,
	display Display,
	handler any,
) (CallableFunction, error) {
	parsedHandler := reflect.ValueOf(handler)
	// Validate the input types match the provided ones.
	err := validateInputTypeCompatibility(inputs, parsedHandler)
	if err != nil {
		return nil, err
	}
	// Validate the output type
	err = validateTypedReturnFunc(parsedHandler, outputsError, output)
	if err != nil {
		return nil, err
	}
	return &CallableFunctionSchema{
		IDValue:           id,
		InputsValue:       inputs,
		StaticOutputValue: output,
		OutputsError:      outputsError,
		DisplayValue:      display,
		Handler:           parsedHandler,
	}, nil
}

func validateTypedReturnFunc(parsedHandler reflect.Value, errorExpected bool, outputType Type) error {
	returnCount := parsedHandler.Type().NumOut()
	var expectedReturnCount int
	// Validate error return
	if errorExpected {
		expectedReturnCount = 1
	} else {
		expectedReturnCount = 0
	}
	// Currently designed to allow only one output type. However, the output type could be an object with multiple fields.
	if outputType != nil {
		expectedReturnCount += 1
	}
	if expectedReturnCount != returnCount {
		return fmt.Errorf("incorrect return count '%d'; expected '%d', with type(s) '%s'",
			returnCount, expectedReturnCount, getReturnTypeString(outputType, errorExpected))
	}

	// Validate error return
	if errorExpected {
		// Validate the last type as error
		handlerLastTypeName := parsedHandler.Type().Out(returnCount - 1).Name()
		if handlerLastTypeName != errorType {
			return fmt.Errorf("expected last return type from handler to be error, but instead found '%s'", handlerLastTypeName)
		}
	}

	// Validate the other return, if applicable
	if outputType != nil {
		// Validate the return type
		expectedType := outputType.ReflectedType()
		handlerType := parsedHandler.Type().Out(0)

		if expectedType != handlerType {
			return fmt.Errorf("mismatched return type. expected %s, handler has %s", expectedType, handlerType)
		}
	}
	return nil
}

// NewDynamicCallableFunction returns a CallableFunction for the dynamically typed function.
//
// - The input types must be specified and match, but you may use any types for instances when there are multiple allowed
// inputs. The return type of the handler should be any.
// - The handler function handles execution of the function.
// - The typeHandler function returns the output type given the input type. If the inputs are invalid, the
// handler should return an error.
func NewDynamicCallableFunction(
	id string,
	inputs []Type,
	display Display,
	handler any,
	typeHandler func(inputType []Type) (Type, error),
) (CallableFunction, error) {
	parsedHandler := reflect.ValueOf(handler)
	// Validate the input types match the provided ones.
	err := validateInputTypeCompatibility(inputs, parsedHandler)
	if err != nil {
		return nil, err
	}
	// Validate the output type
	returnCount := parsedHandler.Type().NumOut()

	switch {
	case returnCount != 2:
		return nil, fmt.Errorf("expected dynamic handler to have two returns, one with any type, and one with error type, but got %d return types", returnCount)
	case parsedHandler.Type().Out(1).Name() != errorType:
		return nil, fmt.Errorf("expected additional return type to be an error return, but got %s", parsedHandler.Type().Out(1).Name())
	case parsedHandler.Type().Out(0).Kind() != reflect.Interface:
		return nil, fmt.Errorf("expected 'any' return type for handler, but got %s", parsedHandler.Type().Out(0))
	}
	return &CallableFunctionSchema{
		IDValue:            id,
		InputsValue:        inputs,
		StaticOutputValue:  nil,
		OutputsError:       true,
		DisplayValue:       display,
		Handler:            parsedHandler,
		DynamicTypeHandler: typeHandler,
	}, nil
}

func validateInputTypeCompatibility(
	inputs []Type,
	handler reflect.Value,
) error {
	// Validate the input types match the provided ones.
	specifiedParams := len(inputs)
	actualParams := handler.Type().NumIn()
	if specifiedParams != actualParams {
		return fmt.Errorf(
			"parameter input counts do not match handler inputs. handler has %d, expected %d",
			actualParams, specifiedParams)
	}
	for i := 0; i < specifiedParams; i++ {
		expectedType := inputs[i].ReflectedType()
		handlerType := handler.Type().In(i)
		if expectedType != handlerType {
			return fmt.Errorf(
				"type mismatch for parameter at index %d. handler has %v, inputs schema at index %d specifies %v",
				i, handlerType, i, expectedType)
		}
	}
	return nil
}

type FunctionSchema struct {
	IDValue      string  `json:"id"`
	InputsValue  []Type  `json:"inputs"`
	OutputValue  Type    `json:"output"`
	DisplayValue Display `json:"display"`
}

func (f FunctionSchema) ID() string {
	return f.IDValue
}

func (f FunctionSchema) Parameters() []Type {
	return f.InputsValue
}

func (f FunctionSchema) ParameterTypeNames() []string {
	parameterNames := make([]string, len(f.Parameters()))
	for i := 0; i < len(f.Parameters()); i++ {
		parameterNames[i] = string(f.Parameters()[i].TypeID())
	}
	return parameterNames
}

func (f FunctionSchema) Output(_ []Type) (Type, error) {
	return f.OutputValue, nil
}

func (f FunctionSchema) Display() Display {
	return f.DisplayValue
}

func (f FunctionSchema) String() string {
	result := f.ID() + "(" + strings.Join(f.ParameterTypeNames(), ", ") + ") "
	result += getReturnTypeString(f.OutputValue, false)
	return result
}

func getReturnTypeString(returnType Type, hasError bool) string {
	switch {
	case returnType != nil:
		if hasError {
			return "(" + string(returnType.TypeID()) + ", error)"
		} else {
			return string(returnType.TypeID())
		}
	case hasError:
		return "error"
	default:
		return "void"
	}
}

type CallableFunctionSchema struct {
	IDValue     string `json:"id"`
	InputsValue []Type `json:"inputs"`
	// The output type when the output type does not change. Nil for void.
	StaticOutputValue Type    `json:"output"`
	OutputsError      bool    `json:"outputs_error"`
	DisplayValue      Display `json:"display"`
	// A callable function whose parameters (if any) match the type schema specified in InputsValue,
	// and whose return value type matches StaticOutputValue, the return type from DynamicTypeHandler,
	// or is void if both StaticOutputValue and DynamicTypeHandler are nil. An error return must be present
	// if OutputsError is true.
	// The handler may also return an error type.
	Handler reflect.Value
	// Returns the output type based on the input type. For advanced use cases. Cannot be void.
	DynamicTypeHandler func(inputType []Type) (Type, error)
}

func (f CallableFunctionSchema) ID() string {
	return f.IDValue
}

func (f CallableFunctionSchema) Parameters() []Type {
	return f.InputsValue
}

func (f CallableFunctionSchema) ParameterTypeNames() []string {
	parameterNames := make([]string, len(f.Parameters()))
	for i := 0; i < len(f.Parameters()); i++ {
		parameterNames[i] = string(f.Parameters()[i].TypeID())
	}
	return parameterNames
}

func (f CallableFunctionSchema) Output(inputType []Type) (Type, bool, error) {
	if f.DynamicTypeHandler == nil {
		return f.StaticOutputValue, f.OutputsError, nil
	} else {
		dynamicTypes, err := f.DynamicTypeHandler(inputType)
		return dynamicTypes, f.OutputsError, err
	}
}

func (f CallableFunctionSchema) Display() Display {
	return f.DisplayValue
}

func (f CallableFunctionSchema) String() string {
	result := f.ID() + "(" + strings.Join(f.ParameterTypeNames(), ", ") + ") "
	if f.DynamicTypeHandler != nil {
		result += "(dynamic, error)"
	} else {
		result += getReturnTypeString(f.StaticOutputValue, f.OutputsError)
	}
	return result
}

func (f CallableFunctionSchema) ToFunctionSchema() (*FunctionSchema, error) {
	if f.DynamicTypeHandler != nil {
		return nil, fmt.Errorf(
			"function '%s' cannot be represented as a FunctionSchema because function has dynamic typing",
			f.ID())
	}
	return &FunctionSchema{
		IDValue:      f.IDValue,
		InputsValue:  f.Parameters(),
		OutputValue:  f.StaticOutputValue,
		DisplayValue: f.DisplayValue,
	}, nil
}
func (f CallableFunctionSchema) Call(arguments []any) (any, error) {
	gotArgs := len(arguments)
	expectedArgs := f.Handler.Type().NumIn()
	if gotArgs != expectedArgs {
		return nil, NewFunctionCallError(fmt.Errorf(
			"incorrect number of args sent to function with ID '%s'. Expected %d, got %d",
			f.ID(),
			expectedArgs,
			gotArgs,
		), false)
	}
	// Convert to reflect values
	args := make([]reflect.Value, gotArgs)
	for i := 0; i < gotArgs; i++ {
		args[i] = reflect.ValueOf(arguments[i])
	}
	result := f.Handler.Call(args)
	gotReturns := len(result)
	expectedReturnVals := 0
	if f.StaticOutputValue != nil || f.DynamicTypeHandler != nil {
		expectedReturnVals = 1
	}
	// Validate return types
	switch {
	case expectedReturnVals == gotReturns:
		// Got expected return with no error return
		if expectedReturnVals == 0 {
			return nil, nil
		} else {
			return result[0].Interface(), nil
		}
	case expectedReturnVals+1 == gotReturns:
		errorVal := result[expectedReturnVals]
		if !errorVal.IsNil() {
			err, isError := errorVal.Interface().(error)
			if !isError {
				return nil, NewFunctionCallError(fmt.Errorf("error return val isn't an error '%w'", err), false)
			}
			if err != nil {
				return nil, NewFunctionCallError(err, true)
			}
		}
		// Expected return plus error return
		if expectedReturnVals == 0 {
			return nil, nil
		} else {
			return result[0].Interface(), nil
		}
	default:
		return nil, NewFunctionCallError(fmt.Errorf("unexpected return count. Expected %d or %d, got %d",
			expectedReturnVals, expectedReturnVals+1, gotReturns), false)
	}
}
