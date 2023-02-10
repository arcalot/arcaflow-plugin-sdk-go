package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.flow.arcalot.io/pluginsdk/atp"
	"go.flow.arcalot.io/pluginsdk/schema"

	"gopkg.in/yaml.v3"
)

func print_usage() {
	fmt.Printf("At least one of --atp or --schema must be specified\n")
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
	if os.Args[1] == "--atp" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := atp.RunATPServer(ctx, os.Stdin, os.Stdout, s); err != nil {
			panic(err)
		}
	} else if os.Args[1] == "--schema" {
		as_yaml_bytes, err := yaml.Marshal(s)
		if err != nil {
			_, _ = os.Stderr.WriteString("Error while marshaling schema to YAML.\n")
			os.Exit(1)
		}
		fmt.Printf("serialized_schema: %v\n", string(as_yaml_bytes))
	} else if os.Args[1] == "--json-schema" {
		as_yaml_bytes, err := json.MarshalIndent(s, "", "    ")
		if err != nil {
			_, _ = os.Stderr.WriteString("Error while marshaling schema to JSON.\n")
			os.Exit(1)
		}
		fmt.Printf("serialized_schema: %v\n", string(as_yaml_bytes))
	} else {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("\"%s\" is not a supported input.\n", os.Args[1]))
		print_usage()
		os.Exit(1)
	}
}
