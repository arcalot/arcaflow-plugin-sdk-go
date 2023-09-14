package atp

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	log "go.arcalot.io/log/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
	"io"
	"strings"
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
// A step can only be sent once, but signals can be sent until the step is over. It is a single session.
type Client interface {
	// ReadSchema reads the schema from the ATP server.
	ReadSchema() (*schema.SchemaSchema, error)
	// Execute executes a step with a given context and returns the resulting output.
	Execute(input schema.Input, receivedSignals chan schema.Input, emittedSignals chan<- schema.Input) (outputID string, outputData any, err error)
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
	atpVersion int64
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

func (c *client) Execute(
	stepData schema.Input,
	receivedSignals chan schema.Input,
	emittedSignals chan<- schema.Input,
) (outputID string, outputData any, err error) {
	c.logger.Debugf("Executing plugin step %s...", stepData.ID)
	if err := c.encoder.Encode(StartWorkMessage{
		StepID: stepData.ID,
		Config: stepData.InputData,
	}); err != nil {
		c.logger.Errorf("Step %s failed to write start work message: %v", stepData.ID, err)
		return "", nil, fmt.Errorf("failed to write work start message (%w)", err)
	}
	c.logger.Debugf("Step %s started, waiting for response...", stepData.ID)

	doneChannel := make(chan bool, 1) // Needs a buffer to not hang.
	defer handleClientClosure(receivedSignals, doneChannel)
	if c.atpVersion > 1 {
		go func() {
			c.executeWriteLoop(stepData, receivedSignals, doneChannel)
		}()
	}
	return c.executeReadLoop(stepData, receivedSignals)
}

// handleClosure is the deferred function that will handle closing of the received channel,
// and notifying the code that it's closed.
// Note: The doneChannel should have a buffer.
func handleClientClosure(receivedSignals chan schema.Input, doneChannel chan bool) {
	doneChannel <- true
	if receivedSignals != nil {
		close(receivedSignals)
	}
}

// Listen for received signals, and send them over ATP if available.
func (c *client) executeWriteLoop(
	stepData schema.Input,
	receivedSignals chan schema.Input,
	doneChannel chan bool,
) {
	// Looped select that gets signals
	for {
		var signal schema.Input
		select {
		case <-doneChannel:
			// Send the client done message
			err := c.encoder.Encode(RuntimeMessage{
				MessageTypeClientDone,
				clientDoneMessage{},
			})
			if err != nil {
				c.logger.Errorf("Step %s failed to write client done message with error: %w", stepData.ID, err)
			}
			return
		case receivedSignal, ok := <-receivedSignals:
			if !ok {
				// It's not supposed to be not ok yet.
				c.logger.Errorf("error in channel preparing to send signal (step %s, signal %s) over ATP",
					stepData.ID, signal.ID)
				return
			}
			signal = receivedSignal
		}
		c.logger.Debugf("Sending signal with ID '%s' to step with ID '%s'", signal.ID, stepData.ID)
		if err := c.encoder.Encode(RuntimeMessage{
			MessageTypeSignal,
			signalMessage{
				StepID:   stepData.ID,
				SignalID: signal.ID,
				Data:     signal.InputData,
			}}); err != nil {
			c.logger.Errorf("Step %s failed to write signal (%s) with error: %w", stepData.ID, signal.ID, err)
			return
		}
		c.logger.Debugf("Successfully sent signal with ID '%s' to step with ID '%s'", signal.ID, stepData.ID)
	}
}

// executeReadLoop handles the reading of work done, signals, or any other outputs from the plugins.
// It branches off with different logic for ATP versions 1 and 2.
func (c *client) executeReadLoop(
	stepData schema.Input, emittedSignals chan<- schema.Input,
) (outputID string, outputData any, err error) {
	cborReader := c.decMode.NewDecoder(c.channel)
	if c.atpVersion >= 2 {
		return c.executeReadLoopV2(cborReader, stepData, emittedSignals)
	} else {
		return c.executeReadLoopV1(cborReader, stepData)
	}
}

// executeReadLoopV1 is the legacy read loop function, that only waits for work done.
func (c *client) executeReadLoopV1(
	cborReader *cbor.Decoder,
	stepData schema.Input,
) (outputID string, outputData any, err error) {
	var doneMessage workDoneMessage
	if err := cborReader.Decode(&doneMessage); err != nil {
		c.logger.Errorf("Failed to read or decode work done message: (%w) for step %s", err, stepData.ID)
		return "", nil,
			fmt.Errorf("failed to read or decode work done message (%w) for step %s", err, stepData.ID)
	}
	return c.handleWorkDone(stepData, doneMessage)
}

// executeReadLoopV2 is the new read loop function, that supports the RuntimeMessage loop.
func (c *client) executeReadLoopV2(
	cborReader *cbor.Decoder,
	stepData schema.Input,
	emittedSignals chan<- schema.Input,
) (outputID string, outputData any, err error) {
	// Loop and get all messages
	// The message is generic, so we must find the type and decode the full message next.
	var runtimeMessage DecodedRuntimeMessage
	for {
		if err := cborReader.Decode(&runtimeMessage); err != nil {
			c.logger.Errorf("Step %s failed to read or decode runtime message: %v", stepData.ID, err)
			return "", nil,
				fmt.Errorf("failed to read or decode runtime message (%w)", err)
		}
		switch runtimeMessage.MessageID {
		case MessageTypeWorkDone:
			var doneMessage workDoneMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &doneMessage); err != nil {
				c.logger.Errorf("Failed to decode work done message (%v) for step ID %s ", err, stepData.ID)
				return "", nil,
					fmt.Errorf("failed to read work done message (%w)", err)
			}
			return c.handleWorkDone(stepData, doneMessage)
		case MessageTypeSignal:
			var signalMessage signalMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &signalMessage); err != nil {
				c.logger.Errorf("Step %s failed to decode signal message: %v", stepData.ID, err)
			}
			if stepData.ID != signalMessage.StepID {
				c.logger.Warningf("Step %s sent signal %s, but the step ID '%s' sent by the plugin does not match. Ignoring signal.",
					stepData.ID, signalMessage.SignalID, signalMessage.StepID)
				continue // Don't process the signal
			}
			if emittedSignals == nil {
				c.logger.Warningf("Step '%s' sent signal '%s'. Ignoring; signal handling is not implemented (emittedSignals is nil).",
					stepData.ID, signalMessage.SignalID)
			} else {
				c.logger.Debugf("Got signal from step '%s' with ID '%s'", stepData.ID, signalMessage.SignalID)
				emittedSignals <- signalMessage.ToInput()
			}
		default:
			c.logger.Warningf("Step %s sent unknown message type: %s", stepData.ID, runtimeMessage.MessageID)
		}
	}
}

func (c *client) handleWorkDone(
	stepData schema.Input,
	doneMessage workDoneMessage,
) (outputID string, outputData any, err error) {
	c.logger.Debugf("Step %s completed with output ID '%s'.", stepData.ID, doneMessage.OutputID)

	// Print debug logs from the step as debug.
	debugLogs := strings.Split(doneMessage.DebugLogs, "\n")
	for _, line := range debugLogs {
		if strings.TrimSpace(line) != "" {
			c.logger.Debugf("Step %s debug: %s", stepData.ID, line)
		}
	}

	return doneMessage.OutputID, doneMessage.OutputData, nil
}
