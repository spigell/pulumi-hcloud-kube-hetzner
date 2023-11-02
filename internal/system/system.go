package system

import (
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os/microos"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type System struct {
	ID      string
	ctx     *pulumi.Context
	KeyPair *keypair.ECDSAKeyPair
	OS      os.OperationSystem
}

type SysProvisioned struct {
	OS os.Provisioned
}

func New(ctx *pulumi.Context, id string, pair *keypair.ECDSAKeyPair) *System {
	return &System{
		ID:      id,
		ctx:     ctx,
		KeyPair: pair,
	}
}

func (s *System) MicroOS() *microos.MicroOS {
	os := microos.New(s.ID)

	return os
}

func (s *System) SetOS(os os.OperationSystem) *System {
	s.OS = os

	return s
}

func (s *System) Up(server *hetzner.Server) (*SysProvisioned, error) {
	os, err := s.OS.Up(s.ctx, server)
	if err != nil {
		err = fmt.Errorf("error while preparing: %w", err)
		return nil, err
	}

	//	cfg, err := m.System.ConfigureSSHD("k3s", k3s.GetRequirdSSHDConfig())
	//	if err != nil {
	//		err = fmt.Errorf("error configure sshd service for k3s cluster: %w", err)
	//		return err
	//	}

	//	reboot, _ := cluster.Os.Reboot([]map[string]pulumi.Resource{pkgs, cfg})
	//
	//	wgCluster, err := cluster.Wireguard.Manage([]map[string]pulumi.Resource{reboot})
	//	if err != nil {
	// 	err = fmt.Errorf("error creating a wireguard cluster: %w", err)
	// 	ctx.Log.Error(err.Error(), nil)
	// 	return err
	// }

	// k3sCluster, err := cluster.K3s.Manage(wgCluster.Peers, []map[string]pulumi.Resource{wgCluster.Resources})
	// if err != nil {
	// 	ctx.Log.Error(err.Error(), nil)
	// 	return err
	// }

	// err = cluster.Firewalls.Manage([]map[string]pulumi.Resource{reboot})
	// if err != nil {
	// 	ctx.Log.Error(err.Error(), nil)
	// 	return err
	// }

	// ctx.Export("os:wireguard:info", wgCluster.ConvertPeersToMapMap())
	// ctx.Export("os:wireguard:config", pulumi.ToSecret(wgCluster.MasterConfig))

	// ctx.Export("os:vpn:address", pulumi.Unsecret(
	// 	pulumi.Sprintf("%s:%d", utils.ExtractValueFromPulumiMapMap(infraLayerNodeInfo, cluster.K3s.Leader.ID, "ip"), wgCluster.ListenPort)),
	//	)

	//	ctx.Export("os:k3s:kubeconfig", pulumi.ToSecret(k3sCluster.Kubeconfig))
	//
	return &SysProvisioned{
		OS: os,
	}, nil
}
