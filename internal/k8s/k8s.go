package k8s

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions/k3s"
)

type K8S struct {
	ctx    *pulumi.Context
	distr  string
	addons []addons.Addon
}

func New(ctx *pulumi.Context, adds *addons.Addons) *K8S {
	return &K8S{
		ctx:    ctx,
		addons: addons.New(adds),
	}
}

func (k *K8S) K3S() *k3s.K3S {
	k.distr = k3s.DistrName

	return k3s.New(k.ctx)
}

func (k *K8S) Distr() string {
	return k.distr
}

func (k *K8S) Addons() []addons.Addon {
	return k.addons
}

func (k *K8S) Validate() error {
	return addons.Validate(k.addons)
}

func (k *K8S) Up(kubeconfig pulumi.AnyOutput, deps []pulumi.Resource) error {
	prov, err := k.Provider(kubeconfig, deps)

	if err != nil {
		return err
	}

	return k.NewRunner().Run(prov)
}
