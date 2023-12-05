package manager

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
)

type ClusterManager struct {
	ctx *pulumi.Context

	nodes map[string]*Node
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
		ctx:   ctx,
		nodes: nodes,
	}
}

func (m *ClusterManager) ManageNodes(provider *kubernetes.Provider) error {
	for _, node := range m.nodes {
		// Create NodePatch
		_, err := corev1.NewNodePatch(m.ctx, fmt.Sprintf("taints-%s", node.ID), &corev1.NodePatchArgs{
			Metadata: &metav1.ObjectMetaPatchArgs{
				Name: pulumi.String(node.ID),
				Annotations: pulumi.StringMap{
					"pulumi.com/patchForce": pulumi.String("true"),
				},
			},
			Spec: &corev1.NodeSpecPatchArgs{
				Taints: toTaints(node.Taints),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		_, err = corev1.NewNodePatch(m.ctx, fmt.Sprintf("labels-%s", node.ID), &corev1.NodePatchArgs{
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
	}

	return nil
}

func toTaints(taints []string) corev1.TaintPatchArray {
	var t corev1.TaintPatchArray

	for _, taint := range taints {
		keyValue, effect := strings.Split(taint, ":")[0], strings.Split(taint, ":")[1]

		// Value is optional.
		key, value := strings.Split(keyValue, "=")[0], ""
		if l := len(strings.Split(keyValue, "=")); l == 2 {
			value = strings.Split(keyValue, "=")[1]
		}

		t = append(t, corev1.TaintPatchArgs{
			Key:    pulumi.String(key),
			Value:  pulumi.String(value),
			Effect: pulumi.String(effect),
		})
	}

	return t
}
