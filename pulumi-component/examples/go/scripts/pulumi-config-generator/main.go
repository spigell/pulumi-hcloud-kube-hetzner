package main

import (
	"fmt"
	"log"
	"os"
)



func main() {
	// Read first arg
	if len(os.Args) == 1 {
		log.Fatal("File must be provided. Exiting")
	}
	source := os.Args[1]

	newImage := os.Getenv("HCLOUD_IMAGE")
	if newImage == "" {
		log.Print("HCLOUD_IMAGE env variable is not provided.")
	}

	fmt.Println(NewConfig(source).WithReplaceProjectName().Content)
}
