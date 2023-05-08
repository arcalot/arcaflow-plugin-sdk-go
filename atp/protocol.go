package atp

type HelloMessage struct {
	Version int64 `cbor:"version"`
	Schema  any   `cbor:"schema"`
}

type StartWorkMessage struct {
	StepID string `cbor:"id"`
	Config any    `cbor:"config"`
}

type workDoneMessage struct {
	OutputID   string `cbor:"output_id"`
	OutputData any    `cbor:"output_data"`
	DebugLogs  string `cbor:"debug_logs"`
}
