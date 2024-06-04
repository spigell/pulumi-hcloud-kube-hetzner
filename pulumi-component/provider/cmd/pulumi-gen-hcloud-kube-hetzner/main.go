package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	dotnetgen "github.com/pulumi/pulumi/pkg/v3/codegen/dotnet"
	gogen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	nodejsgen "github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	pygen "github.com/pulumi/pulumi/pkg/v3/codegen/python"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/provider/cmd/pulumi-gen-hcloud-kube-hetzner/resources"
)

const tool = "Pulumi SDK Generator"

// Language is the SDK language.
type Language string

const (
	NodeJS Language = "nodejs"
	DotNet Language = "dotnet"
	Go     Language = "go"
	Python Language = "python"
	Schema Language = "schema"
)

func main() {
	printUsage := func() {
		fmt.Printf("Usage: %s <language> <out-dir> [schema-file] [version]\n", os.Args[0])
	}

	args := os.Args[1:]
	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	language, outdir := Language(args[0]), args[1]

	var schemaFile string
	var version string
	if language != Schema {
		if len(args) < 4 {
			printUsage()
			os.Exit(1)
		}
		schemaFile, version = args[2], args[3]
	}

	switch language {
	case NodeJS:
		genNodeJS(readSchema(schemaFile, version), outdir)
	case DotNet:
		genDotNet(readSchema(schemaFile, version), outdir)
	case Go:
		genGo(readSchema(schemaFile, version), outdir)
	case Python:
		genPython(readSchema(schemaFile, version), outdir)
	case Schema:
		pkgSpec := generateSchema()
		mustWritePulumiSchema(pkgSpec, outdir)
	default:
		panic(fmt.Sprintf("Unrecognized language %q", language))
	}
}

func generateSchema() schema.PackageSpec {
	types, err := resources.GatherClusterTypes("../../../../internal")
	if err != nil {
		panic(err)
	}

	return schema.PackageSpec{
		Name:              "hcloud-kube-hetzner",
		Description:       "Hetzner Cloud Kubernetes",
		License:           "Apache-2.0",
		Keywords:          []string{"pulumi", "hetzner", "k3s", "category/infrastructure", "kind/component", "kubernetes"},
		Publisher:         "spigell",
		Repository:        "https://github.com/spigell/pulumi-hcloud-kube-hetzner",
		PluginDownloadURL: "github://api.github.com/spigell/pulumi-hcloud-kube-hetzner",
		Types:             types,
		Resources: map[string]schema.ResourceSpec{
			"hcloud-kube-hetzner:index:Cluster": {
				IsComponent: true,
				ObjectTypeSpec: schema.ObjectTypeSpec{
					Description: "Component for creating a Hetzner Cloud Kubernetes cluster.",
					Properties: map[string]schema.PropertySpec{
						phkh.KubeconfigKey: {
							TypeSpec:    schema.TypeSpec{Type: "string"},
							Description: "The kubeconfig for the cluster.",
						},
						phkh.PrivatekeyKey: {
							TypeSpec:    schema.TypeSpec{Type: "string"},
							Description: "The private key for nodes.",
						},
						phkh.HetznerServersKey: {
							TypeSpec: schema.TypeSpec{
								Type: "array",
								Items: &schema.TypeSpec{
									Ref: "#types/" + resources.ClusterServersOutputsName,
								},
							},
							Description: "Information about hetnzer servers.",
						},
						"config": {
							TypeSpec: schema.TypeSpec{
								Type: "object",
								Ref:  "#types/" + resources.ClusterTypePrefix + ":" + resources.ClusterConfigType,
							},
						},
					},
				},
				InputProperties: map[string]schema.PropertySpec{
					"config": {
						TypeSpec: schema.TypeSpec{
							OneOf: []schema.TypeSpec{
								{
									Ref: "#types/" + resources.ClusterTypePrefix + ":" + resources.ClusterConfigType,
								},
								{
									Type:                 "object",
									AdditionalProperties: &schema.TypeSpec{Type: "string"},
								},
							},
						},
						Description: "Configuration for the cluster. \n" +
							"Can be Struct or pulumi.Map types. \n" +
							"Despite of the fact that SDK can accept multiple types it is recommended to use strong typep struct if possible. \n" +
							"Caution: Not all configuration options for k3s cluster are available. \n" +
							"Additional information can be found at https://github.com/spigell/pulumi-hcloud-kube-hetzner/blob/main/docs/parameters.md",
					},
				},
			},
		},
		Language: map[string]schema.RawMessage{
			"csharp": rawMessage(map[string]any{
				"packageReferences": map[string]string{
					"Pulumi":            "3.*",
					"Pulumi.Kubernetes": "4.*",
				},
			}),
			"python": rawMessage(map[string]any{
				"requires": map[string]string{
					"pulumi":            ">=3.0.0,<4.0.0",
					"pulumi-kubernetes": ">=4.0.0,<5.0.0",
				},
				"usesIOClasses":                true,
				"liftSingleValueMethodReturns": true,
				"pyproject": map[string]any{
					"enabled": true,
				},
			}),
			"nodejs": rawMessage(map[string]any{
				"packageName": "@spigell/hcloud-kube-hetzner",
				"devDependencies": map[string]any{
					"typescript":  "^4.3.5",
					"@types/node": "^20.0.0",
				},
				"dependencies": map[string]any{
					"@pulumi/pulumi":       "^3.0.0",
					"@pulumi/command":      "0.11.1",
					"@pulumi/kubernetes":   "^4.0.0",
					"@spigell/pulumi-file": "0.0.6",
					"@pulumi/hcloud":       "1.19.1",
				},
			}),
			"go": rawMessage(map[string]interface{}{
				"generateResourceContainerTypes": true,
				"importBasePath":                 "github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner",
			}),
		},
	}
}

