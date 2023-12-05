package k8s

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/ccm"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions/k3s"
)

type K8S struct {
	ctx    *pulumi.Context
	distr  string
	addons []addons.Addon

	mgmt *manager.ClusterManager
	runner *Runner
}

func New(ctx *pulumi.Context, adds *addons.Addons, nodes map[string]*manager.Node) *K8S {
	mgmt := manager.New(ctx, nodes)
	addons := addons.New(adds)

	return &K8S{
		ctx:    ctx,
		addons: addons,
		mgmt:   mgmt,
		runner: NewRunner(ctx, addons).WithClusterManager(mgmt),
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

func (k *K8S) CCM() *ccm.CCM {
	return k.addon(ccm.Name).(*ccm.CCM)
}

func (k *K8S) Validate() error {
	if err := k.mgmt.ValidateNodePatches(); err != nil {
		return fmt.Errorf("failed to validate node patches: %w", err)
	}
	return addons.Validate(k.addons)
}

func (k *K8S) Up(kubeconfig pulumi.AnyOutput, deps []pulumi.Resource) error {
	prov, err := k.Provider(kubeconfig, deps)
	if err != nil {
		return err
	}

	if err := k.mgmt.ManageNodes(prov); err != nil {
		return err
	}

	return k.runner.Run(prov)
}

func (k *K8S) addon(name string) addons.Addon {
	for _, addon := range k.addons {
		if addon.Name() == name {
			return addon
		}
	}

	return nil
}
