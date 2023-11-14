package phkh

import (
	"errors"
	"fmt"
	"net"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/sshd"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	externalip "github.com/glendc/go-external-ip"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	defaultKube = "k3s"
)

// Cluster is a collection of Hetzner, System and k8s clusters.
type Compiled struct {
	SysCluster system.Cluster
	Hetzner    *hetzner.Hetzner
	K8S        *k8s.K8S
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

	ip, err := externalIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get external IP: %w", err)
	}

	s := make(system.Cluster, 0)
	for _, node := range nodes {
		sys := system.New(ctx, node.ID, keyPair).
			WithCommunicationMethod(variables.PublicCommunicationMethod).
			WithK8SEndpointType(config.K8S.KubeApiEndpoint.Type)

		if node.Leader {
			sys.MarkAsLeader()
		}
		os := sys.MicroOS()

		if config.Network.Hetzner.Enabled {
			sys.WithCommunicationMethod(variables.InternalCommunicationMethod)
		}

		fw, err := infra.FirewallConfigByID(node.ID, infra.FindInPools(node.ID))
		if err != nil {
			if !errors.Is(err, hetzner.ErrFirewallDisabled) {
				return nil, fmt.Errorf("failed to get firewall config for node: %w", err)
			}
		}

		// Add firewall rules for SSH access from my IP
		if fw != nil && node.Server.Firewall.Hetzner.SSH.DisallowOwnIp == false {
			fw.AddRules(sshd.HetznerRulesWithSources([]string{ip2Net(ip)}))
		}

		switch kube {
		case defaultKube:
			os.SetupSSHD(&sshd.Config{
				// TODO: make it discoverable from k3s module
				AcceptEnv: "INSTALL_K3S_*",
			})
			os.AddK3SModule(node.Role, node.K3s)

			if fw != nil && ! config.K8S.KubeApiEndpoint.Firewall.HetznerPublic.DisallowOwnIp && node.Role == variables.ServerRole {
				fw.AddRules(k3s.HetznerRulesWithSources([]string{ip2Net(ip)}))
			}

			// Firewall rule is needed only for public networks
			if config.K8S.KubeApiEndpoint.Type == variables.PublicCommunicationMethod {
				if fw != nil {
					if node.Role == variables.ServerRole {
						fw.AddRules(k3s.HetznerRulesWithSources(config.K8S.KubeApiEndpoint.Firewall.HetznerPublic.AllowedIps))
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
			sys.WithCommunicationMethod(variables.WgCommunicationMethod)

			if fw != nil {
				fw.AddRules(os.Wireguard().HetznerRulesWithSources(config.Network.Wireguard.Firewall.Hetzner.AllowedIps))

				if !config.Network.Wireguard.Firewall.Hetzner.DisallowOwnIp {
					fw.AddRules(os.Wireguard().HetznerRulesWithSources([]string{ip2Net(ip)}))
				}
			}
		}

		s = append(s, sys.WithOS(os))
	}

	kubeCluster := k8s.New(ctx)

	return &Compiled{
		Hetzner:    infra,
		SysCluster: s,
		K8S:        kubeCluster,
	}, nil
}

func externalIP() (net.IP, error) {
	consensus := externalip.DefaultConsensus(nil, nil)
	consensus.UseIPProtocol(4)

	return consensus.ExternalIP()
}

func ip2Net(ip net.IP) string {
	return ip.String() + "/32"
}