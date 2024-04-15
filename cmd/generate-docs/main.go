package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/docs"
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

	fmt.Println(parameters)
}
