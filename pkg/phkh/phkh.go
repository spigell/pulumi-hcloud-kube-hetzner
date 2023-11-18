package phkh

import (
	"fmt"

	"github.com/sanity-io/litter"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions/k3s"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PHKH struct {
	config   *config.Config
	compiled *Compiled
	state    *State
}

func New(ctx *pulumi.Context) (*PHKH, error) {
	cfg := config.New(ctx).WithInited()
	state, err := state(ctx)
	if err != nil {
		return nil, err
	}

	keys, err := state.sshKeyPair()
	if err != nil {
		return nil, err
	}

	token, err := state.k3sToken()
	if err != nil {
		return nil, err
	}

	compiled, err := compile(ctx, token, cfg, keys)
	if err != nil {
		return nil, err
	}

	return &PHKH{
		config:   cfg,
		compiled: compiled,
		state:    state,
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

	cloud, err := c.compiled.Hetzner.Up(hetznerInfo, keys)
	if err != nil {
		return err
	}
	sys, err := c.compiled.SysCluster.Up(wgInfo, cloud)
	if err != nil {
		return err
	}

	switch distr := c.compiled.K8S.Distr(); distr {
	case k3s.DistrName:
		c.state.exportKubeconfig(sys.K3s.KubeconfigForExport)
		c.state.exportK3SToken(sys.K3s.Token)
		err = c.compiled.K8S.Up(sys.K3s.KubeconfigForUsage, sys.Resources)

		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported kubernetes distribution: %s", distr)
	}

	c.state.exportHetznerInfra(cloud)
	c.state.exportSSHKeyPair(keys)
	c.state.exportWGInfo(sys.Wireguard)

	return nil
}

// DumpConfig returns a string representation of the parsed config.
// This is useful for debugging.
func (c *PHKH) DumpConfig() string {
	return litter.Sdump(c.config)
}
