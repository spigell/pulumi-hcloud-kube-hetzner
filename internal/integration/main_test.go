package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, err := New(ctx)
	if err != nil {
		log.Fatalf("failed to create integration: %v", err)
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

	out, err := i.Stack.Outputs(ctx)
	if err != nil {
		log.Fatalf("failed to get stack outputs: %v", err)
	}

	wg, ok := out[phkh.WGMasterConKey].Value.(string)
	if !ok {
		log.Fatalf("failed to get wg master connection string from stack outputs")
	}

	up, err := wireguard.Up(wg)
	if err != nil {
		up.Close()
		log.Fatalf("failed to run UP for wireguard: %v", err)
	}

	exitVal := m.Run()

	log.Println("Tearing down...")

	if err := up.Close(); err != nil {
		log.Fatalf("failed to run CLOSE for wireguard device: %v. Please remove it manually", err)
	}

	os.Exit(exitVal)
}
