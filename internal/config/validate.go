package config

import (
	"errors"
	"fmt"
	"slices"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
)

var (
	errNoLeader       = errors.New("there is no a leader. Please set it in config")
	errAgentLeader    = errors.New("agent can't be a leader")
	errManyLeaders    = errors.New("there is more than one leader")
	errK8SUnknownType = errors.New(fmt.Sprintf("unknown k8s endpoint type. Valid types: %v", validConnectionTypes))

	validConnectionTypes = []string{
		variables.DefaultCommunicationMethod,
		variables.WgCommunicationMethod,
		variables.InternalCommunicationMethod,
	}
)

func (c *Config) Validate(nodes []*Node) error {
	leaderFounded := false
	// k8s endpoint types are the same as communication methods.
	// Let's reuse it
	if c.K8S.Endpoint.Type != "" {
		if !slices.Contains(validConnectionTypes, c.K8S.Endpoint.Type) {
			return errK8SUnknownType
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
	return nil
}
