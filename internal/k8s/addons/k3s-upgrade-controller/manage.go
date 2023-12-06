package upgrader

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
)

const (
	helmRepo  = "https://nimbolus.github.io/helm-charts"
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

	// Use Chart in sake of Transformations.
	deployed, err := helmv3.NewChart(ctx, name, helmv3.ChartArgs{
		Chart:     pulumi.String(helmChart),
		Namespace: ns.Metadata.Name().Elem(),
		Version:   pulumi.String(u.helm.Version),
		FetchArgs: &helmv3.FetchArgs{
			Repo: pulumi.String(helmRepo),
		},
		Values: pulumi.Map{
			"tolerations": pulumi.ToMapArray(manager.ComputeTolerationsFromNodes(nodes)),
		},
		Transformations: []yaml.Transformation{
			func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
				if state["kind"] == "Deployment" {
					spec := state["spec"].(map[string]interface{})
					podSpec := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})
					// There is only one container in pod spec.
					container := podSpec["containers"].([]interface{})[0].(map[string]interface{})
					volumeMounts := container["volumeMounts"].([]interface{})
					container["volumeMounts"] = append(volumeMounts, map[string]string{
						"name":      "ca-certificates",
						"mountPath": "/var/lib/ca-certificates",
					})

					volumes := podSpec["volumes"].([]interface{})
					podSpec["volumes"] = append(volumes, map[string]interface{}{
						"name": "ca-certificates",
						// This is a path to ca-certificates on host.
						// It is hardcoded in k3s.
						"hostPath": map[string]string{
							"path": "/var/lib/ca-certificates",
							"type": "Directory",
						},
					},
					)
				}
			},
		},
	},
		pulumi.Provider(prov),
		pulumi.DeleteBeforeReplace(true),
	)

	deps = append(deps, deployed)

	if err != nil {
		return fmt.Errorf("unable to create helm release: %w", err)
	}

	return u.DeployPlans(ctx, ns, prov, deps, nodes)
}
