package schema

import (
	"fmt"
	"reflect"
)

type Function interface {
	ID() string
	Parameters() []Type
	Output([]Type) (Type, error)
	Display() Display
	String() string
}

type CallableFunction interface {
	Function
	ToFunctionSchema() (*FunctionSchema, error)
	Call(arguments []any) (any, error)
}

const errorType = "error"

// NewCallableFunction creates a CallableFunction schema type for the strictly typed function.
//
// - The handler types must match the input and output types specified.
// - The return type of the handler is determined by the value specified for output. If output is nil, it's a void
// function that may have no return, or a single error return.
// - If output is not nil, the return type must match, plus it may have an error return type, too.
func NewCallableFunction(
	id string,
	inputs []Type,
	output Type,
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
	if output == nil {
		err := validateVoidFunc(parsedHandler)
		if err != nil {
			return nil, err
		}
	} else {
		err := validateTypedReturnFunc(parsedHandler, output)
		if err != nil {
			return nil, err
		}
	}
	return &CallableFunctionSchema{
		IDValue:            id,
		InputsValue:        inputs,
		DefaultOutputValue: output,
		DisplayValue:       display,
		Handler:            parsedHandler,
	}, nil
}

func validateVoidFunc(parsedHandler reflect.Value) error {
	returnCount := parsedHandler.Type().NumOut()
	// A void function can have no returns or an error return
	if returnCount > 1 {
		return fmt.Errorf(
			"parameter output is nil, meaning it's a void function, or a function with just an error return, but got %d return types",
			returnCount,
		)
	} else if returnCount == 1 {
		// Validate that it's just an error return
		returnTypeName := parsedHandler.Type().Out(0).Name()
		if returnTypeName != errorType {
			return fmt.Errorf("expected void or error return, but got %s", returnTypeName)
		}
	}
	return nil
}

