package atp_test

import (
	"context"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"sync"

	"go.arcalot.io/assert"
	"go.arcalot.io/log/v2"
	"go.flow.arcalot.io/pluginsdk/atp"
	"go.flow.arcalot.io/pluginsdk/schema"
	"io"
	"testing"
)

type helloWorldInput struct {
	Name string `json:"name"`
}

type helloWorldOutput struct {
	Message string `json:"message"`
}

func helloWorldHandler(_ context.Context, input helloWorldInput) (string, any) {
	return "success", helloWorldOutput{
		Message: fmt.Sprintf("Hello, %s!", input.Name),
	}
}

var helloWorldSchema = schema.NewCallableSchema(
	schema.NewCallableStep[helloWorldInput](
		"hello-world",
		schema.NewScopeSchema(
			schema.NewStructMappedObjectSchema[helloWorldInput](
				"Input",
				map[string]*schema.PropertySchema{
					"name": schema.NewPropertySchema(
						schema.NewStringSchema(nil, nil, nil),
						nil,
						true,
						nil,
						nil,
						nil,
						nil,
						nil,
					),
				},
			),
		),
		map[string]*schema.StepOutputSchema{
			"success": schema.NewStepOutputSchema(
				schema.NewScopeSchema(
					schema.NewStructMappedObjectSchema[helloWorldOutput](
						"Output",
						map[string]*schema.PropertySchema{
							"message": schema.NewPropertySchema(
								schema.NewStringSchema(nil, nil, nil),
								nil,
								true,
								nil,
								nil,
								nil,
								nil,
								nil,
							),
						},
					),
				),
				nil,
				false,
			),
		},
		nil,
		helloWorldHandler,
	),
)

type channel struct {
	io.Reader
	io.Writer
	context.Context
	cancel func()
}

func (c channel) Close() error {
	c.cancel()
	return nil
}

func TestProtocol_Client_Execute(t *testing.T) {
	// Client ReadSchema and Execute happy path.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	go func() {
		defer wg.Done()
		assert.NoError(t, atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		))
	}()

	go func() {
		defer wg.Done()
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  cancel,
		}, log.NewTestLogger(t))

		_, err := cli.ReadSchema(nil)
		assert.NoError(t, err)

		outputID, outputData, err := cli.Execute(ctx, "hello-world", map[string]any{"name": "Arca Lot"})
		assert.NoError(t, err)
		assert.Equals(t, outputID, "success")
		assert.Equals(t, outputData.(map[any]any)["message"].(string), "Hello, Arca Lot!")
	}()

	wg.Wait()
}

func TestProtocol_Client_ReadSchema(t *testing.T) {
	// Client ReadSchema happy path.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer cancel()

	go func() {
		defer wg.Done()
		assert.NoError(t, atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		))
	}()

	go func() {
		// terminate the protocol execution
		// because it will not be completed
		defer cancel()
		defer wg.Done()
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  cancel,
		}, log.NewTestLogger(t))
		_, err := cli.ReadSchema(nil)
		assert.NoError(t, err)
	}()

	wg.Wait()
}

