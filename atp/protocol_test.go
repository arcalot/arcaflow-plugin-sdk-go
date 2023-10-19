package atp_test

import (
	"context"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"go.arcalot.io/assert"
	"go.arcalot.io/log/v2"
	"go.flow.arcalot.io/pluginsdk/atp"
	"go.flow.arcalot.io/pluginsdk/schema"
	"io"
	"sync"
	"testing"
	"time"
)

type helloWorldInput struct {
	Name string `json:"name"`
}

type helloWorldOutput struct {
	Message string `json:"message"`
}

var helloWorldInputSchema = schema.NewScopeSchema(
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
)

func helloWorldStepHandler(_ context.Context, _ any, input helloWorldInput) (string, any) {
	return "success", helloWorldOutput{
		Message: fmt.Sprintf("Hello, %s!", input.Name),
	}
}

func panickingHelloWorldStepHandler(_ context.Context, _ any, input helloWorldInput) (string, any) {
	panic("abcde")
}

func helloWorldSignalHandler(_ context.Context, test any, input helloWorldInput) {
	// Does nothing at the moment
}

var helloWorldCallableSignal = schema.NewCallableSignal(
	"hello-world-signal",
	helloWorldInputSchema,
	nil,
	helloWorldSignalHandler,
)

var helloWorldSchema = schema.NewCallableSchema(
	schema.NewCallableStepWithSignals[any, helloWorldInput](
		/* id */ "hello-world",
		/* input */ helloWorldInputSchema,
		/* outputs */ map[string]*schema.StepOutputSchema{
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
		/* signal handlers */ map[string]schema.CallableSignal{
			"hello-world-signal": helloWorldCallableSignal,
		},
		/* signal emitters */ map[string]*schema.SignalSchema{
			"hello-world-signal": helloWorldCallableSignal.ToSignalSchema(),
		},
		/* Display */ nil,
		/* Initializer */ nil,
		/* step handler */ helloWorldStepHandler,
	),
)

var panickingHelloWorldSchema = schema.NewCallableSchema(
	schema.NewCallableStepWithSignals[any, helloWorldInput](
		/* id */ "hello-world",
		/* input */ helloWorldInputSchema,
		/* outputs */ map[string]*schema.StepOutputSchema{
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
		/* signal handlers */ map[string]schema.CallableSignal{
			"hello-world-signal": helloWorldCallableSignal,
		},
		/* signal emitters */ map[string]*schema.SignalSchema{
			"hello-world-signal": helloWorldCallableSignal.ToSignalSchema(),
		},
		/* Display */ nil,
		/* Initializer */ nil,
		/* step handler */ panickingHelloWorldStepHandler,
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
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)
		assert.Equals(t, len(errors), 0)
	}()

	go func() {
		defer wg.Done()
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  cancel,
		}, log.NewTestLogger(t))

		_, err := cli.ReadSchema()
		assert.NoError(t, err)

		result := cli.Execute(
			schema.Input{
				RunID:     t.Name(),
				ID:        "hello-world",
				InputData: map[string]any{"name": "Arca Lot"},
			}, nil, nil)
		assert.NoError(t, cli.Close())
		assert.NoError(t, result.Error)
		assert.Equals(t, result.OutputID, "success")
		assert.Equals(t, result.OutputData.(map[any]any)["message"].(string), "Hello, Arca Lot!")
	}()

	wg.Wait()
}

//nolint:funlen
func TestProtocol_Client_ATP_v1(t *testing.T) {
	// Client ReadSchema and Execute atp v1 happy path.
	// This is not a fragile test because the ATP v1 is not changing. It is the legacy supported version.
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	step := "hello-world"
	stepInput := map[string]any{"name": "Arca Lot"}

	go func() {
		defer wg.Done()
		fromClient := cbor.NewDecoder(stdinReader)
		toClient := cbor.NewEncoder(stdoutWriter)
		// 1: read start output message
		var empty any
		assert.NoError(t, fromClient.Decode(&empty))
		// 2: Send hello message with version set to 1 and the hello-world schema.
		helloMessage := atp.HelloMessage{
			Version: 1,
			Schema:  assert.NoErrorR[any](t)(helloWorldSchema.SelfSerialize()),
		}
		assert.NoError(t, toClient.Encode(&helloMessage))
		// 3: Read work start message
		var workStartMsg atp.WorkStartMessage
		assert.NoError(t, fromClient.Decode(&workStartMsg))
		assert.Equals(t, workStartMsg.StepID, step)
		unserializedInput := assert.NoErrorR[any](t)(helloWorldInputSchema.Unserialize(workStartMsg.Config))
		assert.Equals(t, unserializedInput.(helloWorldInput), helloWorldInput{Name: "Arca Lot"})

		// 4: Send work done message
		workDoneMessage := atp.WorkDoneMessage{
			StepID:     step,
			OutputID:   "success",
			OutputData: map[string]string{"message": "Hello, Arca Lot!"},
			DebugLogs:  "",
		}
		assert.NoError(t, toClient.Encode(&workDoneMessage))

	}()

	go func() {
		defer wg.Done()
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  nil,
		}, log.NewTestLogger(t))

		_, err := cli.ReadSchema()
		assert.NoError(t, err)

		result := cli.Execute(
			schema.Input{
				RunID:     t.Name(),
				ID:        step,
				InputData: stepInput,
			}, nil, nil)
		assert.NoError(t, cli.Close())
		assert.NoError(t, result.Error)
		assert.Equals(t, result.OutputID, "success")
		assert.Equals(t, result.OutputData.(map[any]any)["message"].(string), "Hello, Arca Lot!")
	}()

	wg.Wait()
}

