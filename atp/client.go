package atp

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	log "go.arcalot.io/log/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
	"io"
	"strings"
	"sync"
)

var supportedServerVersions = []int64{1, 3}

// ClientChannel holds the methods to talking to an ATP server (plugin).
type ClientChannel interface {
	io.Reader
	io.Writer
	io.Closer
}

type ExecutionResult struct {
	OutputID   string
	OutputData any
	Error      error
}

func NewErrorExecutionResult(err error) ExecutionResult {
	return ExecutionResult{"", nil, err}
}

// Client is the way to read information from the ATP server and then send a task to it in the form of a step.
// A step can only be sent once, but signals can be sent until the step is over. It is a single session.
type Client interface {
	// ReadSchema reads the schema from the ATP server.
	ReadSchema() (*schema.SchemaSchema, error)
	// Execute executes a step with a given context and returns the resulting output. Assumes you called ReadSchema first.
	Execute(input schema.Input, receivedSignals chan schema.Input, emittedSignals chan<- schema.Input) ExecutionResult
	Close() error
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
		make(chan bool, 5), // Buffer to prevent deadlocks
		make([]schema.Input, 0),
		make(map[string]chan schema.Input),
		make(map[string]chan ExecutionResult),
		make(map[string]chan<- schema.Input),
		sync.Mutex{},
		false,
		false,
		sync.WaitGroup{},
	}
}

func (c *client) Decoder() *cbor.Decoder {
	return c.decoder
}

func (c *client) Encoder() *cbor.Encoder {
	return c.encoder
}

type client struct {
	atpVersion                       int64
	channel                          ClientChannel
	decMode                          cbor.DecMode
	logger                           log.Logger
	decoder                          *cbor.Decoder
	encoder                          *cbor.Encoder
	doneChannel                      chan bool
	runningSteps                     []schema.Input
	runningSignalReceiveLoops        map[string]chan schema.Input    // Run ID to channel of signals to steps
	runningStepResultChannels        map[string]chan ExecutionResult // Run ID to channel of results
	runningStepEmittedSignalChannels map[string]chan<- schema.Input  // Run ID to channel of signals emitted from steps
	mutex                            sync.Mutex
	readLoopRunning                  bool
	done                             bool
	wg                               sync.WaitGroup // For the read loop.
}

func (c *client) sendCBOR(message any) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.encoder.Encode(message)
}

