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
type Compiled struct {
	SysCluster system.Cluster
	Hetzner    *hetzner.Hetzner
}

func compile(ctx *pulumi.Context, token string, config *config.Config, keyPair *keypair.ECDSAKeyPair) (*Compiled, error) {
	// This is the only supported kubernetes distribution right now.
	kube := defaultKube

	// Since token is part of k3s config the easiest method to pass the token to k3s module is via global value.
	// However, we do not want to expose token to the user in DumpConfig().
	config.Defaults.Global.K3s.K3S.Token = token

	nodes, err := config.Nodes()
	if err != nil {
		return nil, err
	}

	if err := config.Validate(nodes); err != nil {
		return nil, err
	}

	infra := hetzner.New(ctx, nodes).WithNetwork(config.Network.Hetzner)

	for _, pool := range config.Nodepools.Agents {
		if pool.Nodes[0].Server.Firewall.Hetzner.DedicatedPool() {
			infra.Firewalls[pool.ID] = pool.Config.Server.Firewall.Hetzner
		}

		for _, node := range pool.Nodes {
			infra.AddToPool(pool.ID, node.ID)
		}

		if config.Network.Hetzner.Enabled {
			infra.Network.PickSubnet(pool.ID, network.FromStart)
		}
	}

	for _, pool := range config.Nodepools.Servers {
		if pool.Nodes[0].Server.Firewall.Hetzner.DedicatedPool() {
			infra.Firewalls[pool.ID] = pool.Config.Server.Firewall.Hetzner
		}

		for _, node := range pool.Nodes {
			infra.AddToPool(pool.ID, node.ID)
		}

		if config.Network.Hetzner.Enabled {
			infra.Network.PickSubnet(pool.ID, network.FromEnd)
		}
	}

	s := make(system.Cluster, 0)
	for _, node := range nodes {
		sys := system.New(ctx, node.ID, keyPair).
			WithCommunicationMethod(variables.DefaultCommunicationMethod).
			WithK8SEndpointType(config.K8S.Endpoint.Type)

		if node.Leader {
			sys.MarkAsLeader()
		}
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

			// Firewall rule is needed only for public networks
			if config.K8S.Endpoint.Type == variables.DefaultCommunicationMethod {
				fw, err := infra.FirewallConfigByID(node.ID, infra.FindInPools(node.ID))
				if err != nil {
					if !errors.Is(err, hetzner.ErrFirewallDisabled) {
						return nil, err
					}
				}

				if fw != nil {
					if node.Role == variables.ServerRole {
						fw.AddRules(k3s.HetznerRulesWithSources(config.K8S.Endpoint.Firewall.HetznerPublic.AllowedIps))
					}
				}
			}

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
			fw, err := infra.FirewallConfigByID(node.ID, infra.FindInPools(node.ID))
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

	return &Compiled{
		Hetzner:    infra,
		SysCluster: s,
	}, nil
}
