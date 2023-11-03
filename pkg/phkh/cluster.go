package phkh

import (
	"errors"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Cluster is a collection of Hetzner, System and k8s clusters.
type Cluster struct {
	SysCluster system.Cluster
	Hetzner    *hetzner.Hetzner
}

func newCluster(ctx *pulumi.Context, config *config.Config, keyPair *keypair.ECDSAKeyPair) (*Cluster, error) {
	leader, followers, err := config.MergeNodesConfiguration()
	if err != nil {
		return nil, err
	}

	allNodes := followers
	allNodes = append(allNodes, leader)

	if err := config.Validate(allNodes); err != nil {
		return nil, err
	}

	infra := hetzner.New(ctx, allNodes).WithNetwork(config.Network)

	if config.Network.Enabled {
		for _, pool := range config.Nodepools.Agents {
			for _, node := range pool.Nodes {
				// Pools are used only in network mode
				infra.AddToPool(pool.ID, node.ID)
			}
			infra.Network.PickSubnet(pool.ID, network.FromStart)
		}

		for _, pool := range config.Nodepools.Servers {
			for _, node := range pool.Nodes {
				infra.AddToPool(pool.ID, node.ID)
			}
			infra.Network.PickSubnet(pool.ID, network.FromEnd)
		}
	}

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
