package k8s

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
)

type Runner struct {
	ctx *pulumi.Context

	addons []addons.Addon
}

func (k *K8S) NewRunner() *Runner {
	return &Runner{
		ctx:    k.ctx,
		addons: k.addons,
	}
}

func (r *Runner) Run(prov *kubernetes.Provider) error {
	for _, addon := range r.addons {
		if addon.Enabled() {
			if err := addon.Manage(r.ctx, prov); err != nil {
				return err
			}
		}
	}

	return nil
}
