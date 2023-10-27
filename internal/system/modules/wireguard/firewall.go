package wireguard

import (
	hetzner "pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"strconv"
)

var FWAllowedIps = []string{
	"0.0.0.0/0",
	"::/0",
}

func (w *Wireguard) HetznerRules() []*hetzner.Rule {
	return []*hetzner.Rule{
		{
			Protocol:    "udp",
			Description: "Allow Wireguard",
			Port:        strconv.Itoa(w.ListenPort),
			SourceIps:   w.Config.Firewall.Hetzner.AllowedIps,
		},
	}
}
