package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
)

var (
	errNoLeader                                 = errors.New("there is no a leader. Please set it in config")
	errAgentLeader                              = errors.New("agent can't be a leader")
	errManyLeaders                              = errors.New("there is more than one leader")
	errK8SUnknownType                           = fmt.Errorf("unknown k8s endpoint type. Valid types: %v", validConnectionTypes)
	errInternalNetworkDisabled                  = errors.New("internal endpoint type requires hetzner network to be enabled")
	errCCMNetworkingWithInternalNetworkDisabled = errors.New("Hetzner CCM networking is required hetzner network to be enabled")
	errWGNetworkDisabled                        = errors.New("wireguard endpoint type requires wireguard to be enabled")

	validConnectionTypes = []string{
		variables.PublicCommunicationMethod,
		variables.WgCommunicationMethod,
		variables.InternalCommunicationMethod,
	}
)

// Validate validates config globally.
// If checking requires different parts of the configuration it should be done here, in config package.
// If checking requires only one specific part of the configuration in Validate() method of that part.
func (c *Config) Validate(nodes []*Node) error {
	errs := make([]string, 0)
	validators := []func() error{
		c.ValidateCCM,
	}

	for _, validator := range validators {
		if err := validator(); err != nil {
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

		if c.K8S.KubeAPIEndpoint.Type == variables.InternalCommunicationMethod && !c.Network.Hetzner.Enabled {
			return errInternalNetworkDisabled
		}

		if c.K8S.KubeAPIEndpoint.Type == variables.WgCommunicationMethod && !c.Network.Wireguard.Enabled {
			return errWGNetworkDisabled
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

func (c *Config) ValidateCCM() error {
	if !c.Network.Hetzner.Enabled && c.K8S.Addons.CCM.Networking {
		return errCCMNetworkingWithInternalNetworkDisabled
	}

	return nil
}
