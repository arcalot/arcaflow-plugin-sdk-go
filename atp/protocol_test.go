package atp_test

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"

	"go.flow.arcalot.io/pluginsdk/atp"
	"go.flow.arcalot.io/pluginsdk/schema"
)

type helloWorldInput struct {
	Name string `json:"name"`
}

type helloWorldOutput struct {
	Message string `json:"message"`
}

func helloWorldHandler(input helloWorldInput) (string, any) {
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

func TestProtocol(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var testError error

	go func() {
		defer wg.Done()
		defer cancel()

		if err := atp.RunATPServer(
			ctx,
			stdinReader,
			stdoutWriter,
			helloWorldSchema,
		); err != nil {
			testError = err
		}
	}()
	go func() {
		defer wg.Done()
		defer cancel()

		cli := atp.NewClient(channel{
			Reader:  stdoutReader,
			Writer:  stdinWriter,
			Context: ctx,
			cancel:  cancel,
		})

		_, err := cli.ReadSchema()
		if err != nil {
			testError = err
			return
		}
		outputID, outputData, _ := cli.Execute(ctx, "hello-world", map[string]any{"name": "Arca Lot"})
		if outputID != "success" {
			testError = fmt.Errorf("Invalid output ID: %s", outputID)
			return
		}
		if outputMessage := outputData.(map[any]any)["message"].(string); outputMessage != "Hello, Arca Lot!" {
			testError = fmt.Errorf("Invalid output message: %s", outputMessage)
			return
		}
	}()
	wg.Wait()
	if testError != nil {
		t.Fatal(testError)
	}
}
