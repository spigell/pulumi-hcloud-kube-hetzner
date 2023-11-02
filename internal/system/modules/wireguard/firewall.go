package wireguard

import (
	"strconv"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
)

var FWAllowedIps = []string{
	"0.0.0.0/0",
	"::/0",
}

func (w *Wireguard) HetznerRules() []*firewall.Rule {
	return []*firewall.Rule{
		{
			Protocol:    "udp",
			Description: "Allow Wireguard",
			Port:        strconv.Itoa(w.ListenPort),
			SourceIps:   w.Config.Firewall.Hetzner.AllowedIps,
		},
	}
}
