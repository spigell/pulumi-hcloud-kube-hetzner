package config

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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
	// ID is id of group of servers. It is used through entire program as key for the group.
	// Required.
	// Default is not specified.
	ID string
	// Config is the default node configuration for group
	Config *NodeConfig
	// Nodes is a list of nodes inside of the group.
	Nodes []*NodeConfig
}

type NetworkConfig struct {
	// Hetzner specifies network configuration for private networking.
	Hetzner *network.Config
}

func (n *NodepoolConfig) GetID() string {
	return n.ID
}

type NodeConfig struct {
	// ID is id of server. It is used through entire program as key.
	// Required.
	// Default is not specified.
	ID string
	// Leader specify leader of multi-muster cluster.
	// Required if number of master more than 1.
	// Default is not specified.
	Leader bool
	// Server is configuration of hetzner server.
	Server *ServerConfig
	// K3S is configuration of k3s cluster.
	K3s *k3s.Config
	// K8S is common configuration for nodes.
	K8S  *k8sconfig.NodeConfig
	// Role specifes role of server (server or agent). Do not set manually.
	// Default is computed.
	Role string
}

func (n *NodeConfig) GetID() string {
	return n.ID
}

type ServerConfig struct {
	// ServerType specifies the type of server to be provisioned (e.g., "cx11", "cx21").
	// Default is cx21.
	ServerType string `json:"server-type" yaml:"server-type"`

	// Hostname is the desired hostname to assign to the server.
	// Default is `phkh-${name-of-stack}-${id-of-node}`.
	Hostname string

	// Firewall points to an optional configuration for a firewall to be associated with the server.
	Firewall *FirewallConfig

	// Location specifies the physical location or data center where the server will be hosted (e.g., "fsn1").
	// Default is hel1.
	Location string

	// AdditionalSSHKeys contains a list of additional public SSH keys to install in the server's user account.
	AdditionalSSHKeys []string `json:"additional-ssh-keys" yaml:"additional-ssh-keys"`

	// UserName is the primary user account name that will be created on the server.
	// Default is rancher.
	UserName string `json:"user-name" yaml:"user-name"`

	// UserPasswd is the password for the primary user account on the server.
	UserPasswd string `json:"user-password" yaml:"user-password"`

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

	if d.Global.K3s.K3S == nil {
		d.Global.K3s.K3S = &k3s.K3sConfig{}
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

func (no *NodepoolsConfig) WithInited(ctx *pulumi.Context) *NodepoolsConfig {
	no.Agents = initNodepools(ctx, no.Agents)
	no.Servers = initNodepools(ctx, no.Servers)

	return no
}

func initNodepools(ctx *pulumi.Context, pools []*NodepoolConfig) []*NodepoolConfig {
	no := make([]*NodepoolConfig, 0)

	for i, pool := range pools {
		no = append(no, pool)
		if pool.Config == nil {
			no[i].Config = &NodeConfig{}
		}

		if pool.Config.K8S == nil {
			no[i].Config.K8S = &k8sconfig.NodeConfig{}
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

			if node.Server.Hostname == "" {
				no[i].Nodes[j].Server.Hostname = fmt.Sprintf("%s-%s-%s", ServerNamePrefix, ctx.Stack(), node.ID)
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
