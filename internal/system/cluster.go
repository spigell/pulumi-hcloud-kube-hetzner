package system

import (
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Cluster []*System

type Deployed struct {
	Wireguard *WgCluster
	K3s       *k3s.Outputs
	Resources []pulumi.Resource
}

func (c *Cluster) Up(wgInfo map[string]*wireguard.WgConfig, deps *hetzner.Deployed) (*Deployed, error) {
	provisionedWGPeers := c.NewWgCluster(wgInfo, deps.Servers)

	// We must wait for all nodes to be ready before we can use kube api.
	// resources is used to keep for all k3s modules Resources().
	// It is enough for waiting.
	resources := make([]pulumi.Resource, 0)

	kubeDependencies := make(map[string][]pulumi.Resource)

	leaderIPS := map[string]pulumi.StringOutput{
		variables.InternalCommunicationMethod: pulumi.String(deps.Servers[c.Leader().ID].InternalIP).ToStringOutput(),
		variables.DefaultCommunicationMethod:  deps.Servers[c.Leader().ID].Connection.IP,
	}

	if c.Leader().OS.Wireguard() != nil {
		leaderIPS[variables.WgCommunicationMethod] = pulumi.String(c.Leader().OS.Wireguard().Self.PrivateAddr).ToStringOutput()
	}

	var k3sOutputs *k3s.Outputs
	for _, v := range *c {
		// Cluster is sorted by seniority.
		// So, agents and non-leader servers will wait for leader to be ready.
		// After that, agents will wait for non-leader servers.
		// kubeDependencies["leader"] is used to wait for leader.
		v.kubeDependecies = kubeDependencies

		for k, module := range v.OS.Modules() {
			if k == variables.K3s {
				v.OS.Modules()[k] = module.(*k3s.K3S).WithSysInfo(v.info).WithLeaderIP(
					leaderIPS[v.info.CommunicationMethod()],
				)
			}
		}

		s, err := v.Up(deps.Servers[v.ID])
		if err != nil {
			return nil, fmt.Errorf("error while provisioning system %s: %w", v.ID, err)
		}

		for k, module := range s.OS.Modules() {
			if k == variables.Wireguard {
				provisionedWGPeers.Peers[v.ID] = module.Value().(pulumi.AnyOutput)
			}
			if k == variables.K3s {
				// Cluster is sorted by seniority.
				// So, workers and non-leader nodes will wait for leader to be ready.
				if v.ID == c.Leader().ID {
					v.kubeDependecies["leader"] = module.Resources()
					resources = append(resources, module.Resources()...)

					k3sOutputs = module.Value().(*k3s.Outputs)

					// Replace leader IP in kubeconfig with IP based on specified method.
					k3sOutputs.Kubeconfig = pulumi.All(
						k3sOutputs.Kubeconfig, leaderIPS[v.info.K8SEndpointType()],
					).ApplyT(
						func(args []interface{}) interface{} {
							kubeconfig := args[0].(*api.Config)
							ip := args[1].(string)
							kubeconfig.Clusters["default"].Server = fmt.Sprintf("https://%s:6443", ip)

							return kubeconfig
						},
					).(pulumi.AnyOutput)
				}
			}
		}
	}

	return &Deployed{
		Wireguard: provisionedWGPeers,
		K3s:       k3sOutputs,
		Resources: resources,
	}, nil
}

// Leader returns the first element of the cluster.
// TO DO: make it better.
func (c *Cluster) Leader() *System {
	for _, v := range *c {
		return v
	}
	return nil
}
