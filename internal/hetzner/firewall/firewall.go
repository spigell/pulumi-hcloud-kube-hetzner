package firewall

import (
	"fmt"
	"strconv"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"

	hcloudgo "github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
)

var (
	ICMPRule = &RuleConfig{
		Protocol:    "icmp",
		Description: "Allow ICMP",
		Port:        "",
		SourceIps: []string{
			"0.0.0.0/0",
			"::/0",
		},
	}
	SSHRule = &RuleConfig{
		Protocol:    "tcp",
		Description: "Allow SSH",
		Port:        "22",
		// It can be changed by user
		SourceIps: []string{
			"0.0.0.0/0",
			"::/0",
		},
	}
)

type Firewall struct {
	firewall *hcloud.Firewall

	Config *Config
}

func New(config *Config) *Firewall {
	return &Firewall{
		Config: config,
	}
}

func (f *Firewall) Up(ctx *program.Context, name string) (*Firewall, error) {
	// f.Config.rules = make([]*Rule, 0)
	var rules hcloud.FirewallRuleArray

	if *f.Config.AllowICMP {
		f.Config.rules = append(f.Config.rules, ICMPRule)
	}

	if ssh := f.Config.SSH; *ssh.Allow {
		if ssh.AllowedIps != nil {
			SSHRule.SourceIps = f.Config.SSH.AllowedIps
		}

		if len(SSHRule.SourceIps) > 0 {
			f.Config.rules = append(f.Config.rules, SSHRule)
		}
	}

	if f.Config.AdditionalRules != nil {
		f.Config.rules = append(f.Config.rules, f.Config.AdditionalRules...)
	}

	for _, rule := range f.Config.rules {
		if rule.Protocol == "" {
			rule.Protocol = string(hcloudgo.FirewallRuleProtocolTCP)
		}

		r := hcloud.FirewallRuleArgs{
			Direction:   pulumi.String((hcloudgo.FirewallRuleDirectionIn)),
			Description: pulumi.String(rule.Description),
			Protocol:    pulumi.String(rule.Protocol),
			Port:        pulumi.String(rule.Port),
			SourceIps:   pulumi.ToStringArray(rule.SourceIps),
		}

		if rule.pulumiSourceIps != nil {
			r.SourceIps = rule.pulumiSourceIps
		}

		rules = append(rules, r)
	}

	created, err := hcloud.NewFirewall(ctx.Context(), name, &hcloud.FirewallArgs{
		Name:  pulumi.String(fmt.Sprintf("%s-%s-%s", ctx.Context().Project(), ctx.Context().Stack(), name)),
		Rules: rules,
	}, ctx.Options()...)
	if err != nil {
		return nil, err
	}

	f.firewall = created

	return f, nil
}

func (f *Firewall) Attach(ctx *program.Context, name string, serverIDs pulumi.IntArray) (*hcloud.FirewallAttachment, error) {
	created, err := hcloud.NewFirewallAttachment(ctx.Context(), name, &hcloud.FirewallAttachmentArgs{
		//nolint: gocritic // this is the only way to convert string to int
		FirewallId: f.firewall.ID().ToStringOutput().ApplyT(func(id string) (int, error) {
			return strconv.Atoi(id)
		}).(pulumi.IntOutput),
		ServerIds: serverIDs,
	}, ctx.Options()...)
	if err != nil {
		return nil, err
	}
	return created, nil
}
