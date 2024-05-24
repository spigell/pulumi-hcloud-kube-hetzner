package k3supgrader

import (
	"strings"

	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	upgradev1 "github.com/spigell/pulumi-hcloud-kube-hetzner/crds/generated/rancher/upgrade/v1"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

const (
	ControlPlanNodesPlanName = "k3s-control-plane-nodes"
	AgentNodesPlanName       = "k3s-agent-nodes"
	channelAPIService        = "https://update.k3s.io/v1-release/channels"
)

var planEnabledNodeSelector = upgradev1.PlanSpecNodeSelectorMatchExpressionsArray{
	&upgradev1.PlanSpecNodeSelectorMatchExpressionsArgs{
		Key:      pulumi.String(ControlLabelKey),
		Operator: pulumi.String("Exists"),
	},
	&upgradev1.PlanSpecNodeSelectorMatchExpressionsArgs{
		Key:      pulumi.String(ControlLabelKey),
		Operator: pulumi.String("NotIn"),
		Values:   pulumi.StringArray{pulumi.String("false")},
	},
}

func (u *Upgrader) DeployPlans(ctx *program.Context, ns *corev1.Namespace, prov *kubernetes.Provider, deps pulumi.ResourceArrayOutput, nodes map[string]*manager.Node) error {
	plans := map[string]*upgradev1.PlanSpecArgs{
		ControlPlanNodesPlanName: {
			Concurrency:        pulumi.Int(1),
			ServiceAccountName: pulumi.String(u.serviceAccountName),
			Cordon:             pulumi.Bool(true),
			Upgrade: &upgradev1.PlanSpecUpgradeArgs{
				Image: pulumi.String("rancher/k3s-upgrade"),
			},
			NodeSelector: &upgradev1.PlanSpecNodeSelectorArgs{
				MatchExpressions: append(planEnabledNodeSelector,
					&upgradev1.PlanSpecNodeSelectorMatchExpressionsArgs{
						Key:      pulumi.String("node-role.kubernetes.io/master"),
						Operator: pulumi.String("Exists"),
					},
				),
			},
			Tolerations: getAllTolerationsFromNodes(nodes),
		},
		AgentNodesPlanName: {
			Concurrency:        pulumi.Int(1),
			ServiceAccountName: pulumi.String(u.serviceAccountName),
			Upgrade: &upgradev1.PlanSpecUpgradeArgs{
				Image: pulumi.String("rancher/k3s-upgrade"),
			},
			Drain: &upgradev1.PlanSpecDrainArgs{
				Force:                    pulumi.Bool(true),
				SkipWaitForDeleteTimeout: pulumi.Int(60),
			},
			NodeSelector: &upgradev1.PlanSpecNodeSelectorArgs{
				MatchExpressions: append(planEnabledNodeSelector,
					&upgradev1.PlanSpecNodeSelectorMatchExpressionsArgs{
						Key:      pulumi.String("node-role.kubernetes.io/master"),
						Operator: pulumi.String("DoesNotExist"),
					},
				),
			},
			Tolerations: getAllTolerationsFromNodes(nodes),
		},
	}

	for name, spec := range plans {
		spec = u.specifyVersionAndChannel(spec)
		if _, err := program.PulumiRun(ctx, upgradev1.NewPlan, name, &upgradev1.PlanArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(name),
				Namespace: ns.Metadata.Name(),
			},
			Spec: spec,
		},
			pulumi.Provider(prov),
			pulumi.DependsOnInputs(deps),
		); err != nil {
			return err
		}
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

func (u *Upgrader) specifyVersionAndChannel(spec *upgradev1.PlanSpecArgs) *upgradev1.PlanSpecArgs {
	if u.channel != "" {
		spec.Channel = pulumi.Sprintf("%s/%s", channelAPIService, u.channel)
	}

	if u.version != "" {
		spec.Version = pulumi.String(u.version)
	}

	return spec
}
