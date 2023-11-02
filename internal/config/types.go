package config

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
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

func (n *Nodepool) GetID() string {
	return n.ID
}

type Node struct {
	ID        string
	Leader    bool
	Wireguard *Wireguard
	Server    *Server
	K3s       *K3s
	Role      string
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
	Hetzner   *firewall.Config
	Firewalld *FWFirewalld
}

type FWFirewalld struct {
	Enabled      bool
	InternalZone *InternalZone
	PublicZone   *PublicZone
}

type InternalZone struct {
	RestrictToSources []*RestrictToSource
}

type RestrictToSource struct {
	CIDR string
	Name string
	Main bool
}

type PublicZone struct {
	RemoveSSHService bool
}

type K3s struct {
	Version            string
	CleanDataOnUpgrade bool
	Config             K3sConfig
}

type K3sConfig struct {
	Token                     string
	Server                    string   `yaml:",omitempty"`
	FlannelIface              string   `json:"-" yaml:"flannel-iface,omitempty"`
	ClusterCidr               string   `json:"cluster-cidr" yaml:"cluster-cidr,omitempty"`
	ServiceCidr               string   `json:"service-cidr" yaml:"service-cidr,omitempty"`
	ClusterDomain             string   `json:"cluster-domain" yaml:"cluster-domain,omitempty"`
	ClusterDNS                string   `json:"cluster-dns" yaml:"cluster-dns,omitempty"`
	WriteKubeconfigMode       string   `json:"-" yaml:"write-kubeconfig-mode,omitempty"`
	NodeIP                    string   `json:"-" yaml:"node-ip,omitempty"`
	BindAddress               string   `json:"-" yaml:"bind-address,omitempty"`
	ClusterInit               bool     `json:"-" yaml:"cluster-init,omitempty"`
	NodeLabels                []string `json:"node-label" yaml:"node-label,omitempty"`
	FlannelBackend            string   `json:"flannel-backend" yaml:"flannel-backend,omitempty"`
	DisableNetworkPolicy      bool     `json:"disable-network-policy" yaml:"disable-network-policy,omitempty"`
	NodeTaints                []string `json:"node-taint" yaml:"node-taint,omitempty"`
	KubeleteArgs              []string `json:"kubelet-arg" yaml:"kubelet-arg,omitempty"`
	KubeControllerManagerArgs []string `json:"kube-controller-manager-arg" yaml:"kube-controller-manager-arg,omitempty"`
	KubeAPIServerArgs         []string `json:"kube-apiserver-arg" yaml:"kube-apiserver-arg,omitempty"`
	DisableCloudController    bool     `json:"disable-cloud-controller" yaml:"disable-cloud-controller,omitempty"`
	Disable                   []string
}

type Wireguard struct {
	Enabled         bool
	IP              string
	Firewall        *WGFirewall
	CIDR            string           `json:"cidr"`
	AdditionalPeers []AdditionalPeer `json:"additional-peers" yaml:"additional-peers"`
}

type WGFirewall struct {
	Firewalld *ServiceFirewall
	Hetzner   *ServiceFirewall
}

type ServiceFirewall struct {
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips"`
}

type AdditionalPeer struct {
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips"`
	Endpoint   string
	PublicKey  string
}
