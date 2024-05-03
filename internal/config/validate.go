package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/k3supgrader"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
)

var (
	errNoLeader                      = errors.New("there is no a leader. Please set it in config")
	errAgentLeader                   = errors.New("agent can't be a leader")
	errManyLeaders                   = errors.New("there is more than one leader")
	errK8SUnknownType                = fmt.Errorf("unknown k8s endpoint type. Valid types: %v", validConnectionTypes)
	errInternalNetworkDisabled       = errors.New("internal endpoint type requires hetzner network to be enabled")
	errConflictBetweenUpgradeMethods = errors.New("node doesn't have `k3s-upgrade=false` label but k3s-upgrade-controller is enabled and version is set")
	errVersionMustBeSetManually      = errors.New("k3s-upgrade-controller is disabled and version is not set. It must be set manually")

	validConnectionTypes = []string{
		variables.PublicCommunicationMethod.String(),
		variables.InternalCommunicationMethod.String(),
	}
)

// Validate validates config globally.
// If checking requires different parts of the configuration it should be done here, in config package.
// If checking requires only one specific part of the configuration in Validate() method of that part.
func (c *Config) Validate(nodes []*NodeConfig) error {
	errs := make([]string, 0)
	validators := make([]func([]*NodeConfig) error, 0)

	if ccm := c.K8S.Addons.CCM; ccm != nil {
		validators = append(validators, c.ValidateCCM)
	}

	if k3sUpgrader := c.K8S.Addons.K3SSystemUpgrader; k3sUpgrader != nil {
		validators = append(validators, c.ValidateK3SUpgradeController)
	}

	for _, validator := range validators {
		if err := validator(nodes); err != nil {
			errs = append(errs, err.Error())
		}
	}

	leaderFounded := false
	// k8s endpoint types are the same as communication methods.
	// Let's reuse it
	if c.K8S.KubeAPIEndpoint.Type != "" {
		if !slices.Contains(validConnectionTypes, c.K8S.KubeAPIEndpoint.Type) {
			return errK8SUnknownType
		}

		if c.K8S.KubeAPIEndpoint.Type == variables.InternalCommunicationMethod.String() && !c.Network.Hetzner.Enabled {
			return errInternalNetworkDisabled
		}
	}
	for _, node := range nodes {
		if node.Leader {
			if node.Role == AgentRole {
				return errAgentLeader
			}

			if !leaderFounded {
				leaderFounded = true
			} else {
				return errManyLeaders
			}
		}
	}
	if !leaderFounded {
		return errNoLeader
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: errors: %s", strings.Join(errs, "|"))
	}

	return nil
}

func (c *Config) ValidateCCM(_ []*NodeConfig) error {
	return nil
}

func (c *Config) ValidateK3SUpgradeController(merged []*NodeConfig) error {
	for _, node := range merged {
		disableLabelFound := findLabel(node, fmt.Sprintf("%s=false", k3supgrader.ControlLabelKey))
		if c.K8S.Addons.K3SSystemUpgrader.Enabled && node.K3s.Version != "" && !disableLabelFound {
			return errConflictBetweenUpgradeMethods
		}

		if !c.K8S.Addons.K3SSystemUpgrader.Enabled && node.K3s.Version == "" && disableLabelFound {
			return errVersionMustBeSetManually
		}
	}

	return nil
}

func findLabel(node *NodeConfig, target string) bool {
	for _, label := range node.K8S.NodeLabels {
		if label == target {
			return true
		}
	}

	return false
}
