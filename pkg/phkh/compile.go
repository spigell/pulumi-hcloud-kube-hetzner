package phkh

import (
	"errors"
	"fmt"
	"net"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/ccm"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions"
	distrK3S "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/sshd"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	externalip "github.com/glendc/go-external-ip"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k3supgrader "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/k3s-upgrade-controller"
)

const (
	defaultKube = distrK3S.DistrName
	// This is the only supported kubernetes distribution right now.
	kube = defaultKube
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

	nodeMap := make(map[string]*manager.Node)
	for _, node := range nodes {
		// By default, use default taints for server node if they are not set and agents nodes exist.
		if node.Role == variables.ServerRole &&
			!node.K3s.DisableDefaultsTaints &&
			len(node.K3s.K3S.NodeTaints) == 0 &&
			len(config.Nodepools.Agents) > 0 {
			node.K3s.K3S.NodeTaints = k3s.DefaultTaints[variables.ServerRole]
		}

		upgradeLabel := fmt.Sprintf("%s=%s", k3supgrader.ControlLabelKey, "false")
		if upgrader := config.K8S.Addons.K3SSystemUpgrader; upgrader != nil {
			upgradeLabel = fmt.Sprintf("%s=%t", k3supgrader.ControlLabelKey, upgrader.Enabled)
		}

		nodeMap[node.ID] = &manager.Node{
			ID:     node.Server.Hostname,
			Taints: node.K3s.K3S.NodeTaints,
			Labels: append(append(node.K3s.K3S.NodeLabels, k3s.NodeManagedLabel), upgradeLabel),
		}
	}

	kubeCluster := k8s.New(ctx, config.K8S.Addons, nodeMap)
	if err := kubeCluster.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate k8s cluster: %w", err)
	}

	return &Compiled{
		Hetzner: infra,
		K8S:     kubeCluster,
	}, nil
}

// compile creates the plan of infrastructure and required steps.
// This need to be refactored.
func compile(ctx *pulumi.Context, token string, config *config.Config, keyPair *keypair.ECDSAKeyPair) (*Compiled, error) { // //lint: gocognit,gocyclo,
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
		fw, err := fwConfig(ctx, compiled, node.ID)
		if err != nil {
			return nil, err
		}

		sys := system.New(ctx, node.ID, keyPair).WithK8SEndpointType(config.K8S.KubeAPIEndpoint.Type)
		os := sys.MicroOS()

		// Network type
		switch {
		// WG over private network
		case config.Network.Hetzner.Enabled && config.Network.Wireguard.Enabled:
			sys.WithCommunicationMethod(variables.WgCommunicationMethod)
			os.SetupWireguard(config.Network.Wireguard)
			fwWithWGRUles(fw, config, os.Wireguard(), ip)
			fwWithSSHRules(fw, node, ip)
		// Plain private network
		case config.Network.Hetzner.Enabled:
			sys.WithCommunicationMethod(variables.InternalCommunicationMethod)
			fwWithSSHRules(fw, node, ip)
		// WG over public network
		case config.Network.Wireguard.Enabled:
			sys.WithCommunicationMethod(variables.WgCommunicationMethod)
			os.SetupWireguard(config.Network.Wireguard)
			fwWithWGRUles(fw, config, os.Wireguard(), ip)
			fwWithSSHRules(fw, node, ip)
		// By default use public network
		default:
			sys.WithCommunicationMethod(variables.PublicCommunicationMethod)
			fwWithSSHRules(fw, node, ip)
		}

		switch kube {
		case defaultKube:
			configureFwForK3s(fw, config, node, ip)
			configureOSForK3S(os, node)

			for _, addon := range compiled.K8S.Addons() {
				if addon.Enabled() {
					//nolint: gocritic
					switch name := addon.Name(); name {
					case ccm.Name:
						// Addon has support for k3s already.
						// It was validated.
						configureK3SNodeForHCCM(ctx, sys, addon.(*ccm.CCM), node)
					}
				}
			}
		default:
			return nil, errors.New("unknown kubernetes distribution")
		}

		s = append(s, sys.WithOS(os))
	}

	// The first node is always leader
	//nolint: gosec // G602
	s[0].MarkAsLeader()

	compiled.SysCluster = s

	if err := compiled.k8sCompile(ctx); err != nil {
		return nil, err
	}

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

