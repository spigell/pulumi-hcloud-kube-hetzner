package hetzner

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

type InterconnectFirewall struct {
	Config *firewall.Config
	Ips    pulumi.StringArray
	IDs    pulumi.IntArray
}

func NewInterconnectFirewall() *InterconnectFirewall {
	enabled := true
	allowICMP := true
	allowSSH := false
	return &InterconnectFirewall{
		Ips: make(pulumi.StringArray, 0),
		IDs: make(pulumi.IntArray, 0),
		Config: &firewall.Config{
			Enabled: &enabled,
			SSH: &firewall.SSHConfig{
				Allow: &allowSSH,
			},
			AllowICMP: &allowICMP,
		},
	}
}

func (i *InterconnectFirewall) Up(ctx *program.Context) error {
	i.Config.AddRules(firewall.NewAllowAllRules().WithPulumiSourceIPs(i.Ips).Rules())
	internalFW, err := firewall.New(i.Config).Up(ctx, "interconnect")
	if err != nil {
		return err
	}

	_, err = internalFW.Attach(ctx, "interconnect", i.IDs)
	if err != nil {
		return err
	}

	return nil
}
