//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

func main() {
	version, found := os.LookupEnv("VERSION")
	if !found {
		log.Fatal("version not found")
	}

	schemaPath, found := os.LookupEnv("SCHEMA")
	if !found {
		log.Fatal("schema not found")
	}

	schemaContents, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		log.Fatal(err)
	}

	var packageSpec schema.PackageSpec
	err = json.Unmarshal(schemaContents, &packageSpec)
	if err != nil {
		log.Fatalf("cannot deserialize schema: %v", err)
	}

	packageSpec.Version = version
	versionedContents, err := json.Marshal(packageSpec)
	if err != nil {
		log.Fatalf("cannot reserialize schema: %v", err)
	}

	err = ioutil.WriteFile("./schema.go", []byte(fmt.Sprintf(`package main

var pulumiSchema = %#v
`, versionedContents)), 0600)
	if err != nil {
		log.Fatal(err)
	}
}
