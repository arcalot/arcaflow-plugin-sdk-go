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
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			_ = stdin.Close()
		case <-workDone:
		}
	}()
	var goroutineError error
	go func() {
		defer wg.Done()
		defer close(workDone)

		serializedSchema, err := s.SelfSerialize()
		if err != nil {
			goroutineError = err
			panic(goroutineError)
		}

		cborStdin := cbor.NewDecoder(stdin)
		cborStdout := cbor.NewEncoder(stdout)

		var empty any
		if err := cborStdin.Decode(&empty); err != nil {
			goroutineError = fmt.Errorf("failed to CBOR-decode start output message (%w)", err)
			panic(goroutineError)
		}

		if err := cborStdout.Encode(helloMessage{1, serializedSchema}); err != nil {
			goroutineError = fmt.Errorf("failed to CBOR-encode schema (%w)", err)
			panic(goroutineError)
		}

		req := startWorkMessage{}
		if err := cborStdin.Decode(&req); err != nil {
			goroutineError = fmt.Errorf("failed to CBOR-decode message (%w)", err)
			panic(goroutineError)
		}
		outputID, outputData, err := s.Call(req.StepID, req.Config)
		if err != nil {
			panic(err)
		}
		if err := cborStdout.Encode(workDoneMessage{
			outputID,
			outputData,
			"",
		}); err != nil {
			goroutineError = fmt.Errorf("failed to encode CBOR response (%w)", err)
			panic(goroutineError)
		}
	}()
	wg.Wait()
	return goroutineError
}
