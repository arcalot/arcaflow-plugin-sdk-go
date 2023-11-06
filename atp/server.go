package atp

import (
	"context"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
	"io"
	"os"
	"sync"
)

// RunATPServer runs an ArcaflowTransportProtocol server with a given schema.
func RunATPServer(
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	pluginSchema *schema.CallableSchema,
) []*ServerError {
	session := initializeATPServerSession(ctx, stdin, stdout, pluginSchema)
	session.wg.Add(1)

	// Run needs to be run in its own goroutine to allow for the closure handling to happen simultaneously.
	go func() {
		session.run()
	}()

	workError := session.handleClosure()

	// Ensure that the session is done.
	session.wg.Wait()
	return workError
}

type atpServerSession struct {
	ctx            context.Context
	wg             *sync.WaitGroup
	stdinCloser    io.ReadCloser
	cborStdin      *cbor.Decoder
	cborStdout     *cbor.Encoder
	runningSteps   map[string]string // Maps run ID to step ID
	workDone       chan ServerError
	runDoneChannel chan bool
	pluginSchema   *schema.CallableSchema
	encoderMutex   sync.Mutex
}

type ServerError struct {
	RunID       string
	Err         error
	StepFatal   bool
	ServerFatal bool
}

func (e ServerError) String() string {
	return fmt.Sprintf("RunID: '%s', err: %s, step fatal: %t, server fatal: %t", e.RunID, e.Err, e.StepFatal, e.ServerFatal)
}

func initializeATPServerSession(
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	pluginSchema *schema.CallableSchema,
) *atpServerSession {
	workDone := make(chan ServerError, 3)
	// The ATP protocol uses CBOR.
	cborStdin := cbor.NewDecoder(stdin)
	cborStdout := cbor.NewEncoder(stdout)
	runDoneChannel := make(chan bool, 3) // Buffer to prevent it from hanging if something unexpected happens.

	return &atpServerSession{
		ctx:            ctx,
		cborStdin:      cborStdin,
		stdinCloser:    stdin,
		cborStdout:     cborStdout,
		workDone:       workDone,
		runDoneChannel: runDoneChannel,
		pluginSchema:   pluginSchema,
		wg:             &sync.WaitGroup{},
		runningSteps:   make(map[string]string),
	}
}

func (s *atpServerSession) sendRuntimeMessage(msgID uint32, runID string, message any) error {
	s.encoderMutex.Lock()
	defer s.encoderMutex.Unlock()
	return s.cborStdout.Encode(RuntimeMessage{
		MessageID:   msgID,
		RunID:       runID,
		MessageData: message,
	})
}

func (s *atpServerSession) handleClosure() []*ServerError {
	// Wait for work done or context complete.
	var errors []*ServerError
closeLoop:
	for {
		select {
		case errorSent, wasError := <-s.workDone:
			if !wasError {
				break closeLoop
			}
			errors = append(errors, &errorSent)
			err := s.sendRuntimeMessage(
				MessageTypeError,
				errorSent.RunID,
				ErrorMessage{
					Error:       errorSent.Err.Error(),
					StepFatal:   errorSent.StepFatal,
					ServerFatal: errorSent.ServerFatal,
				},
			)
			// If that didn't send, just send to stderr now.
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error while sending error message: %s\n", err)
			}
			// If either the error report sending failed, or the error was server fatal, stop here.
			if err != nil || errorSent.ServerFatal {
				err = s.stdinCloser.Close()
				if err != nil {
					return append(errors, &ServerError{
						RunID:       errorSent.RunID,
						Err:         fmt.Errorf("error closing stdin (%w) after workDone error (%v)", err, errorSent),
						StepFatal:   true,
						ServerFatal: true,
					})
				} else {
					break closeLoop
				}
			}
		case <-s.ctx.Done():
			// Likely got sigterm. Just close. Ideally gracefully.
			break closeLoop
		}
	}
	// Now close the pipe that it gets input from.
	return errors
}

func (s *atpServerSession) runATPReadLoop() {
	// The message is generic, so we must find the type and decode the full message next.
	var runtimeMessage DecodedRuntimeMessage
	for {
		// First, decode the message
		// Note: This blocks. To abort early, close stdin.
		if err := s.cborStdin.Decode(&runtimeMessage); err != nil {
			// Failed to decode. If it's done, that's okay. If not, there's a problem.
			done := false
			select {
			case done = <-s.runDoneChannel:
			default:
				// Prevents it from blocking
			}
			if !done {
				s.workDone <- ServerError{
					RunID:       "",
					Err:         fmt.Errorf("failed to read or decode runtime message: %w", err),
					StepFatal:   true,
					ServerFatal: true,
				}
			} // If done, it didn't get the work done message, which is not ideal.
			return
		}
		done := s.onRuntimeMessageReceived(&runtimeMessage)
		if done {
			return
		}
	}
}

