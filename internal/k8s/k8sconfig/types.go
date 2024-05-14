package k8sconfig

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/audit"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
)

type Config struct {
	KubeAPIEndpoint *K8SEndpointConfig    `json:"kube-api-endpoint" yaml:"kube-api-endpoint" mapstructure:"kube-api-endpoint"`
	AuditLog        *audit.AuditLogConfig `json:"audit-log" yaml:"audit-log" mapstructure:"audit-log"`
	Addons          *addons.Config
}

type NodeConfig struct {
	// NodeLabels are used to label the node with key=value.
	NodeLabels []string `json:"node-label" yaml:"node-label,omitempty" mapstructure:"node-label"`
	// NodeTaints configures taint node manager.
	NodeTaints *TaintConfig `json:"node-taint" yaml:"node-taint,omitempty" mapstructure:"node-taint"`
}

type TaintConfig struct {
	// Enable or disable taint management.
	// Default is false.
	Enabled *bool
	// Do not add default taints to the server node.
	// Default is false.
	DisableDefaultsTaints *bool `json:"disable-default-taints" yaml:"disable-default-taints,omitempty" mapstructure:"disable-default-taints"`
	// Taints are used to taint the node with key=value:effect.
	// Default is server node is tainted with a couple of taints if number of agents nodes more than 0.
	// But only if disable-default-taints set to false (default)
	Taints []string
}

type K8SEndpointConfig struct {
	// Type of k8s endpoint: public or private.
	// Default is public.
	Type string
	// Firewall defines configuration for the firewall attached to api access.
	// This is used only for public type since private network considered to be secure.
	Firewall *BasicFirewallConfig
}

type BasicFirewallConfig struct {
	// HetznerPublic is used to describe firewall attached to public k8s api endpoint.
	HetznerPublic *HetnzerBasicFirewallConfig `json:"hetzner-public" yaml:"hetzner-public" mapstructure:"hetzner-public"`
}

type HetnzerBasicFirewallConfig struct {
	// DisallowOwnIP is a security setting that, when enabled, prevents access to the server from deployer own public IP address.
	DisallowOwnIP bool `json:"disallow-own-ip" yaml:"disallow-own-ip" mapstructure:"disallow-own-ip"`

	// AllowedIps specifies a list of IP addresses that are permitted to access the k8s api endpoint.
	// Only traffic from these IPs will be allowed if this list is configured.
	// Default is 0.0.0.0/0 (all ipv4 addresses).
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips" mapstructure:"allowed-ips"`
}

func (k *Config) WithInited() *Config {
	if k.Addons == nil {
		k.Addons = &addons.Config{}
	}

	if k.KubeAPIEndpoint == nil {
		k.KubeAPIEndpoint = &K8SEndpointConfig{}
	}

	if k.KubeAPIEndpoint.Type == "" {
		k.KubeAPIEndpoint.Type = variables.PublicCommunicationMethod.String()
	}

	if k.KubeAPIEndpoint.Firewall == nil {
		k.KubeAPIEndpoint.Firewall = &BasicFirewallConfig{}
	}

	if k.KubeAPIEndpoint.Firewall.HetznerPublic == nil {
		k.KubeAPIEndpoint.Firewall.HetznerPublic = &HetnzerBasicFirewallConfig{}
	}

	if k.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps == nil {
		k.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps = firewall.ICMPRule.SourceIps
	}

	if k.AuditLog == nil {
		k.AuditLog = &audit.AuditLogConfig{}
	}

	return k
}
