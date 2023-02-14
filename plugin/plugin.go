package plugin

import (
	"context"
	"fmt"
	"os"

	"go.flow.arcalot.io/pluginsdk/atp"
	"go.flow.arcalot.io/pluginsdk/schema"

	"gopkg.in/yaml.v3"
)

func print_usage() {
	fmt.Println("At least one of --atp, --schema, or --json-schema must be specified")
	fmt.Println("--atp runs the ATP server to interface with the arcaflow engine.")
	fmt.Println("--schema outputs the arcaflow schema of the plugin as YAML")
	fmt.Println("--json-schema outputs the schema of a specific step's input or output" +
		" according to standardized formats for use with other applications, like" +
		" editors for code autocompletion.")
}

// The run interface for a plugin.
// This is not required, but is recommended for standardization
// of the interface between plugins.
// Allows running ATP or exporting schema.
func Run(s *schema.CallableSchema) {
	if len(os.Args) != 2 {
		print_usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "--atp":
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := atp.RunATPServer(ctx, os.Stdin, os.Stdout, s); err != nil {
			panic(err)
		}
	case "--schema":
		serialized_schema, err := s.SelfSerialize()
		if err != nil {
			_, _ = os.Stderr.WriteString("Error while serializing schema.\n")
			os.Exit(1)
		}
		as_yaml_bytes, err := yaml.Marshal(serialized_schema)
		if err != nil {
			_, _ = os.Stderr.WriteString("Error while marshaling schema to YAML.\n")
			os.Exit(1)
		}
		fmt.Printf("serialized_schema: %v\n", string(as_yaml_bytes))
	case "--json-schema":
		_, _ = os.Stderr.WriteString("Json schema currently isn't supported by the Go SDK plugins.\n")
		os.Exit(1)
	default:
		_, _ = os.Stderr.WriteString(fmt.Sprintf("\"%s\" is not a supported input.\n", os.Args[1]))
		print_usage()
		os.Exit(1)
	}
}
