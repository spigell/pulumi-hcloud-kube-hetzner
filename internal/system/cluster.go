package system

import (
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Cluster []*System

type Deployed struct {
	Wireguard *WgCluster
}

func (c *Cluster) Up(wgInfo map[string]*wireguard.WgConfig, deps *hetzner.Deployed) (*Deployed, error) {
	provisionedWGPeers := c.NewWgCluster(wgInfo, deps.Servers)

	for _, v := range *c {
		s, err := v.Up(deps.Servers[v.ID])
		if err != nil {
			return nil, fmt.Errorf("error while provisioning system %s: %w", v.ID, err)
		}

		for k, module := range s.OS.Modules() {
			if k == "wireguard" {
				provisionedWGPeers.Peers[v.ID] = module.Value().(pulumi.AnyOutput)
			}
		}
	}

	return &Deployed{
		Wireguard: provisionedWGPeers,
	}, nil
}
