# Arcaflow plugin code generator

The Arcaflow Engine golang code generator takes an Arcaflow schema YAML structure loaded using this library (see schema package) and generates the type definitions in Go code.

## Usage

Create your schema file as `schema.yaml` and then add the following to your Go code:

```go
//go:generate go run go.flow.arcalot.io/pluginsdk/cmd/arcaflow-codegen/gen.go@latest
```

You can now run `go generate` to generate your Go code.

## Customizing

You can customize the code generating behavior by specifying one of the following options in `.arcaflow-codegen.yaml`:

| Option     | Default         | Description                     |
|------------|-----------------|---------------------------------|
| `schema`   | `schema.yaml`   | Name of the schema file         |
| `target`   | `schema_gen.go` | Name of the target file         |
| `package`  | `main`          | Name of the generated package   |
| `external` |                 | External references (see below) |

## External references (to be implemented)

Sometimes, you will want to include schemas and data structures from external libraries. You can define the external references in the `arcaflow-codegen.yaml` file as follows:

```yaml
external:
  volume: # ID for the external. This is how you can reference the object in a "ref" type
    schemaSource: getKubernetesVolumeSchema() # this is Go code to return the schema for the external.
    package: k8s.io/api/core/v1 # This is the Go package to reference
    type: Volume # The struct to reference.
```
