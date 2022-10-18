package atp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"go.arcalot.io/log"
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
	// Execute executes a step with a given context and returns the resulting output.
	Execute(ctx context.Context, stepID string, input any) (outputID string, outputData any, err error)
}

// NewClient creates a new ATP client (part of the engine code).
func NewClient(
	channel ClientChannel,
) Client {
	return NewClientWithLogger(channel, nil)
}

// NewClientWithLogger creates a new ATP client (part of the engine code) with a logger.
func NewClientWithLogger(
	channel ClientChannel,
	logger log.Logger,
) Client {
	decMode, err := cbor.DecOptions{
		ExtraReturnErrors: cbor.ExtraDecErrorUnknownField,
	}.DecMode()
	if err != nil {
		panic(err)
	}
	if logger == nil {
		logger = log.NewLogger(log.LevelDebug, log.NewNOOPLogger())
	}
	return &client{
		channel,
		decMode,
		logger,
	}
}

type client struct {
	channel ClientChannel
	decMode cbor.DecMode
	logger  log.Logger
}

func (c *client) ReadSchema() (schema.Schema[schema.Step], error) {
	c.logger.Debugf("Reading plugin schema...")
	cborReader := c.decMode.NewDecoder(c.channel)

	var hello helloMessage
	if err := cborReader.Decode(&hello); err != nil {
		c.logger.Errorf("Failed to decode ATP hello message: %v", err)
		return nil, fmt.Errorf("failed to decode hello message (%w)", err)
	}
	c.logger.Debugf("Hello message read, ATP version %d.", hello.Version)

	if hello.Version != 1 {
		c.logger.Errorf("Incompatible plugin ATP version: %d", hello.Version)
		return nil, fmt.Errorf("Incompatible plugin ATP version: %d", hello.Version)
	}

	unserializedSchema, err := schema.UnserializeSchema(hello.Schema)
	if err != nil {
		c.logger.Errorf("Invalid schema received from plugion: %v", err)
		return nil, fmt.Errorf("invalid schema (%w)", err)
	}
	c.logger.Debugf("Schema unserialization complete.")
	return unserializedSchema, nil
}

func (c client) Execute(ctx context.Context, stepID string, input any) (outputID string, outputData any, err error) {
	c.logger.Debugf("Executing step %s...", stepID)
	cborWriter := cbor.NewEncoder(c.channel)
	if err := cborWriter.Encode(startWorkMessage{
		StepID: stepID,
		Config: input,
	}); err != nil {
		c.logger.Errorf("Step %s failed to write start work message: %v", stepID, err)
		return "", nil, fmt.Errorf("failed to write work start message (%w)", err)
	}
	c.logger.Debugf("Step %s started, waiting for response...", stepID)

	cborReader := c.decMode.NewDecoder(c.channel)
	var doneMessage workDoneMessage
	if err := cborReader.Decode(&doneMessage); err != nil {
		c.logger.Errorf("Step %s failed to read work done message: %v", stepID, err)
		return "", nil, fmt.Errorf("failed to read work done message (%w)", err)
	}
	c.logger.Debugf("Step %s completed with output ID '%s'.", stepID, doneMessage.OutputID)
	debugLogs := strings.Split(doneMessage.DebugLogs, "\n")
	for _, line := range debugLogs {
		if strings.TrimSpace(line) != "" {
			c.logger.Debugf("Step %s debug: %s", stepID, line)
		}
	}

	return doneMessage.OutputID, doneMessage.OutputData, nil
}
