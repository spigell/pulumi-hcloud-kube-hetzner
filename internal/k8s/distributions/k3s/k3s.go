package k3s

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
)

const (
	DistrName = "k3s"
)

type K3S struct {
	ctx *pulumi.Context

	addons []addons.Addon
}

func New(ctx *pulumi.Context) *K3S {
	return &K3S{
		ctx: ctx,
	}
}

func (k *K3S) WithAddons(addons []addons.Addon) *K3S {
	k.addons = addons

	return k
}

func (k *K3S) Validate() error {
	for _, addon := range k.addons {
		if addon.IsEnabled() {
			if !addon.IsSupported(DistrName) {
				return fmt.Errorf("addon %s is not supported for %s", addon.Name(), DistrName)
			}
		}
	}

	return nil
}
