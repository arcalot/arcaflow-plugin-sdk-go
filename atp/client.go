package atp

import (
	"fmt"
	"io"
	"strings"

	"github.com/fxamacker/cbor/v2"
	log "go.arcalot.io/log/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
)

const MinSupportedATPVersion = 1
const MaxSupportedATPVersion = 2

// ClientChannel holds the methods to talking to an ATP server (plugin).
type ClientChannel interface {
	io.Reader
	io.Writer
	io.Closer
}

// Client is the way to read information from the ATP server and then send a task to it in the form of a step.
// A step can only be sent once, but signals can be sent until the step is over.
type Client interface {
	// ReadSchema reads the schema from the ATP server.
	ReadSchema() (*schema.SchemaSchema, error)
	// Execute executes a step with a given context and returns the resulting output.
	Execute(stepID string, input any) (outputID string, outputData any, err error)
	Encoder() *cbor.Encoder
	Decoder() *cbor.Decoder
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
		-1, // unknown
		channel,
		decMode,
		logger,
		decMode.NewDecoder(channel),
		cbor.NewEncoder(channel),
	}
}

func (c *client) Decoder() *cbor.Decoder {
	return c.decoder
}

func (c *client) Encoder() *cbor.Encoder {
	return c.encoder
}

type client struct {
	atpVersion int64 // TODO: Should this be persisted here or returned?
	channel    ClientChannel
	decMode    cbor.DecMode
	logger     log.Logger
	decoder    *cbor.Decoder
	encoder    *cbor.Encoder
}

func (c *client) ReadSchema() (*schema.SchemaSchema, error) {
	c.logger.Debugf("Reading plugin schema...")

	if err := c.encoder.Encode(nil); err != nil {
		c.logger.Errorf("Failed to encode ATP start output message: %v", err)
		return nil, fmt.Errorf("failed to encode start output message (%w)", err)
	}

	var hello HelloMessage
	if err := c.decoder.Decode(&hello); err != nil {
		c.logger.Errorf("Failed to decode ATP hello message: %v", err)
		return nil, fmt.Errorf("failed to decode hello message (%w)", err)
	}
	c.logger.Debugf("Hello message read, ATP version %d.", hello.Version)

	if hello.Version < MinSupportedATPVersion || hello.Version > MaxSupportedATPVersion {
		c.logger.Errorf("Incompatible plugin ATP version: %d; expected between %d and %d.", hello.Version,
			MinSupportedATPVersion, MaxSupportedATPVersion)
		return nil, fmt.Errorf("incompatible plugin ATP version: %d; expected between %d and %d", hello.Version,
			MinSupportedATPVersion, MaxSupportedATPVersion)
	}
	c.atpVersion = hello.Version

	unserializedSchema, err := schema.UnserializeSchema(hello.Schema)
	if err != nil {
		c.logger.Errorf("Invalid schema received from plugin: %v", err)
		return nil, fmt.Errorf("invalid schema (%w)", err)
	}
	c.logger.Debugf("Schema unserialization complete.")
	return unserializedSchema, nil
}

func (c *client) Execute(stepID string, input any) (outputID string, outputData any, err error) {
	c.logger.Debugf("Executing step %s...", stepID)
	if err := c.encoder.Encode(StartWorkMessage{
		StepID: stepID,
		Config: input,
	}); err != nil {
		c.logger.Errorf("Step %s failed to write start work message: %v", stepID, err)
		return "", nil, fmt.Errorf("failed to write work start message (%w)", err)
	}
	c.logger.Debugf("Step %s started, waiting for response...", stepID)

	cborReader := c.decMode.NewDecoder(c.channel)

	var doneMessage workDoneMessage
	if c.atpVersion > 1 {
		// Loop and get all messages

		// Get the generic message, so we can find the type and decide the full message next.
		var runtimeMessage RuntimeMessage
		for {
			if err := cborReader.Decode(&runtimeMessage); err != nil {
				c.logger.Errorf("Step %s failed to read runtime message: %v", stepID, err)
				return "", nil, fmt.Errorf("failed to read runtime message (%w)", err)
			}
			switch runtimeMessage.MessageID {
			case MessageTypeWorkDone:
				doneMessage = runtimeMessage.MessageData.(workDoneMessage)
				break
			case MessageTypeSignal:
				signalMessage := runtimeMessage.MessageData.(signalMessage)
				c.logger.Infof("Step %s sent signal %s. Signal handling is not implemented.",
					signalMessage.SignalID)
			default:
				c.logger.Warningf("Step %s sent unknown message type: %s", stepID, runtimeMessage.MessageID)
			}
		}
	} else {
		if err := cborReader.Decode(&doneMessage); err != nil {
			c.logger.Errorf("Step %s failed to read work done message: %v", stepID, err)
			return "", nil, fmt.Errorf("failed to read work done message (%w)", err)
		}
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
