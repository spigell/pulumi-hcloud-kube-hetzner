package phkh

import (
	"errors"
	"fmt"
	"net"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/ccm"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions"
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

// compile create plan of infrastructure and required steps.
// This need to be refactored.
func compile(ctx *pulumi.Context, token string, config *config.Config, keyPair *keypair.ECDSAKeyPair) (*Compiled, error) { //nolint: gocognit,gocyclo
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
			WithK8SEndpointType(config.K8S.KubeAPIEndpoint.Type)

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
		if fw != nil && !node.Server.Firewall.Hetzner.SSH.DisallowOwnIP {
			fw.AddRules(sshd.HetznerRulesWithSources([]string{ip2Net(ip)}))
		}

		switch kube {
		case defaultKube:
			os.SetupSSHD(&sshd.Config{
				// TODO: make it discoverable from k3s module
				AcceptEnv: "INSTALL_K3S_*",
			})
			os.AddK3SModule(node.Role, node.K3s)

			if fw != nil && !config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.DisallowOwnIP && node.Role == variables.ServerRole {
				fw.AddRules(k3s.HetznerRulesWithSources([]string{ip2Net(ip)}))
			}

			// Firewall rule is needed only for public networks
			if config.K8S.KubeAPIEndpoint.Type == variables.PublicCommunicationMethod {
				if fw != nil {
					if node.Role == variables.ServerRole {
						fw.AddRules(k3s.HetznerRulesWithSources(config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps))
					}
				}
			}

			if config.K8S.Addons.CCM.Enabled {
				node.K3s.K3S.KubeletArgs = append(node.K3s.K3S.KubeletArgs, "cloud-provider=external")
			}

			if node.Role == variables.ServerRole {
				// By default, use default taints for server node if they are not set and agents nodes exist.
				if !node.K3s.DisableDefaultsTaints && len(node.K3s.K3S.NodeTaints) == 0 && len(config.Nodepools.Agents) > 0 {
					node.K3s.K3S.NodeTaints = k3s.DefaultTaints[variables.ServerRole]
				}

				if config.K8S.Addons.CCM.Enabled {
					ctx.Log.Debug("Hetzner CCM is enabled, force disabling built-in cloud-controller", nil)
					node.K3s.K3S.DisableCloudController = true

					if config.K8S.Addons.CCM.LoadbalancersEnabled {
						ctx.Log.Debug("Hetzner CCM is enabled with LB support, force disabling built-in klipper lb", nil)
						node.K3s.K3S.Disable = append(node.K3s.K3S.Disable, "servicelb")
					}
				}
			}
		default:
			return nil, errors.New("unknown kubernetes distribution")
		}

		if config.Network.Wireguard.Enabled {
			os.SetupWireguard(config.Network.Wireguard)
			sys.WithCommunicationMethod(variables.WgCommunicationMethod)

			if fw != nil {
				fw.AddRules(os.Wireguard().HetznerRulesWithSources(config.Network.Wireguard.Firewall.Hetzner.AllowedIps))

				if !config.Network.Wireguard.Firewall.Hetzner.DisallowOwnIP {
					fw.AddRules(os.Wireguard().HetznerRulesWithSources([]string{ip2Net(ip)}))
				}
			}
		}

		s = append(s, sys.WithOS(os))
	}
	kubeCluster := k8s.New(ctx, config.K8S.Addons)
	if err := kubeCluster.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate k8s addons: %w", err)
	}

	var distr distributions.Distribution

	switch kube {
	case defaultKube:
		distr = kubeCluster.K3S().WithAddons(kubeCluster.Addons())
		if err := distr.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate k3s cluster: %w", err)
		}

		for _, addon := range kubeCluster.Addons() {
			switch name := addon.Name(); name {
			case ccm.Name:
				a := addon.(*ccm.CCM)
				a.SetClusterCIDR(s.Leader().OS.Modules()[defaultKube].(*k3s.K3S).Config.K3S.ClusterCidr)
				// Private network is validated already. It is present and enabled.
				if config.K8S.Addons.CCM.Networking {
					a.SetLoadbalancerPrivateIPUsage()
				}
			}
		}
	}

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
