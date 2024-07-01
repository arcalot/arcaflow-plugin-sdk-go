package atp

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	log "go.arcalot.io/log/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
	"io"
	"strings"
	"sync"
	"time"
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
		make(map[string]*executionEntry),
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

type executionEntry struct {
	result    *ExecutionResult
	condition sync.Cond
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
	runningSignalReceiveLoops        map[string]chan schema.Input   // Run ID to channel of signals to steps
	runningStepResultEntries         map[string]*executionEntry     // Run ID to results
	runningStepEmittedSignalChannels map[string]chan<- schema.Input // Run ID to channel of signals emitted from steps
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
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
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
		c.mutex.Unlock()
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
			// add a timeout to the wait to prevent it from causing a deadlock.
			// 5 seconds is arbitrary, but gives it enough time to exit.
			waitedGracefully := waitWithTimeout(time.Second*5, &c.wg)
			if waitedGracefully {
				return fmt.Errorf("client with step '%s' failed to write client done message with error: %w",
					c.getRunningStepIDs(), err)
			} else {
				panic(fmt.Errorf("potential deadlock after client with step '%s' failed to write client done message with error: %w",
					c.getRunningStepIDs(), err))
			}
		}
	}
	c.wg.Wait()
	return nil
}

// waitWithTimeout waits for the provided wait group, aborting the wait if
// the provided timeout expires.
// Returns true if the WaitGroup finished, and false if
// it reached the end of the timeout.
func waitWithTimeout(duration time.Duration, wg *sync.WaitGroup) bool {
	// Run a goroutine to do the waiting
	doneChannel := make(chan bool, 1)
	go func() {
		defer close(doneChannel)
		wg.Wait()
	}()
	select {
	case <-doneChannel:
		return true
	case <-time.After(duration):
		return false
	}
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
	c.mutex.Lock()
	if c.done {
		c.mutex.Unlock()
		// It is important to abort now since Close() was called. This is to prevent the channel
		// from being added to the channel, since Close() uses that map to determine the exit
		// condition. Adding to the map would cause it to never exit.
		c.logger.Warningf("write called loop for run ID %q on done client; aborting", runID)
		return
	}
	// Add the channel to the client so that it can be kept track of
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

// sendExecutionResult finalizes the result entry for processing by the client's caller, and
// closes then removes the channels for the signals.
// The caller must have the mutex locked while calling this function.
func (c *client) sendExecutionResult(runID string, result ExecutionResult) {
	c.logger.Debugf("Sending results for run ID '%s'", runID)
	resultEntry, found := c.runningStepResultEntries[runID]
	if found {
		// Send the result
		resultEntry.result = &result
		resultEntry.condition.Signal()
	} else {
		c.logger.Errorf("Step result entry not found for run ID '%s'. This is either a bug in the ATP "+
			"client, or the plugin erroneously sent a second result.", runID)
	}
	// Now close the signal channel, since it's invalid to send a signal after the step is complete.
	signalChannel, found := c.runningStepEmittedSignalChannels[runID]
	if !found {
		return
	}
	close(signalChannel)
	delete(c.runningStepEmittedSignalChannels, runID)
}

func (c *client) sendErrorToAll(err error) {
	result := NewErrorExecutionResult(err)
	c.mutex.Lock()
	for runID := range c.runningStepResultEntries {
		c.sendExecutionResult(runID, result)
	}
	c.mutex.Unlock()
}

func (c *client) handleWorkDoneMessage(runtimeMessage DecodedRuntimeMessage) {
	var doneMessage WorkDoneMessage
	if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &doneMessage); err != nil {
		c.logger.Errorf("Failed to decode work done message (%v) for run ID '%s' ", err, runtimeMessage.RunID)
		c.mutex.Lock()
		c.sendExecutionResult(runtimeMessage.RunID, NewErrorExecutionResult(
			fmt.Errorf("failed to decode work done message (%w)", err)))
		c.mutex.Unlock()
		return
	}
	c.mutex.Lock()
	c.sendExecutionResult(runtimeMessage.RunID, c.processWorkDone(runtimeMessage.RunID, doneMessage))
	c.mutex.Unlock()
}

func (c *client) handleSignalMessage(runtimeMessage DecodedRuntimeMessage) {
	var signalMessage SignalMessage
	if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &signalMessage); err != nil {
		c.logger.Errorf("ATP client for run ID '%s' failed to decode signal message: %v",
			runtimeMessage.RunID, err)
	}
	c.mutex.Lock()
	signalChannel, found := c.runningStepEmittedSignalChannels[runtimeMessage.RunID]
	c.mutex.Unlock()
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
}

// Returns true if fatal, requiring aborting the read loop.
func (c *client) handleErrorMessage(runtimeMessage DecodedRuntimeMessage) bool {
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
		return true // It's server fatal, so this is the last message from the server.
	} else if errMessage.StepFatal {
		if runtimeMessage.RunID == "" {
			c.sendErrorToAll(fmt.Errorf("step fatal error missing run id (%w)", resultMsg))
		} else {
			c.sendExecutionResult(runtimeMessage.RunID, NewErrorExecutionResult(resultMsg))
		}
	}
	return false
}

func (c *client) hasEntriesRemaining() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	remainingSteps := 0
	for _, resultEntry := range c.runningStepResultEntries {
		// The result is the reliable way to determine if it's done. There is a fraction of
		// time when the entry is still in the map, but it is done.
		if resultEntry.result == nil {
			remainingSteps++
		}
	}
	return remainingSteps != 0
}

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
			c.handleWorkDoneMessage(runtimeMessage)
		case MessageTypeSignal:
			c.handleSignalMessage(runtimeMessage)
		case MessageTypeError:
			fatal := c.handleErrorMessage(runtimeMessage)
			if fatal {
				return
			}
		default:
			c.logger.Warningf("Step with run ID '%s' sent unknown message type: %s", runtimeMessage.RunID,
				runtimeMessage.MessageID)
		}
		// The non-error exit condition is having no more entries remaining.
		if !c.hasEntriesRemaining() {
			return
		}
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
	c.logger.Debugf("Preparing result channels for step with run ID %q", stepData.RunID)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, existing := c.runningStepResultEntries[stepData.RunID]
	if existing {
		return fmt.Errorf("duplicate run ID given '%s'", stepData.RunID)
	}
	// Set up the signal and step results channels
	resultEntry := executionEntry{
		result:    nil,
		condition: sync.Cond{L: &c.mutex},
	}
	c.runningStepResultEntries[stepData.RunID] = &resultEntry
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

// getResultV2 communicates with the RuntimeMessage loop to get the ExecutionResult.
func (c *client) getResultV2(
	stepData schema.Input,
) ExecutionResult {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	resultEntry, found := c.runningStepResultEntries[stepData.RunID]
	if !found {
		return NewErrorExecutionResult(
			fmt.Errorf("could not find result entry for step with run ID '%s'. Existing entries: %v",
				stepData.RunID, c.runningStepResultEntries),
		)
	}
	if resultEntry.result == nil {
		// Wait for the result
		resultEntry.condition.Wait()
	}
	if resultEntry.result == nil {
		return NewErrorExecutionResult(
			fmt.Errorf("did not receive result from results entry in ATP client for step with run ID '%s'",
				stepData.RunID),
		)
	}
	// Deletion of the entry needs to be done in this function after waiting for
	// the value to ensure the value's lifetime is long enough in the map.
	// It cannot be removed on the sender's side, since that would cause a race.
	delete(c.runningStepResultEntries, stepData.RunID)
	return *resultEntry.result
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
