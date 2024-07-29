package atp

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"go.arcalot.io/log/v2"
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
// Currently used only by tests in the Python- and Test-deployers.
//
//goland:noinspection GoUnusedExportedFunction
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
		make(map[string]chan<- schema.Input),
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
	runningSignalReceiveLoops        map[string]chan<- schema.Input // Run ID to channel of signals to steps
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
		err = fmt.Errorf("unsupported plugin version: %w", err)
		c.logger.Errorf(err.Error())
		return nil, err
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
			defer c.handleStepComplete(stepData.RunID)
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

	return c.getResult(stepData, cborReader)
}

// handleStepComplete performs cleanup actions for a step when its execution is
// complete.  Currently, this consists solely of removing its entry from the
// map of channels used to send it signals.
func (c *client) handleStepComplete(runID string) {
	c.logger.Infof("Closing signal channel for finished step")
	c.mutex.Lock()
	delete(c.runningSignalReceiveLoops, runID)
	c.mutex.Unlock()
}

// Close Tells the client that it's done, and can stop listening for more requests.
func (c *client) Close() error {
	c.mutex.Lock()
	if c.done {
		c.mutex.Unlock()
		return nil
	}
	c.done = true
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
		// Close() was called, so exit now.
		// Failure to exit now may result in this receivedSignals channel not getting
		// closed, resulting in this function hanging.
		c.logger.Warningf(
			"write called loop for run ID %q on done client; skipping receive loop",
			runID,
		)
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
			c.logger.Errorf(
				"Client with steps '%s' failed to write signal (%s) with run id %q with error: %v",
				c.getRunningStepIDs(),
				signal.ID,
				signal.RunID,
				err,
			)
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
	delete(c.runningStepEmittedSignalChannels, runID)
	close(signalChannel)
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
	var result ExecutionResult
	if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &doneMessage); err != nil {
		c.logger.Errorf("Failed to decode work done message (%v) for run ID '%s' ", err, runtimeMessage.RunID)
		result = NewErrorExecutionResult(fmt.Errorf("failed to decode work done message (%w)", err))
	} else {
		result = c.processWorkDone(runtimeMessage.RunID, doneMessage)
	}
	c.mutex.Lock()
	c.sendExecutionResult(runtimeMessage.RunID, result)
	c.mutex.Unlock()
}

func (c *client) handleSignalMessage(runtimeMessage DecodedRuntimeMessage) {
	var signalMessage SignalMessage
	if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &signalMessage); err != nil {
		c.logger.Errorf("ATP client for run ID '%s' failed to decode signal message: %v",
			runtimeMessage.RunID, err)
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock() // Hold lock until we send to the channel to prevent premature closing of the channel.
	signalChannel, found := c.runningStepEmittedSignalChannels[runtimeMessage.RunID]
	if !found {
		c.logger.Warningf(
			"Step with run ID '%s' sent signal '%s'. Ignoring; signal handling is not implemented "+
				"(emittedSignals is nil).",
			runtimeMessage.RunID, signalMessage.SignalID)
		return
	}
	c.logger.Debugf("Got signal from step with run ID '%s' with ID '%s'", runtimeMessage.RunID,
		signalMessage.SignalID)
	signalChannel <- signalMessage.ToInput(runtimeMessage.RunID)
}

// Returns true if the error is fatal.
func (c *client) handleErrorMessage(runtimeMessage DecodedRuntimeMessage) bool {
	var errMessage ErrorMessage
	if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &errMessage); err != nil {
		c.logger.Errorf("Step with run ID '%s' failed to decode error message: %v",
			runtimeMessage.RunID, err)
	}
	errorMessageStr := errMessage.ToString(runtimeMessage.RunID)
	resultMsg := fmt.Errorf("step with run ID %q sent error message: %s", runtimeMessage.RunID, errorMessageStr)
	c.logger.Errorf(resultMsg.Error())
	if errMessage.ServerFatal {
		c.sendErrorToAll(resultMsg)
		return true // It's server fatal, so this is the last message from the server.
	} else if errMessage.StepFatal {
		if runtimeMessage.RunID == "" {
			c.sendErrorToAll(fmt.Errorf("step fatal error missing run id (%w)", resultMsg))
		} else {
			c.mutex.Lock()
			c.sendExecutionResult(runtimeMessage.RunID, NewErrorExecutionResult(resultMsg))
			c.mutex.Unlock()
		}
	}
	return false
}

func (c *client) hasEntriesRemaining() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, resultEntry := range c.runningStepResultEntries {
		// If any result is nil then we're not done.
		// Context: There is a fraction of time when the entry is still in the map
		// following completion. It is set to a non-nil value when done.
		if resultEntry.result == nil {
			return true
		}
	}
	return false
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
			c.logger.Errorf(
				"ATP client for steps '%s' failed to read or decode runtime message: %v",
				c.getRunningStepIDs(),
				err,
			)
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
			if c.handleErrorMessage(runtimeMessage) {
				return // Fatal
			}
		default:
			c.logger.Warningf(
				"Step with run ID '%s' sent unknown message type: %d",
				runtimeMessage.RunID,
				runtimeMessage.MessageID,
			)
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
		err = fmt.Errorf("failed to read or decode work done message (%w) for step %s", err, stepData.ID)
		c.logger.Errorf(err.Error())
		return NewErrorExecutionResult(err)
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
func (c *client) getResultV2(stepData schema.Input) ExecutionResult {
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
		panic(fmt.Errorf("did not receive result from results entry in ATP client for step with run ID '%s'",
			stepData.RunID))
	}
	// Now that we've received the result for this step, remove it from the list of running steps.
	// We do this here because the sender cannot tell when the message has been received, and so
	// it cannot tell when it is safe to remove the entry from the map.
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
