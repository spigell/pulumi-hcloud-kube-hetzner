package k3s

import (
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

const (
	cfgPath = "/etc/rancher/k3s/config.yaml"
)

type Compiled struct {
	Config *K3sConfig
}

type Config struct {
	// version is used to determine if k3s should be upgraded if auto-upgrade is disabled.
	// If the version is changed, k3s will be upgraded.
	Version string
	// [Experimental] clean-data-on-upgrade is used to delete all data while upgrade.
	// This is based on the script https://docs.k3s.io/upgrades/killall
	CleanDataOnUpgrade bool `json:"clean-data-on-upgrade" yaml:"clean-data-on-upgrade,omitempty"`
	// The real config of k3s service.
	K3S *K3sConfig `json:"config" yaml:"config"`
}

type K3sConfig struct { //nolint: revive // This name is better
	// Token used for nodes to join the cluster, generated automatically.
	Token string `json:"-" yaml:"token"`

	// Server specifies the address of the main server node (leader) in the cluster, generated automatically.
	Server string `json:"-" yaml:"server,omitempty"`

	// FlannelIface specifies the network interface that Flannel should use.
	FlannelIface string `json:"-" yaml:"flannel-iface,omitempty"`

	// WriteKubeconfigMode defines the file permission mode for the kubeconfig file on disk.
	WriteKubeconfigMode string `json:"-" yaml:"write-kubeconfig-mode,omitempty"`

	// AdvertiseAddr specifies the IP address that the server uses to advertise to members of the cluster.
	AdvertiseAddr string `json:"-" yaml:"advertise-address,omitempty"`

	// NodeIP specifies the IP address to advertise for this node.
	NodeIP string `json:"-" yaml:"node-ip,omitempty"`

	// BindAddress is the IP address that the server should bind to for API server traffic.
	BindAddress string `json:"-" yaml:"bind-address,omitempty"`

	// ClusterInit indicates whether this node should initialize a new cluster.
	ClusterInit bool `json:"-" yaml:"cluster-init,omitempty"`

	// ExternalNodeIP specifies the external IP address of the node.
	ExternalNodeIP string `json:"-" yaml:"node-external-ip,omitempty"`

	// TLSSanSecurity enables or disables the addition of TLS SANs (Subject Alternative Names).
	TLSSanSecurity bool `json:"-" yaml:"tls-san-security,omitempty"`

	// TLSSan adds specific TLS SANs for securing communication to the K3s server.
	TLSSan string `json:"-" yaml:"tls-san,omitempty"`

	// NodeName specifies the name of the node within the cluster.
	NodeName string `json:"-" yaml:"node-name,omitempty"`

	// ClusterCidr defines the IP range from which pod IPs shall be allocated.
	// Default is 10.141.0.0/16.
	ClusterCidr string `json:"cluster-cidr" yaml:"cluster-cidr,omitempty"`

	// ServiceCidr defines the IP range from which service cluster IPs are allocated.
	// Default is 10.140.0.0/16.
	ServiceCidr string `json:"service-cidr" yaml:"service-cidr,omitempty"`

	// ClusterDomain specifies the domain name of the cluster.
	ClusterDomain string `json:"cluster-domain" yaml:"cluster-domain,omitempty"`

	// ClusterDNS specifies the IP address of the DNS service within the cluster.
	// Default is autopicked.
	ClusterDNS string `json:"cluster-dns" yaml:"cluster-dns,omitempty"`

	// FlannelBackend determines the type of backend used for Flannel, a networking solution.
	FlannelBackend string `json:"flannel-backend" yaml:"flannel-backend,omitempty"`

	// DisableNetworkPolicy determines whether to disable network policies.
	DisableNetworkPolicy bool `json:"disable-network-policy" yaml:"disable-network-policy,omitempty"`

	// KubeletArgs allows passing additional arguments to the kubelet service.
	KubeletArgs []string `json:"kubelet-arg" yaml:"kubelet-arg,omitempty"`

	// KubeControllerManagerArgs allows passing additional arguments to the Kubernetes controller manager.
	KubeControllerManagerArgs []string `json:"kube-controller-manager-arg" yaml:"kube-controller-manager-arg,omitempty"`

	// KubeCloudControllerManagerArgs allows passing additional arguments to the Kubernetes cloud controller manager.
	KubeCloudControllerManagerArgs []string `json:"kube-cloud-controller-manager-arg" yaml:"kube-cloud-controller-manager-arg,omitempty"`

	// KubeAPIServerArgs allows passing additional arguments to the Kubernetes API server.
	KubeAPIServerArgs []string `json:"kube-apiserver-arg" yaml:"kube-apiserver-arg,omitempty"`

	// DisableCloudController determines whether to disable the integrated cloud controller manager.
	// Default is false, but will be true if ccm is enabled.
	DisableCloudController bool `json:"disable-cloud-controller" yaml:"disable-cloud-controller,omitempty"`

	// Disable lists components or features to disable.
	Disable []string `yaml:"disable,omitempty"`

	// NodeLables set labels on registration
	NodeLabels []string `json:"node-label" yaml:"node-label,omitempty"`

	// NodeTaints are used to taint the node with key=value:effect.
	// By default, server node is tainted with a couple of taints if number of agents nodes more than 0.
	NodeTaints []string `json:"node-taint" yaml:"node-taint,omitempty"`
}

func (c *Compiled) render() ([]byte, error) {
	return yaml.Marshal(&c.Config)
}

func (k *K3sConfig) WithoutDuplicates() *K3sConfig {
	k.Disable = slices.Compact(k.Disable)
	k.KubeletArgs = slices.Compact(k.KubeletArgs)
	k.KubeControllerManagerArgs = slices.Compact(k.KubeControllerManagerArgs)
	k.KubeAPIServerArgs = slices.Compact(k.KubeAPIServerArgs)

	return k
}
