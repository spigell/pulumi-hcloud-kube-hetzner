package k3s

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
)

func HetznerRulesWithSources(sources []string) []*firewall.Rule {
	return []*firewall.Rule{
		{
			Protocol:    "tcp",
			Description: "Allow KubeAPI server",
			Port:        "6443",
			SourceIps:   sources,
		},
	}
}
