package config

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	agentRole  = "agent"
	serverRole = "server"
)

type Config struct {
	Organization string
	Nodes        *Nodes
	Defaults     *Defaults
}

func New(ctx *pulumi.Context) *Config {
	var defaults *Defaults
	var nodes *Nodes
	c := config.New(ctx, "")

	c.RequireSecretObject("defaults", &defaults)
	c.RequireSecretObject("nodes", &nodes)

	return &Config{
		Nodes:        nodes,
		Defaults:     defaults,
		Organization: c.Require("organization"),
	}
}

func (c *Config) MergeNodesConfiguration() (*Node, []*Node, error) {
	var leader *Node
	followers := make([]*Node, 0)

	for i, a := range c.Nodes.Agents {
		a.Role = agentRole
		if hetznerFirewallConfigured(a.Server) {
			c.Nodes.Servers[i].Server.Firewall.Hetzner.MarkAsDedicated()
		}
		agent, err := merge(*a, *c.Defaults)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to merge the agent config: %w", err)
		}
		followers = append(followers, &agent)
	}
	for i, s := range c.Nodes.Servers {
		s.Role = serverRole
		if c.Nodes.Servers[i].Leader {
			s, err := merge(*s, *c.Defaults)
			leader = &s
			if hetznerFirewallConfigured(s.Server) {
				c.Nodes.Servers[i].Server.Firewall.Hetzner.MarkAsDedicated()
			}
			if err != nil {
				return nil, nil, fmt.Errorf("failed to merge the leader config: %w", err)
			}
			continue
		}
		if hetznerFirewallConfigured(c.Nodes.Servers[i].Server) {
			c.Nodes.Servers[i].Server.Firewall.Hetzner.MarkAsDedicated()
		}
		s, err := merge(*s, *c.Defaults)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to merge server config: %w", err)
		}
		followers = append(followers, &s)
	}

	return leader, followers, nil
}

func merge(node Node, defaults Defaults) (Node, error) {
	global := defaults.Global
	agents := defaults.Agents
	servers := defaults.Servers

	switch role := node.Role; role {
	case agentRole:
		if err := mergo.Merge(agents, global, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if err := mergo.Merge(&node, agents, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
	case serverRole:
		if err := mergo.Merge(servers, global, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if err := mergo.Merge(&node, servers, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
	}
	return node, nil
}

func hetznerFirewallConfigured(server *Server) bool {
	if server != nil && server.Firewall != nil && server.Firewall.Hetzner != nil {
		return true
	}
	return false
}

type BoolTransformer struct{}

// A Transformer for mergo to avoid overwriting false values from node level.
func (b BoolTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(bool(true)) {
		return func(dst, src reflect.Value) error {
			// Do not overwrite false from node level!
			return nil
		}
	}
	return nil
}
