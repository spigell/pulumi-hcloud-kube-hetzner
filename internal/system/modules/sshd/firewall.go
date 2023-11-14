package sshd

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
)

func HetznerRulesWithSources(sources []string) []*firewall.Rule {
	return []*firewall.Rule{
		{
			Protocol:    "tcp",
			Description: "Allow SSH",
			Port:        "22",
			SourceIps:   sources,
		},
	}
}
