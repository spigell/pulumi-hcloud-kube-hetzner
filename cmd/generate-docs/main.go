package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/docs"
)

const (
	header = `
**Notice**: This document is autogenerated.

*The contents are dynamically generated from the system's current configuration settings and code annotations.
As such, this document may be updated frequently and without prior notice as system configurations or source code are updated.
Users are advised to refer to the latest version of this document for the most accurate and up-to-date information.*

**Do not edit manually!**
---
`
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("path to internal package is required")
	}

	dir := os.Args[1]
	parameters, err := docs.RenderParametersTable(dir)
	// Check if docs is empty
	if err != nil {
		fmt.Printf(`
			No documentation was generated.
			Please check the presence of Go files and structs.
			Error: %s
		`, err.Error())
	}

	fmt.Println(strings.ReplaceAll(header, "    ", "") + parameters)
}