func rawMessage(v interface{}) schema.RawMessage {
	bytes, err := json.Marshal(v)
	contract.Assertf(err == nil, fmt.Errorf("error: %w", err).Error())
	return bytes
}

func readSchema(schemaPath string, version string) *schema.Package {
	// Read in, decode, and import the schema.
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		panic(err)
	}

	var pkgSpec schema.PackageSpec
	if err = json.Unmarshal(schemaBytes, &pkgSpec); err != nil {
		panic(err)
	}
	pkgSpec.Version = version

	pkg, err := schema.ImportSpec(pkgSpec, nil)
	if err != nil {
		panic(err)
	}
	return pkg
}

func genDotNet(pkg *schema.Package, outdir string) {
	files, err := dotnetgen.GeneratePackage(tool, pkg, map[string][]byte{}, nil)
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func genGo(pkg *schema.Package, outdir string) {
	files, err := gogen.GeneratePackage(tool, pkg)
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func genPython(pkg *schema.Package, outdir string) {
	files, err := pygen.GeneratePackage(tool, pkg, map[string][]byte{})
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func genNodeJS(pkg *schema.Package, outdir string) {
	files, err := nodejsgen.GeneratePackage(tool, pkg, map[string][]byte{}, nil)
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func mustWriteFiles(rootDir string, files map[string][]byte) {
	for filename, contents := range files {
		mustWriteFile(rootDir, filename, contents)
	}
}

func mustWriteFile(rootDir, filename string, contents []byte) {
	outPath := filepath.Join(rootDir, filename)

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		panic(err)
	}
	err := os.WriteFile(outPath, contents, 0o600)
	if err != nil {
		panic(err)
	}
}

func mustWritePulumiSchema(pkgSpec schema.PackageSpec, outdir string) {
	schemaJSON, err := json.MarshalIndent(pkgSpec, "", "    ")
	if err != nil {
		panic(errors.Wrap(err, "marshaling Pulumi schema"))
	}
	mustWriteFile(outdir, "schema.json", schemaJSON)
}