func validateTypedReturnFunc(parsedHandler reflect.Value, outputType Type) error {
	returnCount := parsedHandler.Type().NumOut()

	if returnCount > 2 || returnCount < 1 {
		return fmt.Errorf("expected handler to have one return, or one plus an error return, but got %d return types", returnCount)
	} else {
		// Validate the return type
		expectedType := outputType.ReflectedType()
		handlerType := parsedHandler.Type().Out(0)
		if expectedType != handlerType {
			return fmt.Errorf("mismatched return type. expected %s, handler has %s", expectedType, handlerType)
		}
		// Validate error return, if applicable.
		if returnCount == 2 && parsedHandler.Type().Out(1).Name() != errorType {
			return fmt.Errorf("expected additional return type to be an error return, but got %s", parsedHandler.Type().Out(1).Name())
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
		return nil, fmt.Errorf("expected dynamic handler to have two returns, one any and one error, but got %d return types", returnCount)
	case parsedHandler.Type().Out(1).Name() != errorType:
		return nil, fmt.Errorf("expected additional return type to be an error return, but got %s", parsedHandler.Type().Out(1).Name())
	case parsedHandler.Type().Out(0).Kind() != reflect.Interface:
		return nil, fmt.Errorf("expected 'any' return type for handler, but got %s", parsedHandler.Type().Out(0))
	}
	return &CallableFunctionSchema{
		IDValue:            id,
		InputsValue:        inputs,
		DefaultOutputValue: nil,
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
			"parameter inputs do not match handler inputs. handler has %d, expected %d",
			actualParams, specifiedParams)
	}
	for i := 0; i < len(inputs); i++ {
		expectedType := inputs[i].ReflectedType()
		handlerType := handler.Type().In(i)
		if expectedType != handlerType {
			return fmt.Errorf(
				"type mismatch for parameter at index %d. handler has %v, inputs specifies %v",
				i, handlerType, expectedType)
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

func (f FunctionSchema) Output(_ []Type) (Type, error) {
	return f.OutputValue, nil
}

func (f FunctionSchema) Display() Display {
	return f.DisplayValue
}

func (f FunctionSchema) String() string {
	result := f.ID() + "("
	for i := 0; i < len(f.Parameters()); i++ {
		if i != 0 {
			result += ", "
		}
		result += string(f.Parameters()[i].TypeID())
	}
	result += ") "
	if f.OutputValue != nil {
		result += string(f.OutputValue.TypeID())
	} else {
		result += "void"
	}
	return result
}

type CallableFunctionSchema struct {
	IDValue            string  `json:"id"`
	InputsValue        []Type  `json:"inputs"`
	DefaultOutputValue Type    `json:"output"`
	DisplayValue       Display `json:"display"`
	// Should be a function call with any amount of parameters, and 0, 1 (err or data),
	// or two return types (err and data).
	// Params should match types in InputsValue.
	// Return types should match OutputValue, or not be there if nil. Plus may have one addition return type: error.
	Handler reflect.Value
	// Returns the output type based on the input type. For advanced use cases. Cannot be void.
	DynamicTypeHandler func(inputType []Type) (Type, error)
}

func (s CallableFunctionSchema) ID() string {
	return s.IDValue
}

func (s CallableFunctionSchema) Parameters() []Type {
	return s.InputsValue
}

func (s CallableFunctionSchema) Output(inputType []Type) (Type, error) {
	if s.DynamicTypeHandler == nil {
		return s.DefaultOutputValue, nil
	} else {
		return s.DynamicTypeHandler(inputType)
	}
}

func (s CallableFunctionSchema) Display() Display {
	return s.DisplayValue
}

func (f CallableFunctionSchema) String() string {
	result := f.ID() + "("
	for i := 0; i < len(f.Parameters()); i++ {
		if i != 0 {
			result += ", "
		}
		result += string(f.Parameters()[i].TypeID())
	}
	result += ") "
	switch {
	case f.DynamicTypeHandler != nil:
		result += "dynamic"
	case f.DefaultOutputValue != nil:
		result += string(f.DefaultOutputValue.TypeID())
	default:
		result += "void"
	}
	return result
}

func (s CallableFunctionSchema) ToFunctionSchema() (*FunctionSchema, error) {
	if s.DynamicTypeHandler != nil && s.DefaultOutputValue == nil {
		return nil, fmt.Errorf(
			"function '%s' cannot be represented as a FunctionSchema because function has dynamic typing",
			s.ID())
	}
	return &FunctionSchema{
		IDValue:      s.IDValue,
		InputsValue:  s.Parameters(),
		OutputValue:  s.DefaultOutputValue,
		DisplayValue: s.DisplayValue,
	}, nil
}
func (s CallableFunctionSchema) Call(arguments []any) (any, error) {
	gotArgs := len(arguments)
	expectedArgs := s.Handler.Type().NumIn()
	if gotArgs != expectedArgs {
		return nil, fmt.Errorf(
			"incorrect number of args sent to function with ID '%s'. Expected %d, got %d",
			s.ID(),
			expectedArgs,
			gotArgs,
		)
	}
	// Convert to reflect values
	args := make([]reflect.Value, gotArgs)
	for i := 0; i < gotArgs; i++ {
		args[i] = reflect.ValueOf(arguments[i])
	}
	result := s.Handler.Call(args)
	gotReturns := len(result)
	expectedReturnVals := 0
	if s.DefaultOutputValue != nil || s.DynamicTypeHandler != nil {
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
				return nil, fmt.Errorf("error return val isn't an error '%w'", err)
			}
			if err != nil {
				return nil, fmt.Errorf("function returned error: %w", err)
			}
		}
		// Expected return plus error return
		if expectedReturnVals == 0 {
			return nil, nil
		} else {
			return result[0].Interface(), nil
		}
	default:
		return nil, fmt.Errorf("unexpected return count. Expected %d or %d, got %d",
			expectedReturnVals, expectedReturnVals+1, gotReturns)
	}
}
