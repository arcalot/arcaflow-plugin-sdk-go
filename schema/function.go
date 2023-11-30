package schema

import (
	"fmt"
	"reflect"
)

type Function interface {
	ID() string
	Parameters() []Type
	Output() Type
	Display() Display
}

type CallableFunction interface {
	Function
	ToFunctionSchema() *FunctionSchema
	Call(arguments []any) (any, error)
}

func NewCallableFunction(
	id string,
	inputs []Type,
	output Type,
	display Display,
	handler any,
) (CallableFunction, error) {
	parsedHandler := reflect.ValueOf(handler)
	// Validate the input types match the provided ones.
	specifiedParams := len(inputs)
	actualParams := parsedHandler.Type().NumIn()
	if specifiedParams != actualParams {
		return nil, fmt.Errorf(
			"parameter inputs do not match handler inputs. handler has %d, expected %d",
			actualParams, specifiedParams)
	}
	for i := 0; i < len(inputs); i++ {
		expectedType := inputs[i].ReflectedType()
		handlerType := parsedHandler.Type().In(i)
		if expectedType != handlerType {
			return nil, fmt.Errorf(
				"type mismatch for parameter at index %d. handler has %v, inputs specifies %v",
				i, handlerType, expectedType)
		}
	}
	// Validate the output type
	returnCount := parsedHandler.Type().NumOut()
	if output == nil {
		if returnCount > 1 {
			return nil, fmt.Errorf("parameter output is nil, meaning it's a void function, or a function with just an error return, but got %d return types", returnCount)
		} else if returnCount == 1 {
			// Validate that it's just an error return
			returnTypeName := parsedHandler.Type().Out(0).Name()
			if returnTypeName != "error" {
				return nil, fmt.Errorf("expected void or error return, but got %s", returnTypeName)
			}
		}
	} else {
		if returnCount > 2 || returnCount < 1 {
			return nil, fmt.Errorf("expected handler to have one return, or one plus an error return, but got %d return types", returnCount)
		} else {
			// Validate the return type
			expectedType := output.ReflectedType()
			handlerType := parsedHandler.Type().Out(0)
			if expectedType != handlerType {
				return nil, fmt.Errorf("mismatched return type. expected %s, handler has %s", expectedType, handlerType)
			}
			// Validate error return, if applicable.
			if returnCount == 2 && parsedHandler.Type().Out(1).Name() != "error" {
				return nil, fmt.Errorf("expected additional return type to be an error return, but got %s", parsedHandler.Type().Out(1).Name())
			}
		}
	}
	return &CallableFunctionSchema{
		IDValue:      id,
		InputsValue:  inputs,
		OutputValue:  output,
		DisplayValue: display,
		Handler:      parsedHandler,
	}, nil
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

func (f FunctionSchema) Output() Type {
	return f.OutputValue
}

func (f FunctionSchema) Display() Display {
	return f.DisplayValue
}

type CallableFunctionSchema struct {
	IDValue      string  `json:"id"`
	InputsValue  []Type  `json:"inputs"`
	OutputValue  Type    `json:"output"`
	DisplayValue Display `json:"display"`
	// Should be a function call with any amount of parameters, and 0, 1 (err or data),
	// or two return types (err and data).
	// Params should match types in InputsValue.
	// Return types should match OutputValue, or not be there if nil. Plus may have one addition return type: error.
	Handler reflect.Value
}

func (s CallableFunctionSchema) ID() string {
	return s.IDValue
}
func (s CallableFunctionSchema) Parameters() []Type {
	return s.InputsValue
}
func (s CallableFunctionSchema) Output() Type {
	return s.OutputValue
}
func (s CallableFunctionSchema) Display() Display {
	return s.DisplayValue
}
func (s CallableFunctionSchema) ToFunctionSchema() *FunctionSchema {
	return &FunctionSchema{
		IDValue:      s.IDValue,
		InputsValue:  s.Parameters(),
		OutputValue:  s.OutputValue,
		DisplayValue: s.DisplayValue,
	}
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
	if s.Output() != nil {
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
