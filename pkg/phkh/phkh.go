package phkh

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PHKH struct {
	Config  *config.Config
	State   *State
	Cluster *Cluster
}

func New(ctx *pulumi.Context) (*PHKH, error) {
	cfg := config.New(ctx)
	state, err := NewState(ctx)
	if err != nil {
		return nil, err
	}

	keys, err := state.SSHKeyPair()
	if err != nil {
		return nil, err
	}

	cluster, err := NewCluster(ctx, cfg, keys)
	if err != nil {
		return nil, err
	}

	return &PHKH{
		Config:  cfg,
		State:   state,
		Cluster: cluster,
	}, nil
}

func (c *PHKH) Up() error {
	hetznerInfo, err := c.State.HetznerInfra()
	if err != nil {
		return err
	}

	keys, err := c.State.SSHKeyPair()
	if err != nil {
		return err
	}

	wgInfo, err := c.State.WGInfo()
	if err != nil {
		return err
	}

	cloud, err := c.Cluster.Hetzner.Up(hetznerInfo, keys)
	if err != nil {
		return err
	}
	sys, err := c.Cluster.SysCluster.Up(wgInfo, cloud)
	if err != nil {
		return err
	}

	c.State.ExportHetznerInfra(cloud)
	c.State.ExportWGInfo(sys.Wireguard)
	c.State.ExportSSHKeyPair(keys)

	return nil
}
