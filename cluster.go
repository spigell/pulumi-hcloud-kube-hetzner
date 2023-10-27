package main

import (
	"errors"
	"pulumi-hcloud-kube-hetzner/internal/config"
	"pulumi-hcloud-kube-hetzner/internal/hetzner"
	"pulumi-hcloud-kube-hetzner/internal/system"
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Cluster struct {
	SysCluster system.Cluster
	Hetzner    *hetzner.Hetzner
}

func NewCluster(ctx *pulumi.Context, config *config.Config, keyPair *keypair.ECDSAKeyPair) (*Cluster, error) {
	leader, followers, err := config.MergeNodesConfiguration()
	if err != nil {
		return nil, err
	}

	allNodes := followers
	allNodes = append(allNodes, leader)

	if err := config.Validate(allNodes); err != nil {
		return nil, err
	}

	infra := hetzner.New(ctx, allNodes)

	s := make(system.Cluster, 0)
	for _, node := range allNodes {
		sys := system.New(ctx, node.ID, keyPair)
		os := sys.MicroOS()
		if node.Wireguard.Enabled {
			os.SetWireguard(node.Wireguard)
			fw, err := infra.FirewallConfigByIDOrRole(node.ID)
			if err != nil {
				if !errors.Is(err, hetzner.ErrFirewallDisabled) {
					return nil, err
				}
			}

			if fw != nil {
				fw.AddRules(os.Wireguard().HetznerRules())
			}
		}
		s = append(s, sys.SetOS(os))
	}

	return &Cluster{
		Hetzner:    infra,
		SysCluster: s,
	}, nil
}
