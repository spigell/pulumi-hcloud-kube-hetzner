package config

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/audit"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
)

type Config struct {
	KubeAPIEndpoint *K8SEndpoint    `json:"kube-api-endpoint" yaml:"kube-api-endpoint"`
	AuditLog        *audit.AuditLog `json:"audit-log" yaml:"audit-log"`
	Addons          *addons.Addons
}

type NodeConfig struct {
	// NodeTaints are used to taint the node with key=value:effect.
	// By default, server node is tainted with a couple of taints if number of agents nodes more than 0.
	NodeTaints []string `json:"node-taint" yaml:"node-taint,omitempty"`
	NodeLabels []string `json:"node-label" yaml:"node-label,omitempty"`
}

type K8SEndpoint struct {
	Type     string
	Firewall *BasicFirewall
}

type BasicFirewall struct {
	HetznerPublic *HetnzerBasicFirewall `json:"hetzner-public" yaml:"hetzner-public"`
}

type HetnzerBasicFirewall struct {
	DisallowOwnIP bool     `json:"disallow-own-ip" yaml:"disallow-own-ip"`
	AllowedIps    []string `json:"allowed-ips" yaml:"allowed-ips"`
}

func (k *Config) WithInited() *Config {
	if k.Addons == nil {
		k.Addons = &addons.Addons{}
	}

	if k.KubeAPIEndpoint == nil {
		k.KubeAPIEndpoint = &K8SEndpoint{}
	}

	if k.KubeAPIEndpoint.Type == "" {
		k.KubeAPIEndpoint.Type = variables.PublicCommunicationMethod.String()
	}

	if k.KubeAPIEndpoint.Firewall == nil {
		k.KubeAPIEndpoint.Firewall = &BasicFirewall{}
	}

	if k.KubeAPIEndpoint.Firewall.HetznerPublic == nil {
		k.KubeAPIEndpoint.Firewall.HetznerPublic = &HetnzerBasicFirewall{}
	}

	if k.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps == nil {
		k.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps = firewall.ICMPRule.SourceIps
	}

	if k.AuditLog == nil {
		k.AuditLog = &audit.AuditLog{}
	}

	return k
}
