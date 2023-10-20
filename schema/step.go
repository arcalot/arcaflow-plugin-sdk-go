package schema

import (
	"context"
	"fmt"
	"sync"
)

// Step holds the definition for a single step, it's input and output definitions.
type Step interface {
	ID() string
	Input() Scope
	Outputs() map[string]*StepOutputSchema
	SignalHandlers() map[string]*SignalSchema
	SignalEmitters() map[string]*SignalSchema
	Display() Display
}

// CallableStep is a step that can be directly called.
type CallableStep interface {
	Step
	ToStepSchema() *StepSchema
	Call(ctx context.Context, runID string, data any) (outputID string, outputData any, err error)
	CallSignal(ctx context.Context, runID string, signalID string, data any) (err error)
}

// NewStepSchema defines a new step.
func NewStepSchema(
	id string,
	input Scope,
	outputs map[string]*StepOutputSchema,
	signalHandlers map[string]*SignalSchema,
	signalEmitters map[string]*SignalSchema,
	display Display,
) *StepSchema {
	return &StepSchema{
		id,
		input,
		outputs,
		signalHandlers,
		signalEmitters,
		display,
	}
}

// StepSchema describes a single step in a schema to execute one task. It has a fixed input and one or more outputs,
// denominated by a string output ID.
type StepSchema struct {
	IDValue             string                       `json:"id"`
	InputValue          Scope                        `json:"input"`
	OutputsValue        map[string]*StepOutputSchema `json:"outputs"`
	SignalHandlersValue map[string]*SignalSchema     `json:"signal_handlers"`
	SignalEmittersValue map[string]*SignalSchema     `json:"signal_emitters"`
	DisplayValue        Display                      `json:"display"`
}

func (s StepSchema) ID() string {
	return s.IDValue
}

func (s StepSchema) Input() Scope {
	return s.InputValue
}

func (s StepSchema) Outputs() map[string]*StepOutputSchema {
	return s.OutputsValue
}

func (s StepSchema) SignalHandlers() map[string]*SignalSchema {
	return s.SignalHandlersValue
}

func (s StepSchema) SignalEmitters() map[string]*SignalSchema {
	return s.SignalEmittersValue
}

func (s StepSchema) Display() Display {
	return s.DisplayValue
}

// NewCallableStep creates a callable step definition.
func NewCallableStep[StepInputType any](
	id string,
	input *ScopeSchema,
	outputs map[string]*StepOutputSchema,
	display Display,
	handler func(context.Context, StepInputType) (string, any),
) CallableStep {
	updatedHandler := func(ctx context.Context, _ any, step StepInputType) (string, any) {
		return handler(ctx, step)
	}
	return &CallableStepSchema[any, StepInputType]{
		IDValue:             id,
		InputValue:          input,
		OutputsValue:        outputs,
		SignalHandlersValue: nil,
		SignalEmittersValue: nil,
		DisplayValue:        display,
		initializer:         nil,
		stepData:            make(map[string]*runningStepData[any]),
		handler:             updatedHandler,
	}
}

// NewCallableStepWithSignals creates a callable step definition, and allows the
// inclusion of signal handlers and emitters.
func NewCallableStepWithSignals[StepData any, StepInputType any](
	id string,
	input *ScopeSchema,
	outputs map[string]*StepOutputSchema,
	signalHandlers map[string]CallableSignal,
	signalEmitters map[string]*SignalSchema,
	display Display,
	initializer func() StepData,
	handler func(context.Context, StepData, StepInputType) (string, any),
) CallableStep {
	wg := sync.WaitGroup{}
	if initializer != nil {
		wg.Add(1)
	}
	return &CallableStepSchema[StepData, StepInputType]{
		IDValue:             id,
		InputValue:          input,
		OutputsValue:        outputs,
		SignalHandlersValue: signalHandlers,
		SignalEmittersValue: signalEmitters,
		DisplayValue:        display,
		initializer:         initializer,
		handler:             handler,
		stepData:            make(map[string]*runningStepData[StepData]),
	}
}

type runningStepData[StepData any] struct {
	runID           string
	initializedData StepData
	startedWG       *sync.WaitGroup // For waiting until the step is started.
	done            bool
}