func fwConfig(ctx *pulumi.Context, compiled *Compiled, id string) (*firewall.Config, error) {
	fw, err := compiled.Hetzner.FirewallConfigByID(id, compiled.Hetzner.FindInPools(id))
	if err != nil {
		if !errors.Is(err, hetzner.ErrFirewallDisabled) {
			return nil, fmt.Errorf("failed to get firewall config for node: %w", err)
		}
		// Create empty firewall config if firewall is disabled.
		ctx.Log.Debug("firewall is disabled for node", nil)
		fw = &firewall.Config{}
	}

	return fw, nil
}

func fwWithWGRUles(fw *firewall.Config, config *config.Config, wg *wireguard.Wireguard, ip net.IP) {
	if allowedIPs := config.Network.Wireguard.Firewall.Hetzner.AllowedIps; len(allowedIPs) > 0 {
		fw.AddRules(wg.HetznerRulesWithSources(allowedIPs))
	}

	if !config.Network.Wireguard.Firewall.Hetzner.DisallowOwnIP {
		fw.AddRules(wg.HetznerRulesWithSources([]string{ip2Net(ip)}))
	}
}

func fwWithSSHRules(fw *firewall.Config, node *config.Node, ip net.IP) {
	// Add firewall rules for SSH access from my IP
	if !node.Server.Firewall.Hetzner.SSH.DisallowOwnIP {
		fw.AddRules(sshd.HetznerRulesWithSources([]string{ip2Net(ip)}))
	}
}

func configureFwForK3s(fw *firewall.Config, config *config.Config, node *config.Node, myIP net.IP) {
	if !config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.DisallowOwnIP && node.Role == variables.ServerRole {
		fw.AddRules(k3s.HetznerRulesWithSources([]string{ip2Net(myIP)}))
	}

	// Firewall rule is needed only for public networks
	if config.K8S.KubeAPIEndpoint.Type == variables.PublicCommunicationMethod.String() {
		if node.Role == variables.ServerRole && len(config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps) > 0 {
			fw.AddRules(k3s.HetznerRulesWithSources(config.K8S.KubeAPIEndpoint.Firewall.HetznerPublic.AllowedIps))
		}
	}
}

func configureOSForK3S(os os.OperationSystem, node *config.Node) {
	os.AddK3SModule(node.Role, node.K3s)
	os.SetupSSHD(&sshd.Config{
		// TODO: make it discoverable from k3s module
		AcceptEnv: "INSTALL_K3S_*",
	})
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

func (c *Compiled) k8sCompile(ctx *pulumi.Context) error {
	var distr distributions.Distribution

	//nolint: gocritic
	switch kube {
	case defaultKube:
		distr = c.K8S.K3S().WithAddons(c.K8S.Addons())
		if err := distr.Validate(); err != nil {
			return fmt.Errorf("failed to validate k3s distr: %w", err)
		}

		for _, addon := range c.K8S.Addons() {
			//nolint: gocritic
			switch name := addon.Name(); name {
			case ccm.Name:
				configureHCCMForK3S(ctx, c.SysCluster.Leader(), addon.(*ccm.CCM))
			}
		}
	}

	return nil
}

func configureHCCMForK3S(ctx *pulumi.Context, leader *system.System, addon addons.Addon) {
	c := addon.(*ccm.CCM)
	c.SetClusterCIDR(leader.OS.Modules()[defaultKube].(*k3s.K3S).Config.K3S.ClusterCidr)

	if leader.CommunicationMethod().HetznerBased() {
		ctx.Log.Debug("Hetzner CCM is enabled with hetzner mode, enabling node controller", nil)
		c.WithEnableNodeController()
	}

	// Private network is validated already. It is present and enabled.
	if c.Networking() {
		c.WithLoadbalancerPrivateIPUsage()
	}
}