func TestProtocol_Client_Execute_Panicking(t *testing.T) {
	// Client ReadSchema and Execute happy path.
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	go func() {
		defer wg.Done()
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			panickingHelloWorldSchema,
		)
		assert.Equals(t, len(errors), 2)
	}()

	go func() {
		defer wg.Done()
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  cancel,
		}, log.NewTestLogger(t))

		_, err := cli.ReadSchema()
		assert.NoError(t, err)

		for _, testID := range []string{"a", "b"} {
			result := cli.Execute(
				schema.Input{
					RunID:     t.Name() + "_" + testID,
					ID:        "hello-world",
					InputData: map[string]any{"name": "Arca Lot"},
				}, nil, nil)
			assert.Error(t, result.Error)
			assert.Contains(t, result.Error.Error(), "abcde")
			assert.Equals(t, result.OutputID, "")
		}
		assert.NoError(t, cli.Close())
	}()

	wg.Wait()
}

func TestProtocol_Client_Execute_Multi_Step_Parallel(t *testing.T) {
	// Runs several steps on one client instance at the same time
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	go func() {
		defer wg.Done()
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)
		assert.Equals(t, len(errors), 0)
	}()

	go func() {
		defer wg.Done()
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  cancel,
		}, log.NewTestLogger(t))

		_, err := cli.ReadSchema()
		assert.NoError(t, err)

		names := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
		stepWg := &sync.WaitGroup{}
		for _, name := range names {
			stepName := name
			stepWg.Add(1)
			go func() {
				defer stepWg.Done()
				result := cli.Execute(
					schema.Input{
						RunID:     t.Name() + "_" + stepName, // Must be unique
						ID:        "hello-world",
						InputData: map[string]any{"name": stepName},
					}, nil, nil)
				assert.NoError(t, result.Error)
				assert.Equals(t, result.OutputID, "success")
				assert.Equals(t, result.OutputData.(map[any]any)["message"].(string), "Hello, "+stepName+"!")
			}()
		}
		stepWg.Wait()
		assert.NoError(t, cli.Close())
	}()

	wg.Wait()
}
func TestProtocol_Client_Execute_Multi_Step_Serial(t *testing.T) {
	// Runs several steps in one client, but with a long enough delay for each one to finish up
	// before the next one runs
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	go func() {
		defer wg.Done()
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)
		assert.Equals(t, len(errors), 0)
	}()

	go func() {
		defer wg.Done()
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  cancel,
		}, log.NewTestLogger(t))

		_, err := cli.ReadSchema()
		assert.NoError(t, err)

		names := []string{"a", "b", "c"}
		waitTime := 0
		stepWg := &sync.WaitGroup{}
		for _, name := range names {
			stepName := name
			stepWg.Add(1)
			stepWaitTime := waitTime
			waitTime += 5
			go func() {
				defer stepWg.Done()
				time.Sleep(time.Duration(stepWaitTime) * time.Millisecond)
				result := cli.Execute(
					schema.Input{
						RunID:     t.Name() + "_" + stepName, // Must be unique
						ID:        "hello-world",
						InputData: map[string]any{"name": stepName},
					}, nil, nil)
				assert.NoError(t, result.Error)
				assert.Equals(t, result.OutputID, "success")
				assert.Equals(t, result.OutputData.(map[any]any)["message"].(string), "Hello, "+stepName+"!")
			}()
		}
		stepWg.Wait()
		assert.NoError(t, cli.Close())
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
		t.Logf("Starting ATP server")
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)
		assert.Equals(t, len(errors), 0)
		t.Logf("ATP server exited without error")

	}()

	go func() {
		// terminate the protocol execution
		// because it will not be completed
		defer func() {
			cancel()
			wg.Done()
		}()
		t.Logf("Starting client.")
		cli := atp.NewClientWithLogger(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: nil,
			cancel:  cancel,
		}, log.NewTestLogger(t))
		_, err := cli.ReadSchema()
		err2 := cli.Close()
		assert.NoError(t, err)
		assert.NoError(t, err2)
		t.Logf("Client exited without error")
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
		_, err := cli.ReadSchema()
		assert.Error(t, err)
	}()

	wg.Wait()

	// We don't wait on error, to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
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
	assert.NoError(t, stdoutWriter.Close()) // Close this, because it's unbuffered, so it would deadlock.

	serverErrors := atp.RunATPServer(
		ctx,
		stdinReader,
		stdoutWriter,
		helloWorldSchema,
	)

	assert.NotNil(t, serverErrors)
	assert.Equals(t, len(serverErrors), 1)
	assert.Equals(t, serverErrors[0].ServerFatal, true)

	// We don't wait on error, to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
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
		_, err := cli.ReadSchema()
		assert.Error(t, err)
	}()

	var empty any
	assert.NoError(t, srvr.decoder.Decode(&empty))

	wg.Wait()

	// We don't wait on error, to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
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

	var serverErrors []*atp.ServerError

	go func() {
		defer wg.Done()
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)

		serverErrors = errors
	}()

	go func() {
		defer wgcli.Done()
		assert.NoError(t, cli.Encoder().Encode(nil))

		// close server's cbor encoder's io pipe
		assert.NoError(t, stdoutWriter.Close())
	}()
	wgcli.Wait()

	wg.Wait()
	assert.NotNil(t, serverErrors)
	assert.Equals(t, len(serverErrors), 1)
	// We don't wait on error, to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
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

	var serverErrors []*atp.ServerError
	go func() {
		defer wg.Done()
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)

		serverErrors = errors
	}()

	go func() {
		defer wg.Done()
		assert.NoError(t, cli.Encoder().Encode(nil))
		var hello atp.HelloMessage
		assert.NoError(t, cli.Decoder().Decode(&hello))

		// close the server's cbor decoder's io pipe
		assert.NoError(t, stdinReader.Close())
		// Now close the client's stdoutWriter, since it would otherwise deadlock.
		assert.NoError(t, stdoutWriter.Close())
	}()

	wg.Wait()
	assert.NotNil(t, serverErrors)
	assert.Equals(t, len(serverErrors), 1)
	// This may make the test more fragile, but checking the error is the only way
	// to know that the error is from where we're testing.
	assert.Contains(t, serverErrors[0].Err.Error(), "failed to read or decode runtime message")
	// We don't wait on error, to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
}

