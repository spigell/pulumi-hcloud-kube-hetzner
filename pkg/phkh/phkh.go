package phkh

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/sanity-io/litter"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PHKH struct {
	config  *config.Config
	cluster *Cluster
	state   *State

}

func New(ctx *pulumi.Context) (*PHKH, error) {
	cfg := config.New(ctx)
	state, err := newState(ctx)
	if err != nil {
		return nil, err
	}

	keys, err := state.sshKeyPair()
	if err != nil {
		return nil, err
	}

	cluster, err := newCluster(ctx, cfg, keys)
	if err != nil {
		return nil, err
	}

	return &PHKH{
		config:  cfg,
		cluster: cluster,
		state:   state,
	}, nil
}

func (c *PHKH) Up() error {
	hetznerInfo, err := c.state.hetznerInfra()
	if err != nil {
		return err
	}

	keys, err := c.state.sshKeyPair()
	if err != nil {
		return err
	}

	wgInfo, err := c.state.wgInfo()
	if err != nil {
		return err
	}

	cloud, err := c.cluster.Hetzner.Up(hetznerInfo, keys)
	if err != nil {
		return err
	}
	sys, err := c.cluster.SysCluster.Up(wgInfo, cloud)
	if err != nil {
		return err
	}

	c.state.exportHetznerInfra(cloud)
	c.state.exportWGInfo(sys.Wireguard)
	c.state.exportSSHKeyPair(keys)

	return nil
}

// DumpConfig returns a string representation of the parsed config.
// This is useful for debugging.
func (c *PHKH) DumpConfig() string {
	return litter.Sdump(c.config)
}
