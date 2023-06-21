package atp

import "github.com/fxamacker/cbor/v2"

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
	MessageTypeSignal          = 2
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
	SignalID uint32 `cbor:"id"`
	Data     any    `cbor:"data"`
}
