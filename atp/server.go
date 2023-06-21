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
	"time"
)

// RunATPServer runs an ArcaflowTransportProtocol server with a given schema.
func RunATPServer( //nolint:funlen
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	s *schema.CallablePluginSchema,
) error {
	subCtx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	workDone := make(chan error, 1)
	var workError error
	go func() {
		defer wg.Done()
		// Wait for work done or context complete.
		select {
		case workError = <-workDone:
		case <-subCtx.Done():
			// Wait up to 20 seconds for work to finish.
			// This context is the same one that's passed into the step. So now we need to wait for it to finish,
			// or exit early.
			// Exiting too early will result in the client (usually the engine's plugin provider) erroring out
			// due to the pipe being closed unexpectedly.
			select {
			case workError = <-workDone:
			case <-time.After(time.Duration(20) * time.Second):
			}
		}
		// Now close the pipe that it gets input from.
		_ = stdin.Close()
	}()
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

	go func() {
		defer wg.Done()
		defer close(workDone)

		// Start by serializing the schema, since the protocol requires sending the schema on the hello message.
		serializedSchema, err := s.SelfSerialize()
		if err != nil {
			workDone <- err
			return
		}

		// The ATP protocol uses CBOR.
		cborStdin := cbor.NewDecoder(stdin)
		cborStdout := cbor.NewEncoder(stdout)

		// First, the start message, which is just an empty message.
		var empty any
		err = cborStdin.Decode(&empty)
		if err != nil {
			workDone <- fmt.Errorf("failed to CBOR-decode start output message (%w)", err)
			return
		}

		// Next, send the hello message, which includes the version and schema.
		err = cborStdout.Encode(HelloMessage{ProtocolVersion, serializedSchema})
		if err != nil {
			workDone <- fmt.Errorf("failed to CBOR-encode schema (%w)", err)
			return
		}

		// Now, get the work message that dictates which step to run and the config info.
		req := StartWorkMessage{}
		err = cborStdin.Decode(&req)
		if err != nil {
			workDone <- fmt.Errorf("failed to CBOR-decode start work message (%w)", err)
			return
		}

		outputID, outputData, err := s.CallStep(subCtx, req.StepID, req.Config)
		if err != nil {
			workDone <- err
			return
		}

		// Lastly, send the work done message.
		messageData, err := cbor.Marshal(workDoneMessage{
			outputID,
			outputData,
			"",
		})
		if err != nil {
			workDone <- fmt.Errorf("failed to encode CBOR response data (%w)", err)
			return
		}
		err = cborStdout.Encode(
			DecodedRuntimeMessage{
				MessageTypeWorkDone,
				messageData,
			},
		)
		if err != nil {
			workDone <- fmt.Errorf("failed to encode CBOR response (%w)", err)
			return
		}

		// finished with no error!
		workDone <- nil
	}()

	// Keep running until both goroutines are done
	wg.Wait()
	return workError
}
