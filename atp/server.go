package atp

import (
	"context"
	"io"
	"sync"

	"go.flow.arcalot.io/pluginsdk/schema"
)

// RunATPServer runs an ArcaflowTransportProtocol server with a given schema.
func RunATPServer(
	ctx context.Context,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	s schema.SchemaType,
) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	workDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = stdin.Close()
		case <-workDone:
		}
	}()
	go func() {
		defer close(workDone)

	}()
	wg.Wait()
}
