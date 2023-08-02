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
) error {
	session := initializeATPServerSession(ctx, stdin, stdout, pluginSchema)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Run needs to be run in its own goroutine to allow for the closure handling to happen simultaneously.
	go func() {
		session.run(wg)
	}()

	workError := session.handleClosure(stdin)

	// Ensure that the session is done.
	wg.Wait()
	return workError
}

type atpServerSession struct {
	ctx          context.Context
	cancel       *context.CancelFunc
	req          StartWorkMessage
	cborStdin    *cbor.Decoder
	cborStdout   *cbor.Encoder
	workDone     chan error
	doneChannel  chan bool
	pluginSchema *schema.CallableSchema
}

func initializeATPServerSession(
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	pluginSchema *schema.CallableSchema,
) *atpServerSession {
	subCtx, cancel := context.WithCancel(ctx)
	workDone := make(chan error, 1)
	// The ATP protocol uses CBOR.
	cborStdin := cbor.NewDecoder(stdin)
	cborStdout := cbor.NewEncoder(stdout)
	doneChannel := make(chan bool, 1) // Buffer to prevent it from hanging if something unexpected happens.

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
		ctx:          subCtx,
		cancel:       &cancel,
		req:          StartWorkMessage{},
		cborStdin:    cborStdin,
		cborStdout:   cborStdout,
		workDone:     workDone,
		doneChannel:  doneChannel,
		pluginSchema: pluginSchema,
	}
}

func (s *atpServerSession) handleClosure(stdin io.ReadCloser) error {
	// Wait for work done or context complete.
	var workError error
	select {
	case workError = <-s.workDone:
	case <-s.ctx.Done():
		// Likely got sigterm. Just close. Ideally gracefully.
	}
	// Now close the pipe that it gets input from.
	_ = stdin.Close()
	return workError
}

func (s *atpServerSession) runATPReadLoop() {
	// The message is generic, so we must find the type and decode the full message next.
	var runtimeMessage DecodedRuntimeMessage
	for {
		// First, decode the message
		if err := s.cborStdin.Decode(&runtimeMessage); err != nil {
			// Failed to decode. If it's done, that's okay. If not, there's a problem.
			done := false
			select {
			case done = <-s.doneChannel:
			default:
				// Prevents it from blocking
			}
			if !done {
				s.workDone <- fmt.Errorf("failed to read or decode runtime message: %w", err)
			}
			return
		}
		switch runtimeMessage.MessageID {
		case MessageTypeSignal:
			var signalMessage signalMessage
			if err := cbor.Unmarshal(runtimeMessage.RawMessageData, &signalMessage); err != nil {
				s.workDone <- fmt.Errorf("failed to decode signal message: %w", err)
			}
			if s.req.StepID != signalMessage.StepID {
				s.workDone <- fmt.Errorf("signal sent with mismatched step ID, got %s, expected %s",
					signalMessage.StepID, s.req.StepID)
			}
			if err := s.pluginSchema.CallSignal(s.ctx, signalMessage.StepID, signalMessage.SignalID, signalMessage.Data); err != nil {
				s.workDone <- fmt.Errorf("failed while running signal ID %s: %w",
					signalMessage.SignalID, err)
			}
		default:
			s.workDone <- fmt.Errorf("unknown message ID received: %d", runtimeMessage.MessageID)
		}
	}
}

func (s *atpServerSession) run(wg *sync.WaitGroup) {
	defer func() {
		s.doneChannel <- true
		close(s.workDone)
		wg.Done()
	}()

	err := s.sendInitialMessagesToClient()
	if err != nil {
		s.workDone <- err
		return
	}

	// Now, get the work message that dictates which step to run and the config info.
	err = s.cborStdin.Decode(&s.req)
	if err != nil {
		s.workDone <- fmt.Errorf("failed to CBOR-decode start work message (%w)", err)
		return
	}

	// Now, loop through stdin inputs until the step ends.
	go func() { // Listen for signals in another thread
		s.runATPReadLoop()
	}()

	// Call the step in the provided callable schema.
	outputID, outputData, err := s.pluginSchema.CallStep(s.ctx, s.req.StepID, s.req.Config)
	if err != nil {
		s.workDone <- err
		return
	}

	// Lastly, send the work done message.
	err = s.cborStdout.Encode(
		RuntimeMessage{
			MessageTypeWorkDone,
			workDoneMessage{
				outputID,
				outputData,
				"",
			},
		},
	)
	if err != nil {
		s.workDone <- fmt.Errorf("failed to encode CBOR response (%w)", err)
		return
	}

	// finished with no error!
	s.workDone <- nil
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
