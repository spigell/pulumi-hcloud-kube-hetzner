package k3s

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
)

func HetznerRulesWithSources(sources []string) []*firewall.RuleConfig {
	return []*firewall.RuleConfig{
		{
			Protocol:    "tcp",
			Description: "Allow KubeAPI server",
			Port:        "6443",
			SourceIps:   sources,
		},
	}
}
