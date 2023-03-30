package main

import (
	"flag"
	"log"
	"os"

	"go.flow.arcalot.io/pluginsdk/codegen"
	"go.flow.arcalot.io/pluginsdk/schema"
	"gopkg.in/yaml.v3"
)

// Generator that converts a schema yaml file to schema and type definitions golang files.
// Usage of gen:
// 	gen schema_input.yaml
//go:generate go run gen.go schema_input.yaml $ARG

func main() {
	configFile := ".arcaflow-codegen.yaml"

	flag.StringVar(&configFile, "config", configFile, "Codegen configuration file")

	cfg, err := readConfigFile(configFile)
	if err != nil {
		cfg = &config{}
	}
	if cfg.PackageName == "" {
		cfg.PackageName = "main"
	}
	if cfg.TargetFile == "" {
		cfg.TargetFile = "schema_gen.go"
	}
	if cfg.SchemaFile == "" {
		cfg.SchemaFile = "schema.yaml"
	}

	pluginSchema := readSchema(cfg.SchemaFile)

	generator := codegen.New()
	content, err := generator.Generate(
		cfg.PackageName,
		pluginSchema,
		cfg.ExternalReferences,
	)
	if err != nil {
		log.Printf("failed to generate code (%v)", err)
	}
	if err := os.WriteFile(cfg.TargetFile, content, 0x600); err != nil {
		log.Fatalf("failed to write target file %s (%v)", cfg.TargetFile, err)
	}
}

func readSchema(schemaFile string) *schema.SchemaSchema {
	fh, err := os.Open(schemaFile) //nolint:gosec
	if err != nil {
		log.Printf("failed to read schema file %s (%v)", schemaFile, err)
	}
	decoder := yaml.NewDecoder(fh)
	var cfg any
	if err := decoder.Decode(&cfg); err != nil {
		log.Printf("failed to decode schema file %s (%v)", schemaFile, err)
	}
	decodedSchema, err := schema.DescribeSchema().Unserialize(cfg)
	if err != nil {
		log.Printf("failed to decode schema file %s (%v)", schemaFile, err)
	}
	return decodedSchema.(*schema.SchemaSchema)
}

func readConfigFile(configFile string) (*config, error) {
	fh, err := os.Open(configFile) //nolint:gosec
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(fh)
	cfg := &config{}
	if err := decoder.Decode(cfg); err != nil {
		log.Printf("failed to decode config file %s (%v)", configFile, err)
	}
	return cfg, nil
}

type config struct {
	PackageName        string                     `yaml:"package"`
	ExternalReferences codegen.ExternalReferences `yaml:"external"`
	TargetFile         string                     `yaml:"target"`
	SchemaFile         string                     `yaml:"schema"`
}
