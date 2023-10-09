package atp

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
)

const ProtocolVersion int64 = 2

type HelloMessage struct {
	Version int64 `cbor:"version"`
	Schema  any   `cbor:"schema"`
}

type WorkStartMessage struct {
	StepID string `cbor:"id"`
	Config any    `cbor:"config"`
}

// All messages that can be contained in a RuntimeMessage struct.
const (
	MessageTypeWorkStart  uint32 = 1
	MessageTypeWorkDone   uint32 = 2
	MessageTypeSignal     uint32 = 3
	MessageTypeClientDone uint32 = 4
	MessageTypeError      uint32 = 5
)

type RuntimeMessage struct {
	MessageID   uint32 `cbor:"id"`
	RunID       string `cbor:"run_id"`
	MessageData any    `cbor:"data"`
}

type DecodedRuntimeMessage struct {
	MessageID      uint32          `cbor:"id"`
	RunID          string          `cbor:"run_id"`
	RawMessageData cbor.RawMessage `cbor:"data"`
}

type WorkDoneMessage struct {
	StepID     string `cbor:"step_id"`
	OutputID   string `cbor:"output_id"`
	OutputData any    `cbor:"output_data"`
	DebugLogs  string `cbor:"debug_logs"`
}

type SignalMessage struct {
	SignalID string `cbor:"signal_id"`
	Data     any    `cbor:"data"`
}

func (s SignalMessage) ToInput(runID string) schema.Input {
	return schema.Input{RunID: runID, ID: s.SignalID, InputData: s.Data}
}

type clientDoneMessage struct {
	// Empty for now.
}

type ErrorMessage struct {
	Error       string `cbor:"error"`
	StepFatal   bool   `cbor:"step_fatal"`
	ServerFatal bool   `cbor:"server_fatal"`
}

func (e ErrorMessage) ToString(runID string) string {
	return fmt.Sprintf("RunID: %s, err: %s, step fatal: %t, server fatal: %t", runID, e.Error, e.StepFatal, e.ServerFatal)
}
