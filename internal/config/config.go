package config

import (
	"fmt"
	"reflect"
	"sort"

	"dario.cat/mergo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	hnetwork "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
)

const (
	AgentRole  = "agent"
	ServerRole = "server"
)

type Config struct {
	Nodepools *Nodepools
	Defaults  *Defaults
	Network   *Network
}

// New returns the configuration for the cluster.
// Nodepools and Nodes returned sorted.
// This is required for the network module to work correctly when user changes order of nodepools and nodes.
func New(ctx *pulumi.Context) *Config {
	var defaults *Defaults
	var nodepools *Nodepools
	var network *Network
	c := config.New(ctx, "")

	c.RequireSecretObject("defaults", &defaults)
	c.RequireSecretObject("nodepools", &nodepools)
	c.RequireSecretObject("network", &network)

	if defaults == nil {
		defaults = &Defaults{}
	}

	if defaults.Global == nil {
		defaults.Global = &Node{}
	}

	if defaults.Agents == nil {
		defaults.Agents = &Node{}
	}

	if defaults.Servers == nil {
		defaults.Servers = &Node{}
	}

	for i, pool := range nodepools.Agents {
		if pool.Config == nil {
			nodepools.Agents[i].Config = &Node{}
		}

		if pool.Config.K3s == nil {
			nodepools.Agents[i].Config.K3s = &k3s.Config{}
		}

		if pool.Config.K3s.K3S == nil {
			nodepools.Agents[i].Config.K3s.K3S = &k3s.K3sConfig{}
		}

		for j, node := range pool.Nodes {
			if node.Server == nil {
				nodepools.Agents[i].Nodes[j].Server = &Server{}
			}

			if node.K3s == nil {
				nodepools.Agents[i].Nodes[j].K3s = &k3s.Config{}
			}

			if node.K3s.K3S == nil {
				nodepools.Agents[i].Nodes[j].K3s.K3S = &k3s.K3sConfig{}
			}
		}
	}

	for i, pool := range nodepools.Servers {
		if pool.Config == nil {
			nodepools.Servers[i].Config = &Node{}
		}

		if pool.Config.K3s == nil {
			nodepools.Servers[i].Config.K3s = &k3s.Config{}
		}

		if pool.Config.K3s.K3S == nil {
			nodepools.Servers[i].Config.K3s.K3S = &k3s.K3sConfig{}
		}

		for j, node := range pool.Nodes {
			if node.Server == nil {
				nodepools.Servers[i].Nodes[j].Server = &Server{}
			}

			if node.K3s == nil {
				nodepools.Servers[i].Nodes[j].K3s = &k3s.Config{}
			}

			if node.K3s.K3S == nil {
				nodepools.Servers[i].Nodes[j].K3s.K3S = &k3s.K3sConfig{}
			}
		}
	}

	// Sort
	nodepools.Agents = sortByID(nodepools.Agents)
	nodepools.Servers = sortByID(nodepools.Servers)

	for i, pool := range nodepools.Agents {
		nodepools.Agents[i].Nodes = sortByID(pool.Nodes)
	}

	for i, pool := range nodepools.Servers {
		nodepools.Servers[i].Nodes = sortByID(pool.Nodes)
	}

	if network == nil {
		network = &Network{}
	}

	if network.Hetzner == nil {
		network.Hetzner = &hnetwork.Config{
			Enabled: false,
		}
	}

	if network.Wireguard == nil {
		network.Wireguard = &wireguard.Config{
			Enabled: false,
		}
	}

	return &Config{
		Nodepools: nodepools,
		Network:   network,
		Defaults:  defaults,
	}
}

// Nodes returns the nodes for the cluster.
// They are sorted by majority.
func (c *Config) Nodes() ([]*Node, error) {
	nodes := make([]*Node, 0)

	for agentpoolIdx, agentpool := range c.Nodepools.Agents {
		for i, a := range agentpool.Nodes {
			a.Role = AgentRole
			if hetznerFirewallConfigured(a.Server) {
				c.Nodepools.Agents[agentpoolIdx].Nodes[i].Server.Firewall.Hetzner.MarkAsDedicated()
			}
			agent, err := merge(*a, agentpool.Config, *c.Defaults)
			if err != nil {
				return nil, fmt.Errorf("failed to merge the agent config: %w", err)
			}

			nodes = append(nodes, &agent)
		}
	}
	for serverpoolIdx, serverpool := range c.Nodepools.Servers {
		for i, s := range serverpool.Nodes {
			s.Role = ServerRole
			if hetznerFirewallConfigured(s.Server) {
				c.Nodepools.Servers[serverpoolIdx].Nodes[i].Server.Firewall.Hetzner.MarkAsDedicated()
			}
			s, err := merge(*s, serverpool.Config, *c.Defaults)
			if err != nil {
				return nil, fmt.Errorf("failed to merge server config: %w", err)
			}
			nodes = append(nodes, &s)
		}
	}

	return sortByMajority(nodes), nil
}

// sortByMajority sorts nodes by majority.
// The first if leader, then other servers, then workers.
func sortByMajority(n []*Node) []*Node {
	nodes := make([]*Node, 0)

	for _, node := range n {
		if node.Leader {
			nodes = append([]*Node{node}, nodes...)
			continue
		}
		if node.Role == ServerRole {
			nodes = append(nodes, node)
			continue
		}
	}

	for _, node := range n {
		if node.Role == AgentRole {
			nodes = append(nodes, node)
		}
	}

	return nodes
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

func merge(node Node, nodepool *Node, defaults Defaults) (Node, error) {
	global := defaults.Global
	agents := defaults.Agents
	servers := defaults.Servers

	if nodepool == nil {
		nodepool = &Node{}
	}

	switch role := node.Role; role {
	case AgentRole:
		if err := mergo.Merge(agents, global, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if err := mergo.Merge(nodepool, agents, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if err := mergo.Merge(&node, nodepool, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
	case ServerRole:
		if err := mergo.Merge(servers, global, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if err := mergo.Merge(nodepool, servers, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
			return node, err
		}
		if err := mergo.Merge(&node, nodepool, mergo.WithAppendSlice, mergo.WithTransformers(BoolTransformer{})); err != nil {
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
