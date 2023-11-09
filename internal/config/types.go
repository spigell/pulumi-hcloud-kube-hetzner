package config

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
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

type K8S struct {
	Endpoint *K8SEndpoint
}

type K8SEndpoint struct {
	Type     string
	Firewall *BasicFirewall
}

type BasicFirewall struct {
	HetznerPublic *HetnzerBasidFirewall `json:"hetzner-public"`
}

type HetnzerBasidFirewall struct {
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips"`
}

type Network struct {
	Hetzner   *network.Config
	Wireguard *wireguard.Config
}

func (n *Nodepool) GetID() string {
	return n.ID
}

type Node struct {
	ID     string
	Leader bool
	Server *Server
	K3s    *k3s.Config
	Role   string
}

func (n *Node) GetID() string {
	return n.ID
}

type Server struct {
	ServerType string `json:"server-type"`
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

func (k *K8S) WithInited() *K8S {
	if k.Endpoint == nil {
		k.Endpoint = &K8SEndpoint{}
	}

	if k.Endpoint.Type == "" {
		k.Endpoint.Type = variables.DefaultCommunicationMethod
	}

	if k.Endpoint.Firewall == nil {
		k.Endpoint.Firewall = &BasicFirewall{
			HetznerPublic: &HetnzerBasidFirewall{
				AllowedIps: firewall.ICMPRule.SourceIps,
			},
		}
	}

	return k

}
