package atp

import (
	"context"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// RunATPServer runs an ArcaflowTransportProtocol server with a given schema.
func RunATPServer(
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	pluginSchema *schema.CallableSchema,
) *ServerError {
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
	cancel         *context.CancelFunc
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
	return fmt.Sprintf("RunID: %s, err: %s, step fatal: %t, server fatal: %t", e.RunID, e.Err, e.StepFatal, e.ServerFatal)
}

func initializeATPServerSession(
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	pluginSchema *schema.CallableSchema,
) *atpServerSession {
	subCtx, cancel := context.WithCancel(ctx)
	workDone := make(chan ServerError, 3)
	// The ATP protocol uses CBOR.
	cborStdin := cbor.NewDecoder(stdin)
	cborStdout := cbor.NewEncoder(stdout)
	runDoneChannel := make(chan bool, 3) // Buffer to prevent it from hanging if something unexpected happens.

	// Cancel the sub context on sigint or sigterm.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			// Got sigterm. So cancel context.
			cancel()
		case <-subCtx.Done():
			// Done. No sigterm.
		}
	}()

	return &atpServerSession{
		ctx:            subCtx,
		cancel:         &cancel,
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

func (s *atpServerSession) handleClosure() *ServerError {
	// Wait for work done or context complete.
	var workError *ServerError
closeLoop:
	for {
		select {
		case errorSent, wasError := <-s.workDone:
			if wasError {
				workError = &errorSent
				err := s.sendRuntimeMessage(
					MessageTypeError,
					errorSent.RunID,
					errorMessage{
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
						return &ServerError{
							RunID:       workError.RunID,
							Err:         fmt.Errorf("error closing stdin (%s) after workDone error (%v)", err, workError),
							StepFatal:   true,
							ServerFatal: true,
						}
					} else {
						break closeLoop
					}
				}
			} else {
				break closeLoop
			}
		case <-s.ctx.Done():
			// Likely got sigterm. Just close. Ideally gracefully.
			break closeLoop
		}
	}
	// Now close the pipe that it gets input from.
	return workError
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
		runID := runtimeMessage.RunID
		switch runtimeMessage.MessageID {
		case MessageTypeWorkStart:
			var workStartMsg WorkStartMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &workStartMsg); err != nil {
				s.workDone <- ServerError{
					RunID:       "",
					Err:         fmt.Errorf("failed to decode work start message: %w", err),
					StepFatal:   true,
					ServerFatal: false,
				}
				continue
			}
			if runID == "" || workStartMsg.StepID == "" {
				s.workDone <- ServerError{
					RunID: "",
					Err: fmt.Errorf("missing runID (%s) or stepID in work start message (%s)",
						runID, workStartMsg.StepID),
					StepFatal:   true,
					ServerFatal: false,
				}
				continue
			}
			s.runningSteps[runID] = workStartMsg.StepID
			s.wg.Add(1) // Wait until the step is done
			go func() {
				s.runStep(runID, workStartMsg)
				s.wg.Done()
			}()
		case MessageTypeSignal:
			var signalMessage signalMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &signalMessage); err != nil {
				s.workDone <- ServerError{
					RunID:       "",
					Err:         fmt.Errorf("failed to decode signal message: %w", err),
					StepFatal:   false,
					ServerFatal: false,
				}
				continue
			}
			if runID == "" {
				s.workDone <- ServerError{
					RunID:       "",
					Err:         fmt.Errorf("RunID missing for signal '%s' in signal message", signalMessage.SignalID),
					StepFatal:   false,
					ServerFatal: false,
				}
				continue
			}
			stepID, found := s.runningSteps[runID]
			if !found {
				s.workDone <- ServerError{
					RunID:       runID,
					Err:         fmt.Errorf("unknown step with run ID '%s' in signal mesage", runID),
					StepFatal:   false,
					ServerFatal: false,
				}
				continue
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
		case MessageTypeClientDone:
			// It's now safe to close the channel
			err := s.stdinCloser.Close()
			if err != nil {
				s.workDone <- ServerError{
					RunID:       "",
					Err:         fmt.Errorf("error while closing stdin on client done: %s", err),
					StepFatal:   true,
					ServerFatal: true,
				}
			}
			return
		default:
			s.workDone <- ServerError{
				RunID: "",
				Err: fmt.Errorf("unknown message ID received: %d. This is a sign of incompatible server and client versions",
					runtimeMessage.MessageID),
				StepFatal:   false,
				ServerFatal: false,
			}
			continue
		}
	}
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
			Err:         fmt.Errorf("error while sending initial messages to client (%s)", err),
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
	outputID, outputData, err := s.pluginSchema.CallStep(s.ctx, runID, req.StepID, req.Config)
	if err != nil {
		s.workDone <- ServerError{
			RunID:       "",
			Err:         fmt.Errorf("error calling step (%s)", err),
			StepFatal:   true,
			ServerFatal: false,
		}
		return
	}
	// Lastly, send the work done message.
	err = s.sendRuntimeMessage(
		MessageTypeWorkDone,
		runID,
		workDoneMessage{
			req.StepID,
			outputID,
			outputData,
			"",
		},
	)
	if err != nil {
		// At this point, the work done message failed to send, so it's likely that sending an errorMessage would fail.
		_, err = fmt.Fprintf(os.Stderr, "error while sending work done message: %s\n", err)
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
