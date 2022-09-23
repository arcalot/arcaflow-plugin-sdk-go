package atp

import (
	"context"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
)

// ClientChannel holds the methods to talking to an ATP server (plugin).
type ClientChannel interface {
	io.Reader
	io.Writer
	io.Closer
}

// Client is the way to read information from the ATP server and then send a task to it. A task can only be sent
// once.
type Client interface {
	// ReadSchema reads the schema from the ATP server.
	ReadSchema() (schema.Schema[schema.Step], error)
	// Execute executes a stepo with a given context and returns the resulting output.
	Execute(ctx context.Context, stepID string, input any) (outputID string, outputData any, debugLogs string)
}

func NewClient(
	channel ClientChannel,
) Client {
	return &client{
		channel,
	}
}

type client struct {
	channel ClientChannel
}

func (c *client) ReadSchema() (schema.Schema[schema.Step], error) {
	cborReader := cbor.NewDecoder(c.channel)

	var hello helloMessage
	if err := cborReader.Decode(&hello); err != nil {
		return nil, fmt.Errorf("failed to decode hello message (%w)", err)
	}

	if hello.Version != 1 {
		return nil, fmt.Errorf("Incompatible ATP client version: %d", hello.Version)
	}

	unserializedSchema, err := schema.UnserializeSchema(hello.Schema)
	if err != nil {
		return nil, fmt.Errorf("Client sent an invalid schema (%w)", err)
	}
	return unserializedSchema, nil
}

func (c client) Execute(ctx context.Context, stepID string, input any) (outputID string, outputData any, debugLogs string) {
	cborWriter := cbor.NewEncoder(c.channel)
	if err := cborWriter.Encode(startWorkMessage{
		StepID: stepID,
		Config: input,
	}); err != nil {
		panic(fmt.Errorf("failed to write work start message (%w)", err))
	}

	cborReader := cbor.NewDecoder(c.channel)
	var doneMessage workDoneMessage
	if err := cborReader.Decode(&doneMessage); err != nil {
		panic(fmt.Errorf("failed to read work done message (%w)", err))
	}

	return doneMessage.OutputID, doneMessage.OutputData, doneMessage.DebugLogs
}
