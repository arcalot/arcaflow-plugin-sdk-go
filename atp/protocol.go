package atp

type helloMessage struct {
	Version int64 `json:"version"`
	Schema  any   `json:"schema"`
}

type startWorkMessage struct {
	StepID string `json:"id"`
	Config any    `json:"config"`
}

type workDoneMessage struct {
	OutputID   string `json:"output_id"`
	OutputData any    `json:"output_data"`
	DebugLogs  string `json:"debug_logs"`
}
