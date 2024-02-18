// This package contains all the integration suites for the pulumi program.
// The integration suites are used to test the pulumi program against a real deployed pulumi stack.
package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

const (
	// changeMe is a placeholder for values that should be changed by the tests in various suites.
	changeMe        = "change-me"
	envConfigSource = "PULUMI_CONFIG_SOURCE"
	envConfigPath   = "PULUMI_STACK_CONFIG"

	exampleK3SPrivateNonHASimple   = "k3s-private-non-ha-simple"
	exampleK3SPrivateNonHaFwRules  = "k3s-private-non-ha-firewall-rules"
	exampleK3SPrivateHANoTaints    = "k3s-private-ha-no-taints"
	exampleK3SPublicNonHADefaults  = "k3s-public-non-ha-with-defaults"
	exampleK3SPublicHAKubeAddons   = "k3s-public-ha-kube-addons"
	exampleK3SPrivateNonHAUpgrader = "k3s-private-non-ha-upgrader"

	testKubeVersion                       = "kube-version"
	testSSHConnectivity                   = "ssh-connectivity"
	testHetznerNodeManagement             = "hetzner-node-management"
	testKubeHetznerCCM                    = "kube-hetzner-ccm"
	testNodeChangeLabelsTaints            = "node-change-labels-taints"
	testKubeCheckTaintsManager            = "kube-node-check-taints-manager"
	testKubeK3SUpgradeControllerPlan      = "kube-k3s-upgrade-controller-plan"
	testKubeK3SUpgradeControllerConfigEnv = "kube-k3s-upgrade-controller-config-env"
)

// TestsByExampleName is a map of tests and their test cases.
// Please use this map to add new tests for examples.
var TestsByExampleName = map[string][]string{
	exampleK3SPrivateNonHASimple: {
		testSSHConnectivity,
		testKubeHetznerCCM,
		testKubeK3SUpgradeControllerPlan,
		testHetznerNodeManagement,
		testKubeCheckTaintsManager,
	},
	exampleK3SPrivateNonHaFwRules: {
		testSSHConnectivity,
	},
	exampleK3SPrivateHANoTaints: {
		testSSHConnectivity,
		testKubeVersion,
		testNodeChangeLabelsTaints,
		testKubeK3SUpgradeControllerPlan,
	},
	exampleK3SPublicNonHADefaults: {
		testSSHConnectivity,
		testNodeChangeLabelsTaints,
	},
	exampleK3SPublicHAKubeAddons: {
		testSSHConnectivity,
		testKubeVersion,
		testKubeHetznerCCM,
		testNodeChangeLabelsTaints,
		testKubeK3SUpgradeControllerPlan,
		testHetznerNodeManagement,
	},
	exampleK3SPrivateNonHAUpgrader: {
		testKubeK3SUpgradeControllerPlan,
		testKubeK3SUpgradeControllerConfigEnv,
	},
}

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

func (i *Integration) Outputs() (map[string]interface{}, error) {
	out, err := i.Stack.Outputs(i.ctx)
	if err != nil {
		return nil, err
	}

	m, ok := out[phkh.PhkhKey]
	if !ok {
		return nil, errors.New("output map does not contain `phkh` key")
	}

	return m.Value.(map[string]interface{}), nil
}

func (i *Integration) Validate() error {
	_, err := i.Outputs()
	if err != nil {
		return fmt.Errorf("failed to get phkh outputs: %w", err)
	}
	_, ok := TestsByExampleName[i.Example.Name]
	if !ok {
		return fmt.Errorf("no tests found for example %s", i.Example.Name)
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

			if err != nil && !auto.IsConcurrentUpdateError(err) {
				return retry.Unrecoverable(err)
			}

			return nil
		},
		retry.Delay(15*time.Second),
		retry.Attempts(10),
	)
}
