package schema

type Input struct {
	// id identifies the step, signal, or any other case where data is being input
	ID string
	// The data being input into the step/signal/other
	InputData any
}