func (c *client) ReadSchema() (*schema.SchemaSchema, error) {
	c.logger.Debugf("Reading plugin schema...")

	if err := c.sendCBOR(nil); err != nil {
		c.logger.Errorf("Failed to encode ATP start output message: %v", err)
		return nil, fmt.Errorf("failed to encode start output message (%w)", err)
	}

	var hello HelloMessage
	if err := c.decoder.Decode(&hello); err != nil {
		c.logger.Errorf("Failed to decode ATP hello message: %v", err)
		return nil, fmt.Errorf("failed to decode hello message (%w)", err)
	}
	c.logger.Debugf("Hello message read, ATP version %d.", hello.Version)

	err := c.validateVersion(hello.Version)

	if err != nil {
		c.logger.Errorf("Unsupported plugin version. %w", err)
		return nil, fmt.Errorf("unsupported plugin version: %w", err)
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

func (c *client) validateVersion(serverVersion int64) error {
	for _, v := range supportedServerVersions {
		if serverVersion == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported atp version '%d', supported versions: %v", serverVersion, supportedServerVersions)
}

func (c *client) Execute(
	stepData schema.Input,
	receivedSignals chan schema.Input,
	emittedSignals chan<- schema.Input,
) ExecutionResult {
	c.logger.Debugf("Executing plugin step %s/%s...", stepData.RunID, stepData.ID)
	if len(stepData.RunID) == 0 {
		return NewErrorExecutionResult(fmt.Errorf("run ID is blank for step %s", stepData.ID))
	}
	var workStartMsg any
	workStartMsg = WorkStartMessage{
		StepID: stepData.ID,
		Config: stepData.InputData,
	}
	cborReader := c.decMode.NewDecoder(c.channel)
	if c.atpVersion > 1 {
		// Wrap it in a runtime message.
		workStartMsg = RuntimeMessage{RunID: stepData.RunID, MessageID: MessageTypeWorkStart, MessageData: workStartMsg}
		// Handle signals to the step
		if receivedSignals != nil {
			go func() {
				c.executeWriteLoop(stepData.RunID, receivedSignals)
			}()
		}
		// Setup channels for ATP v2
		err := c.prepareResultChannels(cborReader, stepData, emittedSignals)
		if err != nil {
			return NewErrorExecutionResult(err)
		}
	}
	if err := c.sendCBOR(workStartMsg); err != nil {
		c.logger.Errorf("Step '%s' failed to write start work message: %v", stepData.ID, err)
		return NewErrorExecutionResult(fmt.Errorf("failed to write work start message (%w)", err))
	}
	c.logger.Debugf("Step '%s' started, waiting for response...", stepData.ID)

	defer c.handleStepComplete(stepData.RunID, receivedSignals)
	return c.getResult(stepData, cborReader)
}

// handleStepComplete is the deferred function that will handle closing of the received channel.
func (c *client) handleStepComplete(runID string, receivedSignals chan schema.Input) {
	if receivedSignals != nil {
		c.logger.Infof("Closing signal channel for finished step")
		// Remove from the map to ensure that the client.Close() method doesn't double-close it
		c.mutex.Lock()
		// Validate that it exists, since Close() could have been called early.
		_, exists := c.runningSignalReceiveLoops[runID]
		if exists {
			delete(c.runningSignalReceiveLoops, runID)
			close(receivedSignals)
		}
		c.mutex.Unlock()
	}
}

// Close Tells the client that it's done, and can stop listening for more requests.
func (c *client) Close() error {
	c.mutex.Lock()
	if c.done {
		return nil
	}
	c.done = true
	// First, close channels that could send signals to the clients
	// This ends the loop
	for runID, signalChannel := range c.runningSignalReceiveLoops {
		c.logger.Infof("Closing signal channel for run ID '%s'", runID)
		delete(c.runningSignalReceiveLoops, runID)
		close(signalChannel)
	}
	c.mutex.Unlock()
	// Now tell the server we're done.
	// Send the client done message
	if c.atpVersion > 1 {
		err := c.sendCBOR(RuntimeMessage{
			MessageTypeClientDone,
			"",
			clientDoneMessage{},
		})
		if err != nil {
			return fmt.Errorf("client with steps '%s' failed to write client done message with error: %w",
				c.getRunningStepIDs(), err)
		}
	}
	c.wg.Wait()
	return nil
}

func (c *client) getRunningStepIDs() string {
	if len(c.runningSteps) == 0 {
		return "No running steps"
	}
	result := ""
	for _, step := range c.runningSteps {
		result += " " + step.RunID + "/" + step.ID
	}
	return result
}

// Listen for received signals, and send them over ATP if available.
func (c *client) executeWriteLoop(
	runID string,
	receivedSignals chan schema.Input,
) {
	// Add the channel to the client so that it can be kept track of
	c.mutex.Lock()
	c.runningSignalReceiveLoops[runID] = receivedSignals
	c.mutex.Unlock()
	defer func() {
		c.mutex.Lock()
		delete(c.runningSignalReceiveLoops, runID)
		c.mutex.Unlock()
	}()
	// Looped select that gets signals
	for {
		signal, ok := <-receivedSignals
		if !ok {
			c.logger.Infof("ATP signal loop done")
			return
		}
		c.logger.Debugf("Sending signal with ID '%s' to step with run ID '%s'", signal.ID, signal.RunID)
		if signal.ID == "" || signal.RunID == "" {
			c.logger.Errorf("Invalid run ID (%s) or signal ID (%s)", signal.ID, signal.RunID)
			return
		}
		if err := c.sendCBOR(RuntimeMessage{
			MessageTypeSignal,
			signal.RunID,
			SignalMessage{
				SignalID: signal.ID,
				Data:     signal.InputData,
			}}); err != nil {
			c.logger.Errorf("Client with steps '%s' failed to write signal (%s) with run id '&s' with error: %w",
				c.getRunningStepIDs(), signal.ID, signal.RunID, err)
			return
		}
		c.logger.Debugf("Successfully sent signal with ID '%s' to step with run ID '%s'", signal.ID, signal.RunID)
	}
}

// sendExecutionResult sends the results to the channel, and closes then removes the channels for the
// step results and the signals.
func (c *client) sendExecutionResult(runID string, result ExecutionResult) {
	c.logger.Debugf("Providing input for run ID '%s'", runID)
	c.mutex.Lock()
	resultChannel, found := c.runningStepResultChannels[runID]
	c.mutex.Unlock()
	if found {
		// Send the result
		resultChannel <- result
		// Close the channel and remove it to detect incorrectly duplicate results.
		close(resultChannel)
		c.mutex.Lock()
		delete(c.runningStepResultChannels, runID)
		c.mutex.Unlock()
	} else {
		c.logger.Errorf("Step result channel not found for run ID '%s'. This is either a bug in the ATP "+
			"client, or the plugin erroneously sent a second result.", runID)
	}
	// Now close the signal channel, since it's invalid to send a signal after the step is complete.
	c.mutex.Lock()
	defer c.mutex.Unlock()
	signalChannel, found := c.runningStepEmittedSignalChannels[runID]
	if !found {
		c.logger.Debugf("Could not find signal output channel for run ID '%s'", runID)
		return
	}
	close(signalChannel)
	delete(c.runningStepEmittedSignalChannels, runID)
}

func (c *client) sendErrorToAll(err error) {
	result := NewErrorExecutionResult(err)
	for runID := range c.runningStepResultChannels {
		c.sendExecutionResult(runID, result)
	}
}

//nolint:funlen
func (c *client) executeReadLoop(cborReader *cbor.Decoder) {
	defer func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.readLoopRunning = false
		c.wg.Done()
	}()
	// Loop and get all messages
	// The message is generic, so we must find the type and decode the full message next.
	var runtimeMessage DecodedRuntimeMessage
	for {
		if err := cborReader.Decode(&runtimeMessage); err != nil {
			c.logger.Errorf("ATP client for steps '%s' failed to read or decode runtime message: %v", c.getRunningStepIDs(), err)
			// This is fatal since the entire structure of the runtime message is invalid.
			c.sendErrorToAll(fmt.Errorf("failed to read or decode runtime message (%w)", err))
			return
		}
		switch runtimeMessage.MessageID {
		case MessageTypeWorkDone:
			var doneMessage WorkDoneMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &doneMessage); err != nil {
				c.logger.Errorf("Failed to decode work done message (%v) for run ID '%s' ", err, runtimeMessage.RunID)
				c.sendExecutionResult(runtimeMessage.RunID, NewErrorExecutionResult(
					fmt.Errorf("failed to decode work done message (%w)", err)))
			}
			c.sendExecutionResult(runtimeMessage.RunID, c.processWorkDone(runtimeMessage.RunID, doneMessage))
		case MessageTypeSignal:
			var signalMessage SignalMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &signalMessage); err != nil {
				c.logger.Errorf("ATP client for run ID '%s' failed to decode signal message: %v",
					runtimeMessage.RunID, err)
			}
			signalChannel, found := c.runningStepEmittedSignalChannels[runtimeMessage.RunID]
			if !found {
				c.logger.Warningf(
					"Step with run ID '%s' sent signal '%s'. Ignoring; signal handling is not implemented "+
						"(emittedSignals is nil).",
					runtimeMessage.RunID, signalMessage.SignalID)
			} else {
				c.logger.Debugf("Got signal from step with run ID '%s' with ID '%s'", runtimeMessage.RunID,
					signalMessage.SignalID)
				signalChannel <- signalMessage.ToInput(runtimeMessage.RunID)
			}
		case MessageTypeError:
			var errMessage ErrorMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &errMessage); err != nil {
				c.logger.Errorf("Step with run ID '%s' failed to decode error message: %v",
					runtimeMessage.RunID, err)
			}
			c.logger.Errorf("Step with run ID '%s' sent error message: %v", runtimeMessage.RunID, errMessage)
			resultMsg := fmt.Errorf("step '%s' sent error message: %s", runtimeMessage.RunID,
				errMessage.ToString(runtimeMessage.RunID))
			if errMessage.ServerFatal {
				c.sendErrorToAll(resultMsg)
				return // It's server fatal, so this is the last message from the server.
			} else if errMessage.StepFatal {
				if runtimeMessage.RunID == "" {
					c.sendErrorToAll(fmt.Errorf("step fatal error missing run id (%w)", resultMsg))
				} else {
					c.sendExecutionResult(runtimeMessage.RunID, NewErrorExecutionResult(resultMsg))
				}
			}
		default:
			c.logger.Warningf("Step with run ID '%s' sent unknown message type: %s", runtimeMessage.RunID,
				runtimeMessage.MessageID)
		}
		c.mutex.Lock()
		if len(c.runningStepResultChannels) == 0 {
			c.mutex.Unlock()
			return // Done
		}
		c.mutex.Unlock()
	}
}