// CallableStepSchema is a step that can be directly called and is typed to a specific input type.
type CallableStepSchema[StepData any, InputType any] struct {
	IDValue             string                       `json:"id"`
	InputValue          *ScopeSchema                 `json:"input"`
	SignalHandlersValue map[string]CallableSignal    `json:"signal_handlers"`
	SignalEmittersValue map[string]*SignalSchema     `json:"signal_emitters"`
	OutputsValue        map[string]*StepOutputSchema `json:"outputs"`
	DisplayValue        Display                      `json:"display"`
	initializer         func() StepData
	initializerMutex    sync.Mutex
	stepData            map[string]*runningStepData[StepData] // Maps run ID to step data
	handler             func(context.Context, StepData, InputType) (string, any)
}

func (s *CallableStepSchema[StepData, InputType]) SignalHandlers() map[string]*SignalSchema {
	handlers := map[string]*SignalSchema{}
	for key, handler := range s.SignalHandlersValue {
		handlers[key] = handler.ToSignalSchema()
	}
	return handlers
}

func (s *CallableStepSchema[StepData, InputType]) SignalEmitters() map[string]*SignalSchema {
	return s.SignalEmittersValue
}

func (s *CallableStepSchema[StepData, InputType]) ID() string {
	return s.IDValue
}

func (s *CallableStepSchema[StepData, InputType]) Input() Scope {
	return s.InputValue
}

func (s *CallableStepSchema[StepData, InputType]) Outputs() map[string]*StepOutputSchema {
	return s.OutputsValue
}

func (s *CallableStepSchema[StepData, InputType]) Display() Display {
	return s.DisplayValue
}

func (s *CallableStepSchema[StepData, InputType]) ToStepSchema() *StepSchema {
	signalHandlers := make(map[string]*SignalSchema, len(s.SignalHandlersValue))
	for k, v := range s.SignalHandlersValue {
		signalHandlers[k] = v.ToSignalSchema()
	}
	return &StepSchema{
		IDValue:             s.IDValue,
		InputValue:          s.InputValue,
		OutputsValue:        s.OutputsValue,
		SignalHandlersValue: signalHandlers,
		SignalEmittersValue: s.SignalEmittersValue,
		DisplayValue:        s.DisplayValue,
	}
}

// Set up the runningStepData struct. This results in a waitgroup available for the signals to wait on, and
// it setups the data shared between the step and signals.
func (s *CallableStepSchema[StepData, InputType]) setupStepData(runID string) *runningStepData[StepData] {
	s.initializerMutex.Lock()
	defer s.initializerMutex.Unlock()

	// This will be called by both the signal and step handlers, so it's important to check to ensure this
	// isn't getting re-done on the second call.
	existingRunningStepData, found := s.stepData[runID]
	if found {
		return existingRunningStepData // Already done
	}
	var stepData StepData
	if s.initializer != nil {
		newInitializedData := s.initializer()
		stepData = newInitializedData
	}
	runningStepData := runningStepData[StepData]{
		runID:           runID,
		initializedData: stepData,
		startedWG:       &sync.WaitGroup{},
		done:            false,
	}
	s.stepData[runID] = &runningStepData
	return &runningStepData
}

func (s *CallableStepSchema[StepData, InputType]) Call(ctx context.Context, runID string, input any) (string, any, error) {
	if err := s.InputValue.Validate(input); err != nil {
		return "", nil, InvalidInputError{err}
	}

	runningStepData := s.setupStepData(runID)
	outputID, outputData := s.handler(ctx, runningStepData.initializedData, input.(InputType))
	output, ok := s.OutputsValue[outputID]
	if !ok {
		return "", nil, InvalidOutputError{
			fmt.Errorf("undeclared output ID: %s", outputID),
		}
	}
	return outputID, outputData, output.Validate(outputData)
}

func (s *CallableStepSchema[StepData, InputType]) CallSignal(
	ctx context.Context,
	runID string,
	signalID string,
	input any,
) error {
	runningStepData := s.setupStepData(runID)
	runningStepData.startedWG.Wait() // Wait for the step to start
	return s.SignalHandlersValue[signalID].Call(ctx, runningStepData.initializedData, input)
}
