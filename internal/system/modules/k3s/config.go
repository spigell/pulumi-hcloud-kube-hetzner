package k3s

import (
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

const (
	cfgPath = "/etc/rancher/k3s/config.yaml"
)

type CompletedConfig struct {
	Config *K3sConfig
}

type Config struct {
	// version is used to determine if k3s should be upgraded if auto-upgrade is disabled.
	// If the version is changed, k3s will be upgraded.
	Version string
	// [Experimental] clean-data-on-upgrade is used to delete all data while upgrade.
	// This is based on the script https://docs.k3s.io/upgrades/killall
	CleanDataOnUpgrade bool `json:"clean-data-on-upgrade" yaml:"clean-data-on-upgrade,omitempty"`
	// Do not add default taints to the server node.
	DisableDefaultsTaints bool `json:"disable-default-taints" yaml:"disable-default-taints,omitempty"`
	// The real config of k3s service.
	K3S *K3sConfig `json:"config" yaml:"config"`
}

type K3sConfig struct { //nolint: revive // This name is better
	// Token is used to join the cluster.
	// It is generated by the program.
	Token string `json:"-"`
	// Server is the address of the main server node (leader).
	// It is generated by the program.
	Server string `json:"-" yaml:",omitempty"`
	// FlannelIface is used to set iface for flannel.
	// It is generated by the program.
	FlannelIface         string   `json:"-" yaml:"flannel-iface,omitempty"`
	WriteKubeconfigMode  string   `json:"-" yaml:"write-kubeconfig-mode,omitempty"`
	AdvertiseAddr        string   `json:"-" yaml:"advertise-address,omitempty"`
	NodeIP               string   `json:"-" yaml:"node-ip,omitempty"`
	BindAddress          string   `json:"-" yaml:"bind-address,omitempty"`
	ClusterInit          bool     `json:"-" yaml:"cluster-init,omitempty"`
	ExternalNodeIP       string   `json:"-" yaml:"node-external-ip,omitempty"`
	TLSSanSecurity       bool     `json:"-" yaml:"tls-san-security,omitempty"`
	TLSSan               string   `json:"-" yaml:"tls-san,omitempty"`
	NodeName             string   `json:"-" yaml:"node-name,omitempty"`
	ClusterCidr          string   `json:"cluster-cidr" yaml:"cluster-cidr,omitempty"`
	ServiceCidr          string   `json:"service-cidr" yaml:"service-cidr,omitempty"`
	ClusterDomain        string   `json:"cluster-domain" yaml:"cluster-domain,omitempty"`
	ClusterDNS           string   `json:"cluster-dns" yaml:"cluster-dns,omitempty"`
	NodeLabels           []string `json:"node-label" yaml:"-"`
	FlannelBackend       string   `json:"flannel-backend" yaml:"flannel-backend,omitempty"`
	DisableNetworkPolicy bool     `json:"disable-network-policy" yaml:"disable-network-policy,omitempty"`
	// NodeTaints are used to taint the node with key=value:effect.
	// By default, server node is tainted with a couple of taints if number of agents nodes more than 0.
	NodeTaints                     []string `json:"node-taint" yaml:"-"`
	KubeletArgs                    []string `json:"kubelet-arg" yaml:"kubelet-arg,omitempty"`
	KubeControllerManagerArgs      []string `json:"kube-controller-manager-arg" yaml:"kube-controller-manager-arg,omitempty"`
	KubeCloudControllerManagerArgs []string `json:"kube-cloud-controller-manager-arg" yaml:"kube-cloud-controller-manager-arg,omitempty"`
	KubeAPIServerArgs              []string `json:"kube-apiserver-arg" yaml:"kube-apiserver-arg,omitempty"`
	DisableCloudController         bool     `json:"disable-cloud-controller" yaml:"disable-cloud-controller,omitempty"`
	// Disable is a list to disable some services.
	Disable []string `yaml:"disable,omitempty"`
}

func (c *CompletedConfig) render() ([]byte, error) {
	return yaml.Marshal(&c.Config)
}

func (k *K3sConfig) WithoutDuplicates() *K3sConfig {
	k.Disable = slices.Compact(k.Disable)
	k.KubeletArgs = slices.Compact(k.KubeletArgs)
	k.KubeControllerManagerArgs = slices.Compact(k.KubeControllerManagerArgs)
	k.KubeAPIServerArgs = slices.Compact(k.KubeAPIServerArgs)

	return k
}
