package atp

import (
	"github.com/fxamacker/cbor/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
)

const ProtocolVersion int64 = 2

type HelloMessage struct {
	Version int64 `cbor:"version"`
	Schema  any   `cbor:"schema"`
}

type StartWorkMessage struct {
	StepID string `cbor:"id"`
	Config any    `cbor:"config"`
}

// All messages that can be contained in a RuntimeMessage struct.
const (
	MessageTypeWorkDone uint32 = 1
	MessageTypeSignal   uint32 = 2
)

type RuntimeMessage struct {
	MessageID   uint32 `cbor:"id"`
	MessageData any    `cbor:"data"`
}

type DecodedRuntimeMessage struct {
	MessageID      uint32          `cbor:"id"`
	RawMessageData cbor.RawMessage `cbor:"data"`
}

type workDoneMessage struct {
	OutputID   string `cbor:"output_id"`
	OutputData any    `cbor:"output_data"`
	DebugLogs  string `cbor:"debug_logs"`
}

type signalMessage struct {
	StepID   string `cbor:"step_id"`
	SignalID string `cbor:"signal_id"`
	Data     any    `cbor:"data"`
}

func (s signalMessage) ToInput() schema.Input {
	return schema.Input{ID: s.SignalID, InputData: s.Data}
}
