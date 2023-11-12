package integration

import (
	"log"
	"os"
	"testing"
)


func TestMain(m *testing.M) {
     if err := validate(); err != nil {
     	log.Fatalf("failed to validate: %v", err)
     }
    exitVal := m.Run()

    os.Exit(exitVal)
}