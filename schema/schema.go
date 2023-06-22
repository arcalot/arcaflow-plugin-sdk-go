package schema

import (
	"context"
	"fmt"
)

// Schema is a collection of steps supported by a plugin.
type Schema[S Step, I Signal] interface {
	Steps() map[string]S
	ListeningSignals() map[string]I
	EmittingSignals() map[string]I

	SelfSerialize() (any, error)
}

//
//// NewSchema builds a new schema with the specified steps.
//func NewSchema(
//	steps map[string]*StepSchema,
//	listeningSignals map[string]*SignalSchema,
//	emittingSignals map[string]*SignalSchema,
//) Schema[Step, Signal] {
//	return &SchemaSchema{
//		steps,
//		listeningSignals,
//		emittingSignals,
//	}
//}

type SchemaSchema struct {
	StepsValue            map[string]*StepSchema   `json:"steps"`
	ListeningSignalsValue map[string]*SignalSchema `json:"listening_signals"`
	EmittingSignalsValue  map[string]*SignalSchema `json:"emitting_signals"`
}

func (s SchemaSchema) SelfSerialize() (any, error) {
	steps := make(map[string]*StepSchema, len(s.StepsValue))
	listeningSignals := make(map[string]*SignalSchema, len(s.ListeningSignalsValue))
	emittingSignals := make(map[string]*SignalSchema, len(s.EmittingSignalsValue))

	for id, step := range s.StepsValue {
		steps[id] = step
	}

	return schemaSchema.Serialize(&SchemaSchema{
		steps,
		listeningSignals,
		emittingSignals,
	})
}

func (s SchemaSchema) Steps() map[string]Step {
	result := make(map[string]Step, len(s.StepsValue))
	for k, v := range s.StepsValue {
		result[k] = v
	}
	return result
}

func (s SchemaSchema) applyScope() {
	for _, step := range s.StepsValue {
		// We can apply an empty scope because the scope does not need another scope.
		step.InputValue.ApplyScope(nil)
		for _, output := range step.OutputsValue {
			output.ApplyScope(nil)
		}
	}
}

func NewPluginCallableSchema(
	steps []CallableStep,
	listening_signals []CallableSignal,
	emitting_signals []SignalSchema,
) *CallablePluginSchema {
	stepMap := make(map[string]CallableStep, len(steps))
	for _, s := range steps {
		stepMap[s.ID()] = s
	}
	listeningSignalMap := make(map[string]CallableSignal, len(listening_signals))
	for _, s := range listening_signals {
		listeningSignalMap[s.ID()] = s
	}
	emittingSignalMap := make(map[string]SignalSchema, len(emitting_signals))
	for _, s := range emitting_signals {
		emittingSignalMap[s.ID()] = s
	}

	return &CallablePluginSchema{
		stepMap,
		listeningSignalMap,
		emittingSignalMap,
	}
}

type CallablePluginSchema struct {
	StepsValue     map[string]CallableStep   `json:"steps"`
	SignalHandlers map[string]CallableSignal `json:"signal_handlers"`
	SignalEmitters map[string]SignalSchema   `json:"signal_emitters"`
}

func (s CallablePluginSchema) CallStep(
	ctx context.Context,
	stepID string,
	serializedInputData any,
) (
	outputID string,
	serializedOutputData any,
	err error,
) {
	step, ok := s.StepsValue[stepID]
	if !ok {
		return "", nil, BadArgumentError{
			Message: fmt.Sprintf("Invalid step called: %s", stepID),
		}
	}
	unserializedInputData, err := step.Input().Unserialize(serializedInputData)
	if err != nil {
		return "", nil, InvalidInputError{err}
	}
	outputID, unserializedOutput, err := step.Call(ctx, unserializedInputData)
	if err != nil {
		return outputID, nil, err
	}
	output := step.Outputs()[outputID]
	serializedData, err := output.Schema().Serialize(unserializedOutput)
	if err != nil {
		return "", nil, InvalidOutputError{err}
	}
	return outputID, serializedData, nil
}

func (s CallablePluginSchema) CallSignal(
	ctx context.Context,
	signalID string,
	serializedInputData any,
) (
	err error,
) {
	signal, ok := s.SignalHandlers[signalID]
	if !ok {
		return BadArgumentError{
			Message: fmt.Sprintf("Invalid signal called: %s", signalID),
		}
	}
	unserializedInputData, err := signal.DataSchema().Unserialize(serializedInputData)
	if err != nil {
		return InvalidInputError{err}
	}
	err = signal.Call(ctx, unserializedInputData)
	if err != nil {
		return err
	}
	return nil
}

func (s CallablePluginSchema) SelfSerialize() (any, error) {
	steps := make(map[string]*StepSchema, len(s.StepsValue))
	receivedSignals := make(map[string]*SignalSchema, len(s.SignalHandlers))
	emittedSignals := make(map[string]*SignalSchema, len(s.SignalEmitters))

	for id, step := range s.StepsValue {
		steps[id] = step.ToStepSchema()
	}

	for id, signal := range s.SignalEmitters {
		emittedSignals[id] = &signal
	}
	for id, signal := range s.SignalHandlers {
		receivedSignals[id] = signal.ToSignalSchema()
	}

	return schemaSchema.Serialize(&SchemaSchema{
		steps,
		receivedSignals,
		emittedSignals,
	})
}
