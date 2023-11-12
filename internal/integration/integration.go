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

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)


const (
	envConfigSource = "PULUMI_CONFIG_SOURCE"
	envConfigPath = "PULUMI_STACK_CONFIG"

	exampleHaServerWithWorkload = "ha-server-with-workload"

	testSSHConnectivity = "ssh-connectivity"
	testKubeVersion = "kube-version"
)
var (
	TestsByExampleName = map[string][]string{
		exampleHaServerWithWorkload: {
			testSSHConnectivity,
			testKubeVersion,
		},
	}

	defaultDeadline = time.Now().Add(10 * time.Minute)
)

type Integration struct {
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
		Stack: stack,
		Example: e,

	}, nil
}

func validate() error {
	ctx := context.Background()
	i, err := New(ctx)

	if err != nil {
		return err
	}

	out, err := i.Stack.Outputs(ctx)
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
	