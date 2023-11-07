package hetzner

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
)

type InterconnectFirewall struct {
	Config *firewall.Config
	Ips    pulumi.StringArray
	Ids    pulumi.IntArray
}

func NewInterconnectFirewall() *InterconnectFirewall {
	return &InterconnectFirewall{
		Ips: make(pulumi.StringArray, 0),
		Ids: make(pulumi.IntArray, 0),
		Config: &firewall.Config{
			Enabled: true,
			SSH: &firewall.SSH{
				Allow: false,
			},
			AllowICMP: false,
		},
	}
}

func (i *InterconnectFirewall) Up(ctx *pulumi.Context) error {
	i.Config.AddRules(firewall.NewAllowAllRules().WithPulumiSourceIPs(i.Ips).Rules())
	internalFW, err := firewall.New(i.Config).Up(ctx, "interconnect")
	if err != nil {
		return err
	}

	_, err = internalFW.Attach(ctx, "interconnect", i.Ids)
	if err != nil {
		return err
	}

	return nil
}