//nolint:funlen
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

	var serverErrors []*atp.ServerError
	var cli_error error
	go func() {
		defer wg.Done()
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)

		serverErrors = errors
	}()

	go func() {
		defer wgcli.Done()
		_, err := cli.ReadSchema()
		assert.NoError(t, err)

		// close client's cbor encoder's io pipe. This is intentionally done incorrectly to cause an error.
		assert.NoError(t, stdinWriter.Close())

		result := cli.Execute(
			schema.Input{
				RunID:     t.Name(),
				ID:        "hello-world",
				InputData: map[string]any{"name": "Arca Lot"},
			}, nil, nil)
		assert.Error(t, cli.Close())
		if result.Error != nil {
			cli_error = result.Error
		}
		// Close the other pipe after to unblock the server
		assert.NoError(t, stdoutWriter.Close())
	}()
	wgcli.Wait()

	wg.Wait()
	assert.NotNil(t, serverErrors)
	assert.Equals(t, len(serverErrors), 1)
	assert.Error(t, cli_error)
	// We don't lock on error to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
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

	atpServer := newATPServer(channel{
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
		req := atp.WorkStartMessage{}
		err := atpServer.decoder.Decode(&req)
		if err != nil {
			srvr_error = err
		}

		// close client's cbor decoder's io pipe
		assert.NoError(t, stdoutReader.Close())
	}()

	go func() {
		defer wg.Done()
		result := cli.Execute(
			schema.Input{
				RunID:     t.Name(),
				ID:        "hello-world",
				InputData: map[string]any{"name": "Arca Lot"},
			}, nil, nil)
		if result.Error != nil {
			cli_error = result.Error
		}
		assert.NoError(t, cli.Close())
	}()
	wg.Wait()
	assert.NoError(t, srvr_error)
	assert.Error(t, cli_error)
	// We don't wait on error, to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
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

	var serverErrors []*atp.ServerError
	var cli_error error

	go func() {
		defer wg.Done()
		errors := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		)
		serverErrors = errors
	}()

	go func() {
		defer wg.Done()
		_, err := cli.ReadSchema()
		assert.NoError(t, err)

		// close server's cbor encoder's io pipe
		assert.NoError(t, stdoutWriter.Close())

		err = cli.Encoder().Encode(atp.WorkStartMessage{
			StepID: "hello-world",
			Config: map[string]any{"name": "Arca Lot"},
		})
		if err != nil {
			cli_error = err
		}
	}()

	wg.Wait()
	assert.NoError(t, cli_error)
	assert.NotNil(t, serverErrors)
	assert.Equals(t, len(serverErrors), 1)

	// We don't wait on error, to prevent deadlocks, so just sleep
	time.Sleep(time.Millisecond * 2)
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
