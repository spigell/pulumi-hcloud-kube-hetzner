package config

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	k8sconfig "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
)

type WithID interface {
	GetID() string
}

type Defaults struct {
	Global  *Node
	Servers *Node
	Agents  *Node
}

type Nodepools struct {
	Servers []*Nodepool
	Agents  []*Nodepool
}

type Nodepool struct {
	ID     string
	Config *Node
	Nodes  []*Node
}

type Network struct {
	Hetzner *network.Config
}

func (n *Nodepool) GetID() string {
	return n.ID
}

type Node struct {
	ID     string
	Leader bool
	Server *Server
	K3s    *k3s.Config
	K8S    *k8sconfig.NodeConfig
	Role   string
}

func (n *Node) GetID() string {
	return n.ID
}

type Server struct {
	ServerType string `json:"server-type" yaml:"server-type"`
	Hostname   string
	Firewall   *Firewall
	Location   string
	UserName   string
	UserPasswd string
	Image      string
}

type Firewall struct {
	Hetzner *firewall.Config
}

func (d *Defaults) WithInited() *Defaults {
	if d == nil {
		d = &Defaults{}
	}

	if d.Global == nil {
		d.Global = &Node{}
	}

	if d.Agents == nil {
		d.Agents = &Node{}
	}

	if d.Servers == nil {
		d.Servers = &Node{}
	}

	if d.Global.K3s == nil {
		d.Global.K3s = &k3s.Config{}
	}

	if d.Global.K3s.K3S == nil {
		d.Global.K3s.K3S = &k3s.K3sConfig{}
	}

	return d
}

func (n *Network) WithInited() *Network {
	if n.Hetzner == nil {
		n.Hetzner = &network.Config{
			Enabled: false,
		}
	}

	return n
}

func (no *Nodepools) WithInited(ctx *pulumi.Context) *Nodepools {
	no.Agents = initNodepools(ctx, no.Agents)
	no.Servers = initNodepools(ctx, no.Servers)

	return no
}

func initNodepools(ctx *pulumi.Context, pools []*Nodepool) []*Nodepool {
	no := make([]*Nodepool, 0)

	for i, pool := range pools {
		no = append(no, pool)
		if pool.Config == nil {
			no[i].Config = &Node{}
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
			no[i].Config.Server = &Server{}
		}

		for j, node := range pool.Nodes {
			if node.Server == nil {
				no[i].Nodes[j].Server = &Server{}
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
