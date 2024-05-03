package sshd

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
)

func HetznerRulesWithSources(sources []string) []*firewall.RuleConfig {
	return []*firewall.RuleConfig{
		{
			Protocol:    "tcp",
			Description: "Allow SSH",
			Port:        "22",
			SourceIps:   sources,
		},
	}
}
