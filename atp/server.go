package atp

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/fxamacker/cbor/v2"
	"go.flow.arcalot.io/pluginsdk/schema"
)

// RunATPServer runs an ArcaflowTransportProtocol server with a given schema.
func RunATPServer( //nolint:funlen
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	s *schema.CallableSchema,
) error {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	workDone := make(chan struct{})
	closed := false // If this becomes a problem, switch this out for an atomic bool.
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			// The context is closed! That means this was instructed to stop.
			// First let the code know that it was closed so that it doesn't panic.
			closed = true
			// Now close the pipe that it gets input from.
			_ = stdin.Close()
		case <-workDone:
			// Done, so let it move on to the deferred wg.Done
		}
	}()
	var goroutineError error
	go func() {
		defer wg.Done()
		defer close(workDone)

		// Start by serializing the schema, since the protocol requires sending the schema on the hello message.
		serializedSchema, err := s.SelfSerialize()
		if err != nil {
			goroutineError = err
			panic(goroutineError)
		}

		// The ATP protocol uses CBOR.
		cborStdin := cbor.NewDecoder(stdin)
		cborStdout := cbor.NewEncoder(stdout)

		if closed { // Stop if closed.
			return
		}

		// First, the start message, which is just an empty message.
		var empty any
		if err := cborStdin.Decode(&empty); err != nil && !closed {
			goroutineError = fmt.Errorf("failed to CBOR-decode start output message (%w)", err)
			panic(goroutineError)
		}

		// Next, send the hello message, which includes the version and schema.
		if err := cborStdout.Encode(helloMessage{1, serializedSchema}); err != nil && !closed {
			goroutineError = fmt.Errorf("failed to CBOR-encode schema (%w)", err)
			panic(goroutineError)
		}

		// Now, get the work message that dictates which step to run and the config info.
		req := startWorkMessage{}
		if err := cborStdin.Decode(&req); err != nil && !closed {
			goroutineError = fmt.Errorf("failed to CBOR-decode start work message (%w)", err)
			panic(goroutineError)
		}
		if closed { // Stop if closed.
			return
		}
		outputID, outputData, err := s.Call(req.StepID, req.Config)
		if err != nil {
			panic(err)
		}
		if closed { // Stop if closed.
			return
		}

		// Lastly, send the work done message.
		if err := cborStdout.Encode(workDoneMessage{
			outputID,
			outputData,
			"",
		}); err != nil && !closed {
			goroutineError = fmt.Errorf("failed to encode CBOR response (%w)", err)
			panic(goroutineError)
		}
	}()

	// Keep running until both goroutines are done
	wg.Wait()
	return goroutineError
}
