package phkh

import (
	"errors"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/sshd"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	defaultKube = "k3s"
)

// Cluster is a collection of Hetzner, System and k8s clusters.
type Cluster struct {
	SysCluster system.Cluster
	Hetzner    *hetzner.Hetzner
}

func newCluster(ctx *pulumi.Context, config *config.Config, keyPair *keypair.ECDSAKeyPair) (*Cluster, error) {
	// This is the only supported kubernetes distribution right now.
	kube := defaultKube

	nodes, err := config.Nodes()
	if err != nil {
		return nil, err
	}

	if err := config.Validate(nodes); err != nil {
		return nil, err
	}

	infra := hetzner.New(ctx, nodes).WithNetwork(config.Network.Hetzner)

	if config.Network.Hetzner.Enabled {
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
	for _, node := range nodes {
		sys := system.New(ctx, node.ID, keyPair).WithCommunicationMethod(variables.DefaultCommunicationMethod)
		os := sys.MicroOS()

		if config.Network.Hetzner.Enabled {
			sys.WithCommunicationMethod(variables.InternalCommunicationMethod)
		}

		switch kube {
		case defaultKube:
			os.SetupSSHD(&sshd.Config{
				// TODO: make it discoverable from k3s module
				AcceptEnv: "INSTALL_K3S_*",
			})
			os.AddK3SModule(node.Role, node.K3s)

			// By default, use default taints for server node if they are not set and agents nodes exist.
			if node.Role == variables.ServerRole &&
				!node.K3s.DisableDefaultsTaints &&
				len(node.K3s.K3S.NodeTaints) == 0 &&
				len(config.Nodepools.Agents) > 0 {
				node.K3s.K3S.NodeTaints = k3s.DefaultTaints[variables.ServerRole]
			}
		default:
			return nil, errors.New("unknown kubernetes distribution")
		}

		if config.Network.Wireguard.Enabled {
			os.SetupWireguard(config.Network.Wireguard)
			fw, err := infra.FirewallConfigByIDOrRole(node.ID)
			if err != nil {
				if !errors.Is(err, hetzner.ErrFirewallDisabled) {
					return nil, err
				}
			}
			sys.WithCommunicationMethod(variables.WgCommunicationMethod)

			if fw != nil {
				fw.AddRules(os.Wireguard().HetznerRules())
			}
		}

		s = append(s, sys.WithOS(os))
	}

	return &Cluster{
		Hetzner:    infra,
		SysCluster: s,
	}, nil
}
