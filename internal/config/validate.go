package config

import (
	"errors"
)

var (
	errNoLeader    = errors.New("there is no a leader. Please set it in config")
	errAgentLeader = errors.New("agent can't be a leader")
	errManyLeaders = errors.New("there is more than one leader")
)

func (c *Config) Validate(nodes []*Node) error {
	leaderFounded := false
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
