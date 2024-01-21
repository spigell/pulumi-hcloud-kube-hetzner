package k8s

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

type Runner struct {
	ctx *program.Context

	addons  []addons.Addon
	manager *manager.ClusterManager
}

func NewRunner(ctx *program.Context, addons []addons.Addon) *Runner {
	return &Runner{
		ctx:    ctx,
		addons: addons,
	}
}

func (r *Runner) WithClusterManager(m *manager.ClusterManager) *Runner {
	r.manager = m

	return r
}

func (r *Runner) Run(prov *kubernetes.Provider) error {
	for _, addon := range r.addons {
		if addon.Enabled() {
			if err := addon.Manage(r.ctx, prov, r.manager); err != nil {
				return err
			}
		}
	}

	return nil
}
