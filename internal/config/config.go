package config

import (
	"fmt"
	"reflect"
	"sort"

	"dario.cat/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/k8sconfig"
)

const (
	AgentRole  = "agent"
	ServerRole = "server"
)

type Config struct {
	// Nodepools is a map with agents and servers defined.
	// Required for at least one server node.
	// Default is not specified.
	Nodepools *NodepoolsConfig
	// Defaults is a map with default settings for agents and servers.
	// Global values for all nodes can be set here as well.
	// Default is not specified.
	Defaults *DefaultConfig
	// Network defines network configuration for cluster.
	// Default is not specified.
	Network *NetworkConfig
	// K8S defines a distribution-agnostic cluster configuration.
	// Default is not specified.
	K8S *k8sconfig.Config
}

func ParseClusterConfig(cfg map[string]any) (*Config, error) {
	return parse(cfg)
}

func parse[T Config](cfg map[string]any) (*T, error) {
	var c *T

	// This tag is not used
	tagName := "not-used-tag"

	if _, ok := cfg["nodepools"]; ok {
		// My config if the kebab-case map.
		tagName = "mapstructure"
	}

	// Copying with custom tag for preserving values from GO and TS SDKs in CamelCase.
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: tagName,
		Result:  &c,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	return c, nil
}

// WithInited returns the parsed configuration for the cluster with all the defaults set.
// Nodepools and Nodes are returned sorted.
// This is required for the network module to work correctly when user changes order of nodepools and nodes.
func (c *Config) WithInited() *Config {
	if c.Network == nil {
		c.Network = &NetworkConfig{}
	}

	if c.Defaults == nil {
		c.Defaults = &DefaultConfig{}
	}

	if c.K8S == nil {
		c.K8S = &k8sconfig.Config{}
	}

	if c.Nodepools == nil {
		c.Nodepools = &NodepoolsConfig{}
	}

	c.Network.WithInited()
	c.Defaults.WithInited()
	c.K8S.WithInited()
	c.Nodepools.WithInited()
	c.Nodepools.SpecifyLeader()

	// Sort
	c.Nodepools.Agents = sortByID(c.Nodepools.Agents)
	c.Nodepools.Servers = sortByID(c.Nodepools.Servers)

	for i, pool := range c.Nodepools.Agents {
		c.Nodepools.Agents[i].Nodes = sortByID(pool.Nodes)
	}

	for i, pool := range c.Nodepools.Servers {
		c.Nodepools.Servers[i].Nodes = sortByID(pool.Nodes)
	}

	return c
}

// Nodes returns the nodes for the cluster.
// They are merged with the defaults and nodepool config values.
// They are sorted by majority as well.
func (c *Config) Nodes() ([]*NodeConfig, error) {
	nodes := make([]*NodeConfig, 0)

	for agentpoolIdx, agentpool := range c.Nodepools.Agents {
		if hetznerFirewallConfigured(agentpool.Config.Server) {
			c.Nodepools.Agents[agentpoolIdx].Config.Server.Firewall.Hetzner.MarkWithDedicatedPool()
		}
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

func (no *NodepoolsConfig) SpecifyLeader() {
	if len(no.Servers) == 1 && len(no.Servers[0].Nodes) == 1 {
		no.Servers[0].Nodes[0].Leader = true
	}
}

// sortByMajority sorts nodes by majority.
// The first if leader, then other servers, then workers.
func sortByMajority(n []*NodeConfig) []*NodeConfig {
	nodes := make([]*NodeConfig, 0)

	for _, node := range n {
		if node.Leader {
			nodes = append([]*NodeConfig{node}, nodes...)
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

func merge(node NodeConfig, nodepool *NodeConfig, defaults DefaultConfig) (NodeConfig, error) {
	global := defaults.Global
	agents := defaults.Agents
	servers := defaults.Servers

	if nodepool == nil {
		nodepool = &NodeConfig{}
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

func hetznerFirewallConfigured(server *ServerConfig) bool {
	if server != nil && server.Firewall != nil && server.Firewall.Hetzner != nil && server.Firewall.Hetzner.Enabled != nil && *server.Firewall.Hetzner.Enabled {
		return true
	}
	return false
}

// BoolTransformer is simple struct for mergo.
// ParameterDoc: none.
type BoolTransformer struct{}

// A Transformer for mergo to avoid overwriting false values from node level.
func (b BoolTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(new(bool)) { // Check for *bool type
		return func(dst, src reflect.Value) error {
			// If dst is nil, we should consider the src value
			if dst.IsNil() {
				dst.Set(src)
			}
			// If dst is set (even to false), do nothing
			return nil
		}
	}
	return nil
}
