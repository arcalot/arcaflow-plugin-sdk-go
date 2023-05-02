package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"go.flow.arcalot.io/pluginsdk/schema"
)

// ExternalReference describes a reference to an external library that should not be generated. The specified ID can be
// used in references and must not actually be described in the scope. Instead, the SchemaSource (go code) will be
// placed into the schema as a way to obtain the schema. The Package and Type variables will indicate the Go type that
// needs to be imported.
type ExternalReference struct {
	SchemaSource string `yaml:"schemaSource"`
	Package      string `yaml:"package"`
	Type         string `yaml:"type"`
}

type ExternalReferences map[string]ExternalReference

func (e ExternalReferences) Packages() []string {
	packages := map[string]struct{}{}
	for _, ref := range e {
		packages[ref.Package] = struct{}{}
	}
	packageList := make([]string, len(packages))
	i := 0
	for p := range packages {
		packageList[i] = p
		i++
	}
	return packageList
}

// CodeGen is the Arcaflow plugin code generator.
type CodeGen interface {
	// Generate generates the code with the specified plugin schema and imports it.
	Generate(
		packageName string,
		pluginSchema schema.Schema[schema.Step],
		// externalReferences contains a map of objects to external types.
		externalReferences ExternalReferences,
	) ([]byte, error)
}

// New returns a new code generator.
func New() CodeGen {
	return &codeGen{}
}

//go:embed templates/*
var templates embed.FS

type codeGen struct {
}

type templateScope struct {
	PackageName        string
	PluginSchema       schema.Schema[schema.Step]
	ExternalReferences ExternalReferences
}

var reNameReplace = regexp.MustCompile(`(^|[-_])([a-zA-Z])`)
var reHandlerNameReplace = regexp.MustCompile(`[-_]([a-zA-Z])`)
var rePackage = regexp.MustCompile(`.+\.`)

//nolint:funlen
func (c *codeGen) Generate(
	packageName string,
	pluginSchema schema.Schema[schema.Step],
	externalReferences ExternalReferences,
) ([]byte, error) {
	tpl := template.New("templates/plugin.go.tpl")
	imports := map[string]struct{}{}
	nameReplace := func(id string) string {
		return reNameReplace.ReplaceAllStringFunc(id, func(i string) string {
			switch i[0] {
			case '-':
				fallthrough
			case '_':
				return strings.ToUpper(string([]byte{i[1]}))
			default:
				return strings.ToUpper(i)
			}

		})
	}
	tpl = tpl.Funcs(template.FuncMap{
		"structName": nameReplace,
		"property":   nameReplace,
		"handlerFuncName": func(id string) string {
			return reHandlerNameReplace.ReplaceAllStringFunc(id, func(i string) string {
				return strings.ToUpper(string([]byte{i[1]}))
			})
		},
		"import": func(packageName string) string {
			imports[packageName] = struct{}{}

			return ""
		},
		"getImports": func() []string {
			importList := make([]string, len(imports))
			i := 0
			for imp := range imports {
				importList[i] = imp
				i++
			}
			return importList
		},
		"partial": func(partial string, data any) string {
			wr := &bytes.Buffer{}
			if err := tpl.ExecuteTemplate(wr, partial+".go.tpl", data); err != nil {
				panic(fmt.Errorf("failed to parse partial %s (%w)", partial, err))
			}
			return wr.String()
		},
		"shortPackage": func(packageName string) string {
			return rePackage.ReplaceAllString(packageName, "")
		},
		"escapeString": strconv.Quote,
		"escapeStringPtr": func(text *string) string {
			if text == nil {
				return "nil"
			}
			return strconv.Quote(*text)
		},
		"prefix": func(input any, prefix string) any {
			switch i := input.(type) {
			case string:
				return strings.ReplaceAll(i, "\n", "\n"+prefix)
			case *string:
				r := strings.ReplaceAll(*i, "\n", "\n"+prefix)
				return &r
			default:
				panic(fmt.Errorf("invalid input type for 'prefix': %T (%v)", input, input))
			}
		},
		"stringList": func(input []string) string {
			if input == nil {
				return "nil"
			}
			items := make([]string, len(input))
			for i, v := range input {
				items[i] = strconv.Quote(v)
			}
			return "[]string{" + strings.Join(items, ", ") + "}"
		},
	})

	fileContents, err := templates.ReadFile("templates/plugin.go.tpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read base template (%w)", err)
	}
	if tpl, err = tpl.Parse(string(fileContents)); err != nil {
		return nil, fmt.Errorf("failed to parse base template (%w)", err)
	}

	for _, scope := range []string{"code", "schema"} {
		if err := c.parsePartialDirectory(scope, tpl); err != nil {
			return nil, err
		}
	}

	buf := &bytes.Buffer{}
	if err := tpl.ExecuteTemplate(buf, "templates/plugin.go.tpl", templateScope{
		packageName,
		pluginSchema,
		externalReferences,
	}); err != nil {
		return nil, fmt.Errorf("failed to execute base template (%w)", err)
	}

	return buf.Bytes(), nil
}

func (c *codeGen) parsePartialDirectory(scope string, tpl *template.Template) error {
	partialFiles, err := templates.ReadDir("templates/" + scope)
	if err != nil {
		return fmt.Errorf("failed to read built-in partials (%w)", err)
	}
	for _, partial := range partialFiles {
		if partial.IsDir() {
			continue
		}
		fileContents, err := templates.ReadFile("templates/" + scope + "/" + partial.Name())
		if err != nil {
			return fmt.Errorf("failed to read partial %s/%s (%w)", scope, partial.Name(), err)
		}
		partialTpl := tpl.New(scope + "/" + partial.Name())
		if _, err := partialTpl.Parse(string(fileContents)); err != nil {
			return fmt.Errorf("failed to parse partial %s/%s (%w)", scope, partial.Name(), err)
		}
	}
	return nil
}
