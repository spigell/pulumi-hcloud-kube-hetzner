package upgrader

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
)

const (
	helmRepo         = "https://nimbolus.github.io/helm-charts"
	helmChart = "system-upgrade-controller"
	namespace = "system-upgrade"
)

func (u *Upgrader) Manage(ctx *pulumi.Context, prov *kubernetes.Provider, nodes map[string]*manager.Node) error {
	deps := make([]pulumi.Resource, 0)
	// Create ns
	ns, err := corev1.NewNamespace(ctx, namespace, &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(namespace),
		},
	}, pulumi.Provider(prov))

	if err != nil {
		return fmt.Errorf("unable to create namespace: %w", err)
	}

	deployed, err := helmv3.NewRelease(ctx, name, &helmv3.ReleaseArgs{
		Chart:     pulumi.String(helmChart),
		Namespace: ns.Metadata.Name(),
		Version:   pulumi.String(u.helm.Version),
		Name:      pulumi.String(name),
		RepositoryOpts: helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(helmRepo),
		},
		Values: pulumi.Map{
			"tolerations": getAllTolerationsFromNodes(nodes),
		}},
		pulumi.Provider(prov),
		pulumi.DeleteBeforeReplace(true),
	)

	deps = append(deps, deployed)

	if err != nil {
		return fmt.Errorf("unable to create helm release: %w", err)
	}

	return u.DeployPlans(ctx, ns, prov, deps, nodes)
}
