package config

import (
	"fmt"
	"pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"reflect"
	"sort"

	"dario.cat/mergo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	agentRole  = "agent"
	serverRole = "server"
)

type Config struct {
	Nodepools *Nodepools
	Defaults  *Defaults
	Network   *network.Config
}

// New returns the configuration for the cluster.
// Nodepools and Nodes returned sorted.
// This is required for the network module to work correctly when user changes order of nodepools and nodes.
func New(ctx *pulumi.Context) *Config {
	var defaults *Defaults
	var nodepools *Nodepools
	var network *network.Config
	c := config.New(ctx, "")

	c.RequireSecretObject("defaults", &defaults)
	c.RequireSecretObject("nodepools", &nodepools)
	c.RequireSecretObject("network", &network)

	nodepools.Agents = sortByID(nodepools.Agents)
	nodepools.Servers = sortByID(nodepools.Servers)

	for i, pool := range nodepools.Agents {
		nodepools.Agents[i].Nodes = sortByID(pool.Nodes)
	}

	for i, pool := range nodepools.Servers {
		nodepools.Servers[i].Nodes = sortByID(pool.Nodes)
	}

	return &Config{
		Nodepools: nodepools,
		Network:   network,
		Defaults:  defaults,
	}
}

func sortByID[W WithID](unsorted []W) []W {
	sorted := make([]W, 0, len(unsorted))
	keys := make([]string, 0, len(unsorted))

	for _, k := range unsorted {
		keys = append(keys, k.GetID())
	}

	sort.Strings(keys)

	for i, k := range keys {
		sorted = append(sorted, unsorted[i])
		if sorted[i].GetID() != k {
			for _, v := range unsorted {
				if v.GetID() == k {
					sorted[i] = v
				}
			}
		}
	}

	return sorted
}

func (c *Config) MergeNodesConfiguration() (*Node, []*Node, error) {
	var leader *Node
	followers := make([]*Node, 0)

	for agentpoolIdx, agentpool := range c.Nodepools.Agents {
		for i, a := range agentpool.Nodes {
			a.Role = agentRole
			if hetznerFirewallConfigured(a.Server) {
				c.Nodepools.Agents[agentpoolIdx].Nodes[i].Server.Firewall.Hetzner.MarkAsDedicated()
			}
			agent, err := merge(*a, agentpool.Config, *c.Defaults)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to merge the agent config: %w", err)
			}
			followers = append(followers, &agent)
		}
	}
	for serverpoolIdx, serverpool := range c.Nodepools.Servers {
		for i, s := range serverpool.Nodes {
			s.Role = serverRole
			if c.Nodepools.Servers[serverpoolIdx].Nodes[i].Leader {
				if hetznerFirewallConfigured(s.Server) {
					c.Nodepools.Servers[serverpoolIdx].Nodes[i].Server.Firewall.Hetzner.MarkAsDedicated()
				}
				s, err := merge(*s, serverpool.Config, *c.Defaults)
				leader = &s
				if err != nil {
					return nil, nil, fmt.Errorf("failed to merge the leader config: %w", err)
				}
				continue
			}
			if hetznerFirewallConfigured(s.Server) {
				c.Nodepools.Servers[serverpoolIdx].Nodes[i].Server.Firewall.Hetzner.MarkAsDedicated()
			}
			s, err := merge(*s, serverpool.Config, *c.Defaults)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to merge server config: %w", err)
			}
			followers = append(followers, &s)
		}
	}

	return leader, followers, nil
}

func merge(node Node, nodepool *Node, defaults Defaults) (Node, error) {
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
		if nodepool != nil {
			if err := mergo.Merge(&node, nodepool, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
				return node, err
			}
		}
	case serverRole:
		if err := mergo.Merge(servers, global, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if err := mergo.Merge(&node, servers, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if nodepool != nil {
			if err := mergo.Merge(&node, nodepool, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
				return node, err
			}
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
