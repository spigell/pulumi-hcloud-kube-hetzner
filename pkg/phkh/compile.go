package phkh

import (
	"errors"
	"fmt"
	"net"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
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

func preCompile(ctx *pulumi.Context, config *config.Config, nodes []*config.Node) (*Compiled, error) {
	if err := config.Validate(nodes); err != nil {
		return nil, err
	}
	infra := hetzner.New(ctx, nodes).WithNetwork(config.Network.Hetzner).WithNodepools(config.Nodepools)

	kubeCluster := k8s.New(ctx, config.K8S.Addons)
	if err := kubeCluster.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate k8s addons: %w", err)
	}

	return &Compiled{
		Hetzner: infra,
		K8S:     kubeCluster,
	}, nil
}

// compile create plan of infrastructure and required steps.
// This need to be refactored.
func compile(ctx *pulumi.Context, token string, config *config.Config, keyPair *keypair.ECDSAKeyPair) (*Compiled, error) { // //lint: gocognit,gocyclo,
	// This is the only supported kubernetes distribution right now.
	kube := defaultKube

	// Since token is part of k3s config the easiest method to pass the token to k3s module is via global value.
	// However, we do not want to expose token to the user in DumpConfig().
	config.Defaults.Global.K3s.K3S.Token = token

	nodes, err := config.Nodes()
	if err != nil {
		return nil, err
	}

	compiled, err := preCompile(ctx, config, nodes)
	if err != nil {
		return nil, err
	}

	ip, err := externalIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get external IP: %w", err)
	}

	s := make(system.Cluster, 0)
	for _, node := range nodes {
		fw, err := compiled.Hetzner.FirewallConfigByID(node.ID, compiled.Hetzner.FindInPools(node.ID))
		if err != nil {
			if !errors.Is(err, hetzner.ErrFirewallDisabled) {
				return nil, fmt.Errorf("failed to get firewall config for node: %w", err)
			}
		}

		sys := system.New(ctx, node.ID, keyPair).WithK8SEndpointType(config.K8S.KubeAPIEndpoint.Type)
		os := sys.MicroOS()

		// Mark node as leader for cluster
		if node.Leader {
			sys.MarkAsLeader()
		}

		// Network type
		switch {
		// WG over private network
		case config.Network.Hetzner.Enabled && config.Network.Wireguard.Enabled:
			sys.WithCommunicationMethod(variables.WgCommunicationMethod)
			os.SetupWireguard(config.Network.Wireguard)
		// Plain private network
		case config.Network.Hetzner.Enabled:
			sys.WithCommunicationMethod(variables.InternalCommunicationMethod)
		// WG over public network
		case config.Network.Wireguard.Enabled:
			sys.WithCommunicationMethod(variables.WgCommunicationMethod)
			os.SetupWireguard(config.Network.Wireguard)
		// By default use public network
		default:
			sys.WithCommunicationMethod(variables.PublicCommunicationMethod)
		}

		// Firewall
		switch {
		case fw == nil:
			ctx.Log.Debug(fmt.Sprintf("Firewall is disabled for node %s", node.ID), nil)

		// Basic wireguard rules
		case sys.CommunicationMethod() == variables.WgCommunicationMethod:
			if allowedIPs := config.Network.Wireguard.Firewall.Hetzner.AllowedIps; len(allowedIPs) > 0 {
				fw.AddRules(os.Wireguard().HetznerRulesWithSources(allowedIPs))
			}

			if !config.Network.Wireguard.Firewall.Hetzner.DisallowOwnIP {
				fw.AddRules(os.Wireguard().HetznerRulesWithSources([]string{ip2Net(ip)}))
			}
		}

		// Add firewall rules for SSH access from my IP
		if fw != nil && !node.Server.Firewall.Hetzner.SSH.DisallowOwnIP {
			fw.AddRules(sshd.HetznerRulesWithSources([]string{ip2Net(ip)}))
		}

		switch kube {
		case defaultKube:
			os.AddK3SModule(node.Role, node.K3s)
			os.SetupSSHD(&sshd.Config{
				// TODO: make it discoverable from k3s module
				AcceptEnv: "INSTALL_K3S_*",
			})
			configureFwForK3s(fw, config, node, ip)

			for _, addon := range compiled.K8S.Addons() {
				if addon.Enabled() {
					switch name := addon.Name(); name {
					case ccm.Name:
						// Addon has support for k3s already.
						// It was validated.
						configureK3SNodeForHCCM(ctx, sys, addon.(*ccm.CCM), node)
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

		s = append(s, sys.WithOS(os))
	}

	var distr distributions.Distribution

	//nolint: gocritic
	switch kube {
	case defaultKube:
		distr = compiled.K8S.K3S().WithAddons(compiled.K8S.Addons())
		if err := distr.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate k3s cluster: %w", err)
		}

		for _, addon := range compiled.K8S.Addons() {
			//nolint: gocritic
			switch name := addon.Name(); name {
			case ccm.Name:
				configureHCCMForK3S(ctx, s.Leader(), addon.(*ccm.CCM))
			}
		}
	}

	compiled.SysCluster = s
	return compiled, nil
}

func externalIP() (net.IP, error) {
	consensus := externalip.DefaultConsensus(nil, nil)
	consensus.UseIPProtocol(4)

	return consensus.ExternalIP()
}

func ip2Net(ip net.IP) string {
	return ip.String() + "/32"
}

func configureFwForK3s(fw *firewall.Config, config *config.Config, node *config.Node, myIP net.IP) *firewall.Config {
	if !config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.DisallowOwnIP && node.Role == variables.ServerRole {
		fw.AddRules(k3s.HetznerRulesWithSources([]string{ip2Net(myIP)}))
	}

	// Firewall rule is needed only for public networks
	if config.K8S.KubeAPIEndpoint.Type == variables.PublicCommunicationMethod.String() {
		if node.Role == variables.ServerRole && len(config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps) > 0 {
			fw.AddRules(k3s.HetznerRulesWithSources(config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps))
		}
	}
	return fw
}

func configureK3SNodeForHCCM(ctx *pulumi.Context, sys *system.System, addon *ccm.CCM, node *config.Node) {
	if sys.CommunicationMethod().HetznerBased() {
		ctx.Log.Debug("Hetzner CCM is enabled in hetzner mode, force set external controller kubelet", nil)
		node.K3s.K3S.KubeletArgs = append(node.K3s.K3S.KubeletArgs, "cloud-provider=external")
	}

	if node.Role == variables.ServerRole {
		if sys.CommunicationMethod().HetznerBased() {
			ctx.Log.Debug("Hetzner CCM is enabled with hetzner mode, force disabling built-in cloud-controller", nil)
			node.K3s.K3S.DisableCloudController = true
		}

		if addon.LoadbalancersEnabled() {
			ctx.Log.Debug("Hetzner CCM is enabled with LB support, force disabling built-in klipper lb", nil)
			node.K3s.K3S.Disable = append(node.K3s.K3S.Disable, "servicelb")
		}
	}
}

func configureHCCMForK3S(ctx *pulumi.Context, leader *system.System, addon *ccm.CCM) {
	addon.SetClusterCIDR(leader.OS.Modules()[defaultKube].(*k3s.K3S).Config.K3S.ClusterCidr)

	if leader.CommunicationMethod().HetznerBased() {
		ctx.Log.Debug("Hetzner CCM is enabled with hetzner mode, enabling node controller", nil)
		addon.WithEnableNodeController()
	}

	// Private network is validated already. It is present and enabled.
	if addon.Networking() {
		addon.WithLoadbalancerPrivateIPUsage()
	}
}
