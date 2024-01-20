package phkh

import (
	"fmt"

	"github.com/sanity-io/litter"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage/k3stoken"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage/sshkeypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PHKH struct {
	config   *config.Config
	compiled *Compiled
	state    *State
	ctx      *pulumi.Context
}

func New(ctx *pulumi.Context, opts []pulumi.ResourceOption) (*PHKH, error) {
	cfg := config.New(ctx).WithInited()
	state, err := state(ctx)
	if err != nil {
		return nil, err
	}

	compiled, err := compile(ctx, opts, cfg)
	if err != nil {
		return nil, err
	}

	return &PHKH{
		config:   cfg,
		compiled: compiled,
		state:    state,
		ctx:      ctx,
	}, nil
}

func (c *PHKH) Up() error {
	keypair, err := sshkeypair.New(c.ctx)
	if err != nil {
		return err
	}

	token, err := k3stoken.New(c.ctx)
	if err != nil {
		return err
	}

	cloud, err := c.compiled.Hetzner.Up(keypair)
	if err != nil {
		return err
	}
	sys, err := c.compiled.SysCluster.Up(token, cloud)
	if err != nil {
		return err
	}

	switch distr := c.compiled.K8S.Distr(); distr {
	case k3s.DistrName:
		c.state.exportKubeconfig(sys.K3s.KubeconfigForExport)
		c.ctx.Export(k3sTokenKey, pulumi.String(sys.K3s.Token))
		err = c.compiled.K8S.Up(sys.K3s.KubeconfigForUsage, sys.Resources)

		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported kubernetes distribution: %s", distr)
	}

	c.state.exportHetznerInfra(cloud)
	c.ctx.Export(KeyPairKey, pulumi.ToSecret(keypair.PrivateKey()))

	return nil
}

// DumpConfig returns a string representation of the parsed config.
// This is useful for debugging.
func (c *PHKH) DumpConfig() string {
	return litter.Sdump(c.config)
}
