package integration

import (
	"context"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, err := New(ctx)
	if err != nil {
		log.Fatalf("failed to create integration: %v", err) //nolint: gocritic
	}
	if err := i.Validate(); err != nil {
		log.Fatalf("failed to validate: %v", err)
	}

	// Skip UP if INTEGRATION_NO_UP_STARTUP is set to true.
	// Up can be time consuming and we don't need it if we run tests locally sometimes.
	if os.Getenv("INTEGRATION_NO_UP_STARTUP") != "true" {
		_, err = i.Stack.Up(ctx)
		if err != nil {
			log.Fatalf("failed to run UP for stack: %v", err)
		}
	}

	exitVal := m.Run()

	log.Println("Tearing down...")

	os.Exit(exitVal)
}
