package k3supgrader

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
)

const (
	helmRepo  = "https://nimbolus.github.io/helm-charts"
	helmChart = "system-upgrade-controller"
	Namespace = "system-upgrade"
)

func (u *Upgrader) Manage(ctx *program.Context, prov *kubernetes.Provider, mgmt *manager.ClusterManager) error {
	if u.helm.ValuesFiles() != nil {
		return fmt.Errorf("values-files is not supported for %s", u.Name())
	}

	// Create ns
	ns, err := corev1.NewNamespace(ctx.Context(), Namespace, &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(Namespace),
		},
	}, append(ctx.Options(), pulumi.Provider(prov))...)
	if err != nil {
		return fmt.Errorf("unable to create namespace: %w", err)
	}

	// Use Chart in sake of Transformations.
	deployed, err := helmv3.NewChart(ctx.Context(), Name, helmv3.ChartArgs{
		Chart:     pulumi.String(helmChart),
		Namespace: ns.Metadata.Name().Elem(),
		Version:   pulumi.String(u.helm.Version),
		FetchArgs: &helmv3.FetchArgs{
			Repo: pulumi.String(helmRepo),
		},
		Values: pulumi.Map{
			"tolerations": pulumi.ToMapArray(manager.ComputeTolerationsFromNodes(mgmt.Nodes())),
			"configEnv":   utils.ToPulumiMap(u.configEnv, "="),
		},
		Transformations: []yaml.Transformation{
			func(state map[string]interface{}, _ ...pulumi.ResourceOption) {
				if state["kind"] == "Deployment" {
					// Deleting taints via in underlayed manager can lead to infinity loop.
					// Skip waiting for the deployment for now.
					metadata := state["metadata"].(map[string]interface{})
					annotations, ok := metadata["annotations"].(map[string]interface{})
					if !ok {
						annotations = make(map[string]interface{})
						metadata["annotations"] = annotations
					}
					annotations["pulumi.com/skipAwait"] = "true"

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
		append(ctx.Options(),
			pulumi.Provider(prov),
			pulumi.DeleteBeforeReplace(true),
		)...)
	if err != nil {
		return fmt.Errorf("unable to create helm release: %w", err)
	}

	return u.DeployPlans(ctx, ns, prov, deployed.Ready, mgmt.Nodes())
}
