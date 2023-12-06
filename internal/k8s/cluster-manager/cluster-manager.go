package manager

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
)

type ClusterManager struct {
	ctx *pulumi.Context

	nodes     map[string]*Node
	resources []pulumi.Resource
}

// Node is a representation of kubernetes node.
// Used for keeping node taints and labels updated.
type Node struct {
	ID     string
	Taints []string
	Labels []string
}

func New(ctx *pulumi.Context, nodes map[string]*Node) *ClusterManager {
	return &ClusterManager{
		ctx:       ctx,
		nodes:     nodes,
		resources: make([]pulumi.Resource, 0),
	}
}

func (m *ClusterManager) Nodes() map[string]*Node {
	return m.nodes
}

func (m *ClusterManager) Resources() []pulumi.Resource {
	return m.resources
}

func (m *ClusterManager) ManageNodes(provider *kubernetes.Provider) error {
	for _, node := range m.nodes {
		existed, err := corev1.GetNode(m.ctx, node.ID, pulumi.ID(node.ID), nil, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		// Create NodePatch
		taints, err := corev1.NewNodePatch(m.ctx, fmt.Sprintf("taints-%s", node.ID), &corev1.NodePatchArgs{
			Metadata: &metav1.ObjectMetaPatchArgs{
				Name: pulumi.String(node.ID),
				Annotations: pulumi.StringMap{
					"pulumi.com/patchForce": pulumi.String("true"),
				},
			},
			Spec: &corev1.NodeSpecPatchArgs{
				Taints: pulumi.All(existed.Spec.Taints(), node.Taints).ApplyT(
					func(args []interface{}) []corev1.TaintPatch {
						current := args[0].([]corev1.Taint)
						additional := args[1].([]string)

						return slices.CompactFunc(
							append(toPatchTaintsFromTaintSlice(current),
								toPatchTaintsFromStringSlice(additional)...,
							),
							func(k, j corev1.TaintPatch) bool {
								return *k.Key == *j.Key && *k.Effect == *j.Effect
							},
						)
					},
				).(corev1.TaintPatchArrayOutput),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		m.resources = append(m.resources, taints)

		labels, err := corev1.NewNodePatch(m.ctx, fmt.Sprintf("labels-%s", node.ID), &corev1.NodePatchArgs{
			Metadata: &metav1.ObjectMetaPatchArgs{
				Name: pulumi.String(node.ID),
				Annotations: pulumi.StringMap{
					"pulumi.com/patchForce": pulumi.String("true"),
				},
				Labels: utils.ToPulumiMap(node.Labels, "="),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		m.resources = append(m.resources, labels)
	}

	return nil
}

func toPatchTaintsFromStringSlice(taints []string) []corev1.TaintPatch {
	t := make([]corev1.TaintPatch, 0)

	for _, taint := range taints {
		keyValue, effect := strings.Split(taint, ":")[0], strings.Split(taint, ":")[1]

		// Value is optional.
		key, value := strings.Split(keyValue, "=")[0], ""
		if l := len(strings.Split(keyValue, "=")); l == 2 {
			value = strings.Split(keyValue, "=")[1]
		}

		t = append(t, corev1.TaintPatch{
			Key:    &key,
			Value:  &value,
			Effect: &effect,
		})
	}

	return t
}

func ComputeTolerationsFromNodes(nodes map[string]*Node) []map[string]interface{} {
	tolerations := make([]map[string]interface{}, 0)
	for _, node := range nodes {
		for _, taint := range node.Taints {
			keyValue, effect := strings.Split(taint, ":")[0], strings.Split(taint, ":")[1]

			// Value is optional.
			key, value := strings.Split(keyValue, "=")[0], ""
			if l := len(strings.Split(keyValue, "=")); l == 2 {
				value = strings.Split(keyValue, "=")[1]
			}

			tolerations = append(tolerations, map[string]interface{}{
				"key":    pulumi.String(key),
				"value":  pulumi.String(value),
				"effect": pulumi.String(effect),
			})
		}
	}

	return tolerations
}

func toPatchTaintsFromTaintSlice(taints []corev1.Taint) []corev1.TaintPatch {
	t := make([]corev1.TaintPatch, 0)

	for i := range taints {
		t = append(t, corev1.TaintPatch{
			Key:    &taints[i].Key,
			Value:  taints[i].Value,
			Effect: &taints[i].Effect,
		})
	}

	return t
}