// onRuntimeMessageReceived handles the runtime message by determining what type it is, and executing the proper path.
// Returns true if termination should be terminated, which should correspond to only client done or fatal server errors.
func (s *atpServerSession) onRuntimeMessageReceived(message *DecodedRuntimeMessage) bool {
	runID := message.RunID
	switch message.MessageID {
	case MessageTypeWorkStart:
		var workStartMsg WorkStartMessage
		if err := cbor.Unmarshal(message.RawMessageData, &workStartMsg); err != nil {
			s.workDone <- ServerError{
				RunID:       runID,
				Err:         fmt.Errorf("failed to decode work start message: %w", err),
				StepFatal:   true,
				ServerFatal: false,
			}
			return false
		}
		s.handleWorkStartMessage(runID, workStartMsg)
		return false
	case MessageTypeSignal:
		var signalMessage SignalMessage
		if err := cbor.Unmarshal(message.RawMessageData, &signalMessage); err != nil {
			s.workDone <- ServerError{
				RunID:       runID,
				Err:         fmt.Errorf("failed to decode signal message: %w", err),
				StepFatal:   false,
				ServerFatal: false,
			}
			return false
		}
		s.handleSignalMessage(runID, signalMessage)

		return false
	case MessageTypeClientDone:
		// It's now safe to close the channel
		err := s.stdinCloser.Close()
		if err != nil {
			s.workDone <- ServerError{
				// this error does not apply to a specific run id
				RunID:       "",
				Err:         fmt.Errorf("error while closing stdin on client done: %w", err),
				StepFatal:   true,
				ServerFatal: true,
			}
		}
		return true // Client done, so terminate loop
	default:
		s.workDone <- ServerError{
			// this error does not apply to a specific run id
			RunID: "",
			Err: fmt.Errorf("unknown message ID received: %d. This is a sign of incompatible server and client versions",
				message.MessageID),
			StepFatal:   false,
			ServerFatal: false,
		}
		return false
	}
}

func (s *atpServerSession) handleWorkStartMessage(runID string, workStartMsg WorkStartMessage) {
	if runID == "" || workStartMsg.StepID == "" {
		s.workDone <- ServerError{
			RunID: "",
			Err: fmt.Errorf("missing runID (%s) or stepID in work start message (%s)",
				runID, workStartMsg.StepID),
			StepFatal:   true,
			ServerFatal: false,
		}
		return
	}
	s.runningSteps[runID] = workStartMsg.StepID
	s.wg.Add(1) // Wait until the step is done
	go func() {
		s.runStep(runID, workStartMsg)
		s.wg.Done()
	}()
}

func (s *atpServerSession) handleSignalMessage(runID string, signalMessage SignalMessage) {
	if runID == "" {
		s.workDone <- ServerError{
			RunID:       "",
			Err:         fmt.Errorf("RunID missing for signal '%s' in signal message", signalMessage.SignalID),
			StepFatal:   false,
			ServerFatal: false,
		}
		return
	}
	stepID, found := s.runningSteps[runID]
	if !found {
		s.workDone <- ServerError{
			RunID:       runID,
			Err:         fmt.Errorf("unknown step with run ID '%s' in signal mesage", runID),
			StepFatal:   false,
			ServerFatal: false,
		}
		return
	}
	s.wg.Add(1) // Wait until the signal handler is done
	go func() {
		if err := s.pluginSchema.CallSignal(
			s.ctx,
			runID,
			stepID,
			signalMessage.SignalID,
			signalMessage.Data,
		); err != nil {
			s.workDone <- ServerError{
				RunID: runID,
				Err: fmt.Errorf("failed while running signal ID %s: %w",
					signalMessage.SignalID, err),
				StepFatal:   false,
				ServerFatal: false,
			}
		}
		s.wg.Done()
	}()
}

func (s *atpServerSession) run() {
	defer func() {
		s.runDoneChannel <- true
		close(s.workDone)
		s.wg.Done()
	}()

	err := s.sendInitialMessagesToClient()
	if err != nil {
		s.workDone <- ServerError{
			RunID:       "",
			Err:         fmt.Errorf("error while sending initial messages to client (%w)", err),
			StepFatal:   true,
			ServerFatal: true,
		}
		return
	}

	// Now, loop through stdin inputs until the step ends.
	s.runATPReadLoop()
}

func (s *atpServerSession) runStep(runID string, req WorkStartMessage) {
	// Call the step in the provided callable schema.
	defer func() {
		// Handle and properly report panics
		if r := recover(); r != nil {
			s.workDone <- ServerError{
				RunID:       runID,
				Err:         fmt.Errorf("panic while running step with Run ID '%s': (%v)", runID, r),
				StepFatal:   true,
				ServerFatal: false,
			}
		}
	}()
	outputID, outputData, err := s.pluginSchema.CallStep(s.ctx, runID, req.StepID, req.Config)
	if err != nil {
		s.workDone <- ServerError{
			RunID:       runID,
			Err:         fmt.Errorf("error calling step (%w)", err),
			StepFatal:   true,
			ServerFatal: false,
		}
		return
	}
	// Lastly, send the work done message.
	err = s.sendRuntimeMessage(
		MessageTypeWorkDone,
		runID,
		WorkDoneMessage{
			req.StepID,
			outputID,
			outputData,
			"",
		},
	)
	if err != nil {
		// At this point, the work done message failed to send, so it's likely that sending an ErrorMessage would fail.
		_, _ = fmt.Fprintf(os.Stderr, "error while sending work done message: %s\n", err)
	}
}

func (s *atpServerSession) sendInitialMessagesToClient() error {
	// Start by serializing the schema, since the protocol requires sending the schema on the hello message.
	serializedSchema, err := s.pluginSchema.SelfSerialize()
	if err != nil {
		return err
	}

	// First, the start message, which is just an empty message.
	var empty any
	err = s.cborStdin.Decode(&empty)
	if err != nil {
		return fmt.Errorf("failed to CBOR-decode start output message (%w)", err)
	}

	// Next, send the hello message, which includes the version and schema.
	err = s.cborStdout.Encode(HelloMessage{ProtocolVersion, serializedSchema})
	if err != nil {
		return fmt.Errorf("failed to CBOR-encode schema (%w)", err)
	}
	return nil
}
