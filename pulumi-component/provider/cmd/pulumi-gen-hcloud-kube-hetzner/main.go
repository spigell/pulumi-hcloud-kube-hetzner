package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	dotnetgen "github.com/pulumi/pulumi/pkg/v3/codegen/dotnet"
	gogen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	nodejsgen "github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	pygen "github.com/pulumi/pulumi/pkg/v3/codegen/python"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <language> <out-dir> <schema-file>\n", os.Args[0])
		os.Exit(1)
	}

	language, outdir, schemaPath := os.Args[1], os.Args[2], os.Args[3]

	err := emitSDK(language, outdir, schemaPath)
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
	}
}

func emitSDK(language, outdir, schemaPath string) error {
	pkg, err := readSchema(schemaPath)
	if err != nil {
		return err
	}

	tool := "Pulumi SDK Generator"
	extraFiles := map[string][]byte{}

	var generator func() (map[string][]byte, error)
	switch language {
	case "dotnet":
		generator = func() (map[string][]byte, error) { return dotnetgen.GeneratePackage(tool, pkg, extraFiles) }
	case "go":
		generator = func() (map[string][]byte, error) { return gogen.GeneratePackage(tool, pkg) }
	case "nodejs":
		generator = func() (map[string][]byte, error) { return nodejsgen.GeneratePackage(tool, pkg, extraFiles, nil) }
	case "python":
		generator = func() (map[string][]byte, error) { return pygen.GeneratePackage(tool, pkg, extraFiles) }
	default:
		return errors.Errorf("Unrecognized language %q", language)
	}

	files, err := generator()
	if err != nil {
		return errors.Wrapf(err, "generating %s package", language)
	}

	for f, contents := range files {
		if err := emitFile(outdir, f, contents); err != nil {
			return errors.Wrapf(err, "emitting file %v", f)
		}
	}

	return nil
}

func readSchema(schemaPath string) (*schema.Package, error) {
	schemaBytes, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading schema")
	}

	var spec schema.PackageSpec
	if err = json.Unmarshal(schemaBytes, &spec); err != nil {
		return nil, errors.Wrap(err, "unmarshalling schema")
	}

	pkg, err := schema.ImportSpec(spec, nil)
	if err != nil {
		return nil, errors.Wrap(err, "importing schema")
	}
	return pkg, nil
}

func emitFile(rootDir, filename string, contents []byte) error {
	outPath := filepath.Join(rootDir, filename)
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(outPath, contents, 0600); err != nil {
		return err
	}
	return nil
}