// executeStep handles the reading of work done, signals, or any other outputs from the plugins.
// It branches off with different logic for ATP versions 1 and 2.
func (c *client) getResult(
	stepData schema.Input,
	cborReader *cbor.Decoder,
) ExecutionResult {
	if c.atpVersion >= 2 {
		return c.getResultV2(stepData)
	} else {
		return c.getResultV1(cborReader, stepData)
	}
}

// getResultV1 is the legacy function that only waits for work done.
func (c *client) getResultV1(
	cborReader *cbor.Decoder,
	stepData schema.Input,
) ExecutionResult {
	var doneMessage WorkDoneMessage
	if err := cborReader.Decode(&doneMessage); err != nil {
		c.logger.Errorf("Failed to read or decode work done message: (%w) for step %s", err, stepData.ID)
		return NewErrorExecutionResult(
			fmt.Errorf("failed to read or decode work done message (%w) for step %s", err, stepData.ID))
	}
	return c.processWorkDone(stepData.RunID, doneMessage)
}

func (c *client) prepareResultChannels(
	cborReader *cbor.Decoder,
	stepData schema.Input,
	emittedSignals chan<- schema.Input,
) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, existing := c.runningStepResultChannels[stepData.RunID]
	if existing {
		return fmt.Errorf("duplicate run ID given '%s'", stepData.RunID)
	}
	// Set up the signal and step results channels
	resultChannel := make(chan ExecutionResult)
	c.runningStepResultChannels[stepData.RunID] = resultChannel
	if emittedSignals != nil {
		c.runningStepEmittedSignalChannels[stepData.RunID] = emittedSignals
	}
	// Run the loop if it isn't running.
	if !c.readLoopRunning {
		// Only a single read loop should be running
		c.wg.Add(1) // Add here, so that it's before the goroutine to prevent race conditions.
		c.readLoopRunning = true
		go func() {
			c.executeReadLoop(cborReader)
		}()
	}
	return nil
}

