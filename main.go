package main

import (
	"pulumi-hcloud-kube-hetzner/internal/config"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx)
		state, err := NewState(ctx, cfg.Organization)
		if err != nil {
			return err
		}

		keys, err := state.SSHKeyPair()
		if err != nil {
			return err
		}

		wgInfo, err := state.WGInfo()
		if err != nil {
			return err
		}

		cluster, err := NewCluster(ctx, cfg, keys)
		if err != nil {
			return err
		}

		cloud, err := cluster.Hetzner.Up(keys)
		if err != nil {
			return err
		}
		sys, err := cluster.SysCluster.Up(wgInfo, cloud)
		if err != nil {
			return err
		}

		state.ExportWGInfo(sys.Wireguard)

		state.ExportSSHKeyPair(keys)

		return nil
	})
}
