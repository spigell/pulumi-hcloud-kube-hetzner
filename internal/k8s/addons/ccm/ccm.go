package ccm

import (
	hvariables "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/k8sconfig/helm"
)

const (
	Name             = "ccm"
	enabledByDefault = false
	HelmRepo         = "https://charts.hetzner.cloud"
)

type Config struct {
	// Enabled is a flag to enable or disable hcloud CCM.
	Enabled bool
	Helm    *helm.Config
	// LoadbalancersEnabled is a flag to enable or disable loadbalancers management. Note: internal loadbalancer for k3s will be disabled.
	LoadbalancersEnabled bool `json:"loadbalancers-enabled" yaml:"loadbalancers-enabled"`
	// DefaultloadbalancerLocation is a default location for the loadbancers.
	LoadbalancersDefaultLocation string `json:"loadbalancers-default-location" yaml:"loadbalancers-default-location"`
	// Token is a hcloud token to access hcloud API for CCM.
	Token string
}

type CCM struct {
	enabled                      bool
	clusterCIDR                  string
	helm                         *helm.Config
	defaultLoadbalancersLocation string
	loadbalancersEnabled         bool
	loadbalancersUsePrivateIP    bool
	token                        string
	networking                   bool
	controllers                  []string
}

func New(cfg *Config) *CCM {
	m := &CCM{}

	if cfg == nil {
		cfg = &Config{
			Enabled: enabledByDefault,
		}
	}

	if cfg.LoadbalancersDefaultLocation == "" {
		m.defaultLoadbalancersLocation = hvariables.DefaultLocation
	}

	m.controllers = []string{
		"cloud-node-lifecycle-controller", "node-route-controller", "service-lb-controller",
	}
	m.helm = cfg.Helm
	m.enabled = cfg.Enabled
	m.token = cfg.Token
	m.loadbalancersEnabled = cfg.LoadbalancersEnabled
	m.defaultLoadbalancersLocation = cfg.LoadbalancersDefaultLocation

	return m
}

func (m *CCM) Helm() *helm.Config {
	return m.helm
}

func (m *CCM) SetHelm(h *helm.Config) {
	m.helm = h
}

func (m *CCM) Name() string {
	return Name
}

func (m *CCM) Enabled() bool {
	return m.enabled
}

func (m *CCM) LoadbalancersEnabled() bool {
	return m.loadbalancersEnabled
}

func (m *CCM) Supported(distr string) bool {
	switch distr {
	case distributions.K3SDistrName:
		return true
	default:
		return false
	}
}

func (m *CCM) SetClusterCIDR(cidr string) {
	m.clusterCIDR = cidr
}

func (m *CCM) WithLoadbalancerPrivateIPUsage() *CCM {
	m.loadbalancersUsePrivateIP = true

	return m
}

func (m *CCM) WithEnableNodeController() *CCM {
	m.controllers = append(m.controllers, "cloud-node-controller")

	return m
}

func (m *CCM) WithNetworking() *CCM {
	m.networking = true

	return m
}
