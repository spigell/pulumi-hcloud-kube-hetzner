package config

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/k8sconfig"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
)

type WithID interface {
	GetID() string
}

type DefaultConfig struct {
	// Global provides configuration settings that are applied to all nodes, unless overridden by specific roles.
	Global *NodeConfig

	// Servers holds configuration settings specific to server nodes, overriding Global settings where specified.
	Servers *NodeConfig

	// Agents holds configuration settings specific to agent nodes, overriding Global settings where specified.
	Agents *NodeConfig
}

type NodepoolsConfig struct {
	// Servers is a list of NodepoolConfig objects, each representing a configuration for a pool of server nodes.
	Servers []*NodepoolConfig

	// Agents is a list of NodepoolConfig objects, each representing a configuration for a pool of agent nodes.
	Agents []*NodepoolConfig
}

type NodepoolConfig struct {
	// PoolID is id of group of servers. It is used through the entire program as key for the group.
	// Required.
	// Default is not specified.
	PoolID string `json:"pool-id" yaml:"pool-id" mapstructure:"pool-id"`
	// Config is the default node configuration for the group.
	Config *NodeConfig
	// Nodes is a list of nodes inside of the group.
	Nodes []*NodeConfig
}

type NetworkConfig struct {
	// Hetzner specifies network configuration for private networking.
	Hetzner *network.Config
}

func (n *NodepoolConfig) GetID() string {
	return n.PoolID
}

type NodeConfig struct {
	// NodeID is the id of a server. It is used throughout the entire program as a key.
	// Required.
	// Default is not specified.
	NodeID string `json:"node-id" yaml:"node-id" mapstructure:"node-id"`
	// Leader specifies the leader of a multi-master cluster.
	// Required if the number of masters is more than 1.
	// Default is not specified.
	Leader bool
	// Server is the configuration of a Hetzner server.
	Server *ServerConfig
	// K3S is the configuration of a k3s cluster.
	K3s *k3s.Config
	// K8S is common configuration for nodes.
	K8S *k8sconfig.NodeConfig
	// Role specifies the role of the server (server or agent).
	// Default is computed.
	Role string `json:"-" yaml:"-" mapstructure:"-"`
}

func (n *NodeConfig) GetID() string {
	return n.NodeID
}

type ServerConfig struct {
	// ServerType specifies the type of server to be provisioned (e.g., "cx11", "cx21").
	// Default is cx21.
	ServerType string `json:"server-type" yaml:"server-type" mapstructure:"server-type"`

	// Hostname is the desired hostname to assign to the server.
	// Default is `phkh-${name-of-stack}-${name-of-cluster}-${id-of-node}`.
	Hostname string

	// Firewall points to an optional configuration for a firewall to be associated with the server.
	Firewall *FirewallConfig

	// Location specifies the physical location or data center where the server will be hosted (e.g., "fsn1").
	// Default is hel1.
	Location string

	// AdditionalSSHKeys contains a list of additional public SSH keys to install in the server's user account.
	AdditionalSSHKeys []string `json:"additional-ssh-keys" yaml:"additional-ssh-keys" mapstructure:"additional-ssh-keys"`

	// UserName is the primary user account name that will be created on the server.
	// Default is rancher.
	UserName string `json:"user-name" yaml:"user-name" mapstructure:"user-name"`

	// UserPasswd is the password for the primary user account on the server.
	UserPasswd string `json:"user-password" yaml:"user-password" mapstructure:"user-password"`

	// Image specifies the operating system image to use for the server (e.g., "ubuntu-20.04" or id of private image).
	// Default is autodiscovered.
	Image string
}

type FirewallConfig struct {
	// Hetzner specify firewall configuration for cloud firewall.
	Hetzner *firewall.Config
}

func (d *DefaultConfig) WithInited() *DefaultConfig {
	if d == nil {
		d = &DefaultConfig{}
	}

	if d.Global == nil {
		d.Global = &NodeConfig{}
	}

	if d.Agents == nil {
		d.Agents = &NodeConfig{}
	}

	if d.Servers == nil {
		d.Servers = &NodeConfig{}
	}

	if d.Global.K3s == nil {
		d.Global.K3s = &k3s.Config{}
	}

	if d.Global.K8S == nil {
		d.Global.K8S = &k8sconfig.NodeConfig{}
	}

	if d.Global.K8S.NodeTaints == nil {
		d.Global.K8S.NodeTaints = &k8sconfig.TaintConfig{}
	}

	if d.Global.K8S.NodeTaints.Enabled == nil {
		d.Global.K8S.NodeTaints.Enabled = new(bool)
	}

	if d.Global.K8S.NodeTaints.DisableDefaultsTaints == nil {
		d.Global.K8S.NodeTaints.DisableDefaultsTaints = new(bool)
	}

	if d.Global.K3s.K3S == nil {
		d.Global.K3s.K3S = &k3s.K3sConfig{}
	}

	if d.Global.Server == nil {
		d.Global.Server = &ServerConfig{}
	}

	if d.Global.Server.Firewall == nil {
		d.Global.Server.Firewall = &FirewallConfig{}
	}

	if d.Global.Server.Firewall.Hetzner == nil {
		d.Global.Server.Firewall.Hetzner = &firewall.Config{}
	}
	if d.Global.Server.Firewall.Hetzner.Enabled == nil {
		d.Global.Server.Firewall.Hetzner.Enabled = new(bool)
	}

	if d.Global.Server.Firewall.Hetzner.SSH == nil {
		d.Global.Server.Firewall.Hetzner.SSH = &firewall.SSHConfig{}
	}

	if d.Global.Server.Firewall.Hetzner.SSH.Allow == nil {
		d.Global.Server.Firewall.Hetzner.SSH.Allow = new(bool)
	}

	if d.Global.Server.Firewall.Hetzner.SSH.DisallowOwnIP == nil {
		d.Global.Server.Firewall.Hetzner.SSH.DisallowOwnIP = new(bool)
	}

	if d.Global.Server.Firewall.Hetzner.AllowICMP == nil {
		d.Global.Server.Firewall.Hetzner.AllowICMP = new(bool)
	}

	return d
}

func (n *NetworkConfig) WithInited() *NetworkConfig {
	if n.Hetzner == nil {
		n.Hetzner = &network.Config{
			Enabled: false,
		}
	}

	return n
}

func (no *NodepoolsConfig) WithInited() *NodepoolsConfig {
	no.Agents = initNodepools(no.Agents)
	no.Servers = initNodepools(no.Servers)

	return no
}

func initNodepools(pools []*NodepoolConfig) []*NodepoolConfig {
	no := make([]*NodepoolConfig, 0)

	for i, pool := range pools {
		no = append(no, pool)
		if pool.Config == nil {
			no[i].Config = &NodeConfig{}
		}

		if pool.Config.K3s == nil {
			no[i].Config.K3s = &k3s.Config{}
		}

		if pool.Config.K3s.K3S == nil {
			no[i].Config.K3s.K3S = &k3s.K3sConfig{}
		}

		if pool.Config.Server == nil {
			no[i].Config.Server = &ServerConfig{}
		}

		for j, node := range pool.Nodes {
			if node.Server == nil {
				no[i].Nodes[j].Server = &ServerConfig{}
			}

			if node.K3s == nil {
				no[i].Nodes[j].K3s = &k3s.Config{}
			}

			if node.K3s.K3S == nil {
				no[i].Nodes[j].K3s.K3S = &k3s.K3sConfig{}
			}
		}
	}

	return no
}
