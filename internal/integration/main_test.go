package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

var (
	// This deadline is used for tests with pulumi command (only with locking, tho).
	withPulumiDeadline = time.Now().Add(20 * time.Minute)
)

func defaultDeadline() time.Time {
	// This is a default deadline for tests.
	//nolint: unused
	return time.Now().Add(5 * time.Minute)
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithDeadline(context.Background(), withPulumiDeadline)
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
		if err := i.UpWithRetry(); err != nil {
			log.Fatalf("failed to run UP for stack: %v", err)
		}
	}

	exitVal := m.Run()

	log.Println("Tearing down...")

	os.Exit(exitVal)
}
