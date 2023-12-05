package upgrader

import (
	"strings"

	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	upgradev1 "github.com/spigell/pulumi-hcloud-kube-hetzner/crds/generated/rancher/upgrade/v1"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
)

const (
	channelApiService = "https://update.k3s.io/v1-release/channels"
)

func (u *Upgrader) DeployPlans(ctx *pulumi.Context, ns *corev1.Namespace, prov *kubernetes.Provider, deps []pulumi.Resource, nodes map[string]*manager.Node) error {
	enabledNodeSelector := upgradev1.PlanSpecNodeSelectorMatchExpressionsArray{
		&upgradev1.PlanSpecNodeSelectorMatchExpressionsArgs{
			Key: pulumi.String(ControlLabelKey),
			Operator: pulumi.String("Exists"),
		},
		&upgradev1.PlanSpecNodeSelectorMatchExpressionsArgs{
			Key: pulumi.String(ControlLabelKey),
			Operator: pulumi.String("NotIn"),
			Values: pulumi.StringArray{pulumi.String("false")},
		},
	}
	controlPlaneSpec := &upgradev1.PlanSpecArgs{
		Concurrency: pulumi.Int(1),
		ServiceAccountName: pulumi.String(u.serviceAccountName),
		Cordon: pulumi.Bool(true),
		Upgrade: &upgradev1.PlanSpecUpgradeArgs{
			Image: pulumi.String("rancher/k3s-upgrade"),
		},
		NodeSelector: &upgradev1.PlanSpecNodeSelectorArgs{
			MatchExpressions: append(enabledNodeSelector,
				&upgradev1.PlanSpecNodeSelectorMatchExpressionsArgs{
					Key: pulumi.String("node-role.kubernetes.io/master"),
					Operator: pulumi.String("Exists"),
				},
			),
		},
		Tolerations: getAllTolerationsFromNodes(nodes),
	}

	controlPlaneSpec.Channel = pulumi.Sprintf("%s/%s", channelApiService, u.channel)

	if u.version != "" {
		controlPlaneSpec.Channel = pulumi.String("")
		controlPlaneSpec.Version = pulumi.String(u.version)
	}

	if _, err := upgradev1.NewPlan(ctx, "k3s-control-plane-nodes", &upgradev1.PlanArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("k3s-control-plane-nodes"),
			Namespace: ns.Metadata.Name(),
		},
		Spec: controlPlaneSpec,
	}, pulumi.Provider(prov), pulumi.DependsOn(deps)); err != nil {
		return err
	}

	return nil
}

func getAllTolerationsFromNodes(nodes map[string]*manager.Node) upgradev1.PlanSpecTolerationsArray {
	var tolerations upgradev1.PlanSpecTolerationsArray

	for _, node := range nodes {
		for _, taint := range node.Taints {
			keyValue, effect := strings.Split(taint, ":")[0], strings.Split(taint, ":")[1]

			// Value is optional.
			key, value := strings.Split(keyValue, "=")[0], ""
			if l := len(strings.Split(keyValue, "=")); l == 2 {
				value = strings.Split(keyValue, "=")[1]
			}

			tolerations = append(tolerations, &upgradev1.PlanSpecTolerationsArgs{
				Key:    pulumi.String(key),
				Value:  pulumi.String(value),
				Effect: pulumi.String(effect),
			})

		}
	}

	return tolerations
}