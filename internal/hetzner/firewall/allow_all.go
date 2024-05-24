package firewall

import (
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AllowAllRules struct {
	rules []*RuleConfig
}

func NewAllowAllRules() *AllowAllRules {
	rules := make([]*RuleConfig, 0)
	rules = append(rules, &RuleConfig{
		Protocol:    "icmp",
		Description: "Allow ICMP for cluster nodes",
		Port:        "",
		SourceIps: []string{
			"0.0.0.0/0",
			"::/0",
		},
	})
	rules = append(rules, &RuleConfig{
		Protocol:    string(hcloud.FirewallRuleProtocolTCP),
		Description: "Allow all tcp for cluster nodes",
		Port:        "any",
		SourceIps: []string{
			"0.0.0.0/0",
			"::/0",
		},
	})
	rules = append(rules, &RuleConfig{
		Protocol:    string(hcloud.FirewallRuleProtocolUDP),
		Description: "Allow all udp for cluster nodes",
		Port:        "any",
		SourceIps: []string{
			"0.0.0.0/0",
			"::/0",
		},
	})

	return &AllowAllRules{
		rules: rules,
	}
}

func (a *AllowAllRules) WithPulumiSourceIPs(ips pulumi.StringArray) *AllowAllRules {
	for id := range a.rules {
		a.rules[id].pulumiSourceIps = ips
	}

	return a
}

func (a *AllowAllRules) Rules() []*RuleConfig {
	return a.rules
}
