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
	workDone := make(chan error, 1)
	var workError error
	go func() {
		defer wg.Done()
		select {
		case workError = <-workDone:
			_ = stdin.Close()
		case <-ctx.Done():
			// Now close the pipe that it gets input from.
			_ = stdin.Close()
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
		err = cborStdout.Encode(HelloMessage{1, serializedSchema})
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

		outputID, outputData, err := s.Call(ctx, req.StepID, req.Config)
		if err != nil {
			workDone <- err
			return
		}

		// Lastly, send the work done message.
		err = cborStdout.Encode(workDoneMessage{
			outputID,
			outputData,
			"",
		})
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