// getResultV2 works with the channels that communicate with the RuntimeMessage loop.
func (c *client) getResultV2(
	stepData schema.Input,
) ExecutionResult {
	c.mutex.Lock()
	resultChannel, found := c.runningStepResultChannels[stepData.RunID]
	c.mutex.Unlock()
	if !found {
		return NewErrorExecutionResult(
			fmt.Errorf("could not find result channel for step with run ID '%s'",
				stepData.RunID),
		)
	}
	// Wait for the result
	result, received := <-resultChannel
	if !received {
		return NewErrorExecutionResult(
			fmt.Errorf("did not receive result from results channel in ATP client for step with run ID '%s'",
				stepData.RunID),
		)
	}
	return result
}

func (c *client) processWorkDone(
	runID string,
	doneMessage WorkDoneMessage,
) ExecutionResult {
	c.logger.Debugf("Step with run ID '%s' completed with output ID '%s'.", runID, doneMessage.OutputID)

	// Print debug logs from the step as debug.
	debugLogs := strings.Split(doneMessage.DebugLogs, "\n")
	for _, line := range debugLogs {
		if strings.TrimSpace(line) != "" {
			c.logger.Debugf("Step '%s' debug: %s", runID, line)
		}
	}

	return ExecutionResult{doneMessage.OutputID, doneMessage.OutputData, nil}
}
