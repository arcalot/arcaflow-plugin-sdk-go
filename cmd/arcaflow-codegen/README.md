# Arcaflow Engine golang code generator

The Arcaflow Engine golang code generator takes an Arcaflow schema YAML structure loaded using this library (see schema package) and generates the type definitions in Go code.

## Usage

### Standalone

Copy a valid Arcaflow schema to be used as input of the code generator into this folder and name it `schema_input.yaml`. Then run:

```
$ [ARG=object_to_ignore] go generate
```

An example for Kubernetes related schemas:
```
$ ARG=ObjectMeta go generate
``` 

The output will be stored in the `typedef_output.go` file.

### From source files

Add this line to the source files:
```
//go:generate go run go.flow.arcalot.io/pluginsdk/cmd/arcaflow-codegen/gen.go@latest schema_input.yaml
```

Optionally, you can specify objects to ignore, i.e. in a Kubernetes scenario:
```
//go:generate go run go.flow.arcalot.io/pluginsdk/cmd/arcaflow-codegen/gen.go@latest schema_input.yaml ObjectMeta
```

The output will be stored in the `typedef_output.go` file.

