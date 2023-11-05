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
	CleanDataOnUpgrade bool `json:"clean-data-on-upgrade"`
	// The real config of k3s service.
	K3S *K3sConfig `json:"config"`
}

type K3sConfig struct {
	Token                string
	Server               string   `yaml:",omitempty"`
	FlannelIface         string   `json:"-" yaml:"flannel-iface,omitempty"`
	ClusterCidr          string   `json:"cluster-cidr" yaml:"cluster-cidr,omitempty"`
	ServiceCidr          string   `json:"service-cidr" yaml:"service-cidr,omitempty"`
	AdvertiseAddr        string   `json:"-" yaml:"advertise-address,omitempty"`
	ClusterDomain        string   `json:"cluster-domain" yaml:"cluster-domain,omitempty"`
	ClusterDNS           string   `json:"cluster-dns" yaml:"cluster-dns,omitempty"`
	WriteKubeconfigMode  string   `json:"-" yaml:"write-kubeconfig-mode,omitempty"`
	NodeIP               string   `json:"-" yaml:"node-ip,omitempty"`
	BindAddress          string   `json:"-" yaml:"bind-address,omitempty"`
	ClusterInit          bool     `json:"-" yaml:"cluster-init,omitempty"`
	NodeLabels           []string `json:"node-label" yaml:"node-label,omitempty"`
	FlannelBackend       string   `json:"flannel-backend" yaml:"flannel-backend,omitempty"`
	DisableNetworkPolicy bool     `json:"disable-network-policy" yaml:"disable-network-policy,omitempty"`
	// NodeTaints are used to taint the node with key=value:effect.
	// By default, server node is tainted with a couple of taints if number of agents nodes more than 0.
	NodeTaints                []string `json:"node-taint" yaml:"node-taint,omitempty"`
	KubeleteArgs              []string `json:"kubelet-arg" yaml:"kubelet-arg,omitempty"`
	KubeControllerManagerArgs []string `json:"kube-controller-manager-arg" yaml:"kube-controller-manager-arg,omitempty"`
	KubeAPIServerArgs         []string `json:"kube-apiserver-arg" yaml:"kube-apiserver-arg,omitempty"`
	ExternalNodeIP            string   `json:"-" yaml:"node-external-ip,omitempty"`
	DisableCloudController    bool     `json:"disable-cloud-controller" yaml:"disable-cloud-controller,omitempty"`
	// Disable is a list to disable some services.
	Disable []string `yaml:"disable,omitempty"`
}

func (c *CompletedConfig) render() ([]byte, error) {
	return yaml.Marshal(&c.Config)
}

func (k *K3sConfig) WithoutDuplicates() *K3sConfig {
	k.NodeLabels = slices.Compact(k.NodeLabels)
	k.Disable = slices.Compact(k.Disable)
	k.NodeTaints = slices.Compact(k.NodeTaints)
	k.KubeleteArgs = slices.Compact(k.KubeleteArgs)
	k.KubeControllerManagerArgs = slices.Compact(k.KubeControllerManagerArgs)
	k.KubeAPIServerArgs = slices.Compact(k.KubeAPIServerArgs)

	return k
}
