package system

import (
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Cluster []*System

type Deployed struct {
	Wireguard *WgCluster
}

func (c *Cluster) Up(wgInfo map[string]*wireguard.WgConfig, deps *hetzner.Deployed) (*Deployed, error) {
	provisionedWGPeers := c.NewWgCluster(wgInfo, deps.Servers)

	kubeDependecies := make(map[string][]pulumi.Resource)

	leaderIPS := map[string]pulumi.StringOutput{
		variables.InternalCommunicationMethod: pulumi.String(deps.Servers[c.Leader().ID].InternalIP).ToStringOutput(),
		variables.DefaultCommunicationMethod:  deps.Servers[c.Leader().ID].Connection.IP,
	}

	if c.Leader().OS.Wireguard() != nil {
		leaderIPS[variables.WgCommunicationMethod] =
			pulumi.String(c.Leader().OS.Wireguard().Self.PrivateAddr).ToStringOutput()
	}

	for _, v := range *c {
		// Cluster is sorted by seniority.
		// So, agents and non-leader servers will wait for leader to be ready.
		// After that, agents will wait for non-leader servers.
		v.kubeDependecies = kubeDependecies

		for k, module := range v.OS.Modules() {
			if k == variables.K3s {
				v.OS.Modules()[k] = module.(*k3s.K3S).WithSysInfo(v.info).WithLeaderIp(
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
			// Cluster is sorted by seniority.
			// So, workers and non-leader nodes will wait for leader to be ready.
			if k == variables.K3s {
				if v.ID == c.Leader().ID {
					kubeDependecies["leader"] = module.Resources()
				}
			}
		}
	}

	return &Deployed{
		Wireguard: provisionedWGPeers,
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