func TestProtocol_Error_Client_StartOutput(t *testing.T) {
	// Induce error on client's encoding of start output message
	// by closing the client's cbor encoder's io pipe, stdinWriter.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	_, stdinWriter := io.Pipe()
	stdoutReader, _ := io.Pipe()
	defer cancel()

	cli := atp.NewClientWithLogger(channel{
		Reader:  stdoutReader,
		Writer:  stdinWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	// close client's cbor encoder's io pipe
	assert.NoError(t, stdinWriter.Close())

	go func() {
		defer wg.Done()
		_, err := cli.ReadSchema(nil)
		assert.Error(t, err)
	}()

	wg.Wait()
}

func TestProtocol_Error_Server_StartOutput(t *testing.T) {
	// Induce error on server's decoding of start output message
	// by closing the server's cbor decoder's io pipe, stdinReader.
	stdinReader, _ := io.Pipe()
	_, stdoutWriter := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// close the server's cbor decoder's io pipe
	assert.NoError(t, stdinReader.Close())

	err := atp.RunATPServer(
		ctx,
		stdinReader,
		stdoutWriter,
		helloWorldSchema,
	)

	assert.Error(t, err)
}

func TestProtocol_Error_Client_Hello(t *testing.T) {
	// Induce error at client's decoding of hello message
	// by closing the client's cbor decoder's io pipe,
	// stdoutReader.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer cancel()

	srvr := newATPServer(channel{
		Reader:  stdinReader,
		Writer:  stdoutWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	cli := atp.NewClientWithLogger(channel{
		Reader:  stdoutReader,
		Writer:  stdinWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	// close client's cbor decoder's io pipe
	assert.NoError(t, stdoutReader.Close())

	go func() {
		defer wg.Done()
		_, err := cli.ReadSchema(nil)
		assert.Error(t, err)
	}()

	var empty any
	assert.NoError(t, srvr.decoder.Decode(&empty))

	wg.Wait()
}

func TestProtocol_Error_Server_Hello(t *testing.T) {
	// Induce error on server's encoding of the hello message
	// by closing the server's cbor encoder's io pipe.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wgcli := &sync.WaitGroup{}
	wgcli.Add(1)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer cancel()

	cli := atp.NewClientWithLogger(channel{
		Reader:  stdoutReader,
		Writer:  stdinWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	var test_error error

	go func() {
		defer wg.Done()
		err := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)

		if err != nil {
			test_error = err
		}
	}()

	go func() {
		defer wgcli.Done()
		assert.NoError(t, cli.Encoder().Encode(nil))

		// close server's cbor encoder's io pipe
		assert.NoError(t, stdoutWriter.Close())
	}()
	wgcli.Wait()

	wg.Wait()
	assert.Error(t, test_error)
}

func TestProtocol_Error_Server_WorkStart(t *testing.T) {
	// Induce error on server's decoding of the start work
	// message by closing the server's cbor decoder's
	// io pipe.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer cancel()

	cli := atp.NewClientWithLogger(channel{
		Reader:  stdoutReader,
		Writer:  stdinWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	var test_error error
	go func() {
		defer wg.Done()
		err := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)

		if err != nil {
			test_error = err
		}
	}()

	go func() {
		defer wg.Done()
		assert.NoError(t, cli.Encoder().Encode(nil))
		var hello atp.HelloMessage
		assert.NoError(t, cli.Decoder().Decode(&hello))

		// close the server's cbor decoder's io pipe
		assert.NoError(t, stdinReader.Close())
	}()

	wg.Wait()
	assert.Error(t, test_error)
}

func TestProtocol_Error_Client_WorkStart(t *testing.T) {
	// Induce error on client's (and server incidentally)
	// start work message by closing the client's cbor
	// encoder's io pipe, stdinWriter, before the client
	// executes a step.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wgcli := &sync.WaitGroup{}
	wgcli.Add(1)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer cancel()

	cli := atp.NewClientWithLogger(channel{
		Reader:  stdoutReader,
		Writer:  stdinWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	var srvr_error error
	var cli_error error
	go func() {
		defer wg.Done()
		err := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)
		if err != nil {
			srvr_error = err
		}
	}()

	go func() {
		defer wgcli.Done()
		_, err := cli.ReadSchema(nil)
		assert.NoError(t, err)

		// close client's cbor encoder's io pipe
		assert.NoError(t, stdinWriter.Close())

		_, _, err = cli.Execute(ctx, "hello-world", map[string]any{"name": "Arca Lot"})
		if err != nil {
			cli_error = err
		}
	}()
	wgcli.Wait()

	wg.Wait()
	assert.Error(t, srvr_error)
	assert.Error(t, cli_error)
}

func TestProtocol_Error_Client_WorkDone(t *testing.T) {
	// Induce error on client cbor decoding of work done
	// message by closing the client's cbor decoder's io pipe,
	// stdoutReader, after the server decodes the start work
	// message.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer cancel()

	srvr := newATPServer(channel{
		Reader:  stdinReader,
		Writer:  stdoutWriter,
		Context: ctx,
		cancel:  cancel,
	},
		log.NewTestLogger(t))

	cli := atp.NewClientWithLogger(channel{
		Reader:  stdoutReader,
		Writer:  stdinWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	var srvr_error error
	var cli_error error

	go func() {
		defer wg.Done()
		req := atp.StartWorkMessage{}
		err := srvr.decoder.Decode(&req)
		if err != nil {
			srvr_error = err
		}

		// close client's cbor decoder's io pipe
		assert.NoError(t, stdoutReader.Close())
	}()

	go func() {
		defer wg.Done()
		_, _, err := cli.Execute(
			ctx, "hello-world", map[string]any{"name": "Arca Lot"})
		if err != nil {
			cli_error = err
		}
	}()

	wg.Wait()
	assert.NoError(t, srvr_error)
	assert.Error(t, cli_error)
}

func TestProtocol_Error_Server_WorkDone(t *testing.T) {
	// Induce server error when it cbor encodes the work done
	// message by closing its encoder's io pipe, stdoutWriter,
	// right before the client decodes the work done message
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	defer cancel()

	cli := atp.NewClientWithLogger(channel{
		Reader:  stdoutReader,
		Writer:  stdinWriter,
		Context: ctx,
		cancel:  cancel,
	}, log.NewTestLogger(t))

	var srvr_error error
	var cli_error error

	go func() {
		defer wg.Done()
		err := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)
		if err != nil {
			srvr_error = err
		}
	}()

	go func() {
		defer wg.Done()
		_, err := cli.ReadSchema(nil)
		assert.NoError(t, err)

		// close server's cbor encoder's io pipe
		assert.NoError(t, stdoutWriter.Close())

		err = cli.Encoder().Encode(atp.StartWorkMessage{
			StepID: "hello-world",
			Config: map[string]any{"name": "Arca Lot"},
		})
		if err != nil {
			cli_error = err
		}
	}()

	wg.Wait()
	assert.NoError(t, cli_error)
	assert.Error(t, srvr_error)
}

// serverChannel holds the methods to talking to an ATP server (plugin).
type serverChannel interface {
	io.Reader
	io.Writer
	io.Closer
}

type atpServer struct {
	channel serverChannel
	decMode cbor.DecMode
	logger  log.Logger
	decoder *cbor.Decoder
	encoder *cbor.Encoder
}

func newATPServer(
	channel serverChannel,
	logger log.Logger,
) atpServer {
	decMode, err := cbor.DecOptions{
		ExtraReturnErrors: cbor.ExtraDecErrorUnknownField,
	}.DecMode()
	if err != nil {
		panic(err)
	}
	if logger == nil {
		logger = log.NewLogger(log.LevelDebug, log.NewNOOPLogger())
	}
	return atpServer{
		channel,
		decMode,
		logger,
		decMode.NewDecoder(channel),
		cbor.NewEncoder(channel),
	}
}
