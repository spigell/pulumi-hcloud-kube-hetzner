// This package contains all the integration suites for the pulumi program.
// The integration suites are used to test the pulumi program against a real deployed pulumi stack.
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

const (
	envConfigSource = "PULUMI_CONFIG_SOURCE"
	envConfigPath   = "PULUMI_STACK_CONFIG"

	exampleK3SPrivateNonHASimple = "k3s-private-non-ha-simple"
	exampleK3SWGNonHaFwRules     = "k3s-wireguard-non-ha-firewall-rules"
	exampleK3SWGHANoTaints       = "k3s-wireguard-ha-no-taints"

	testWGConnectivity  = "wireguard-connectivity"
	testKubeVersion     = "kube-version"
	testSSHConnectivity = "ssh-connectivity"
)

var (
	defaultDeadline = time.Now().Add(5 * time.Minute)

	TestsByExampleName = map[string][]string{
		exampleK3SPrivateNonHASimple: {
			testSSHConnectivity,
			testKubeVersion,
		},
		exampleK3SWGNonHaFwRules: {
			testSSHConnectivity,
			testWGConnectivity,
			testKubeVersion,
		},
		exampleK3SWGHANoTaints: {
			testSSHConnectivity,
			testWGConnectivity,
			testKubeVersion,
		},
	}
)

type Integration struct {
	ctx     context.Context
	Example *Example
	Stack   auto.Stack
}

func New(ctx context.Context) (*Integration, error) {
	workDir := filepath.Dir(os.Getenv(envConfigPath))
	e, err := DiscoverExample(os.Getenv(envConfigSource))
	if err != nil {
		return nil, fmt.Errorf("failed to discover example: %w", err)
	}
	stackName := strings.Split(filepath.Base(os.Getenv(envConfigPath)), ".")[1]

	stack, err := auto.SelectStackLocalSource(ctx, stackName, workDir)
	if err != nil {
		return nil, err
	}
	return &Integration{
		ctx:     ctx,
		Stack:   stack,
		Example: e,
	}, nil
}

func (i *Integration) Validate() error {
	out, err := i.Stack.Outputs(i.ctx)
	if err != nil {
		return fmt.Errorf("failed to get stack outputs: %w", err)
	}

	_, ok := TestsByExampleName[i.Example.Name]
	if !ok {
		return fmt.Errorf("no tests found for example %s", i.Example.Name)
	}

	if len(out) == 0 {
		return fmt.Errorf("stack outputs are empty. Stack is not deployed")
	}

	if os.Getenv(envConfigPath) == "" {
		return fmt.Errorf("env variable %s is required", envConfigPath)
	}

	if os.Getenv(envConfigSource) == "" {
		return fmt.Errorf("env variable %s is required", envConfigSource)
	}

	return nil
}

func (i *Integration) UpWithRetry() error {
	return retry.Do(
		func() error {
			_, err := i.Stack.Up(i.ctx)

			if err != nil && ! auto.IsConcurrentUpdateError(err) {
				return retry.Unrecoverable(err)
			}

			return nil
		},
		retry.Delay(15 * time.Second),
		retry.Attempts(10),
	)
}
