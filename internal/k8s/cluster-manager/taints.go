package manager

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"k8s.io/apimachinery/pkg/api/errors"
	kubeApiMetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	corev1api "k8s.io/api/core/v1"
)

func (m *ClusterManager) ManageTaints(node *Node) error {
	// Create NodePatch
	taints, err := program.PulumiRun(m.ctx, corev1.NewNodePatch, fmt.Sprintf("taints-%s", node.ID), &corev1.NodePatchArgs{
		Metadata: &metav1.ObjectMetaPatchArgs{
			Name: pulumi.String(node.ID),
			Annotations: pulumi.StringMap{
				"pulumi.com/patchForce": pulumi.String("true"),
			},
		},
		Spec: &corev1.NodeSpecPatchArgs{
			// K3S or other controllers tries to take ownership of the node taints sometimes.
			// Trying to get all taints (including added by hands or other operators) and take ownership.
			Taints: m.client.ApplyT(
				func(cli interface{}) ([]corev1.TaintPatch, error) {
					additional := node.Taints
					clientSet := cli.(*kubernetes.Clientset)

					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					node, err := clientSet.CoreV1().Nodes().Get(ctx, node.ID, kubeApiMetav1.GetOptions{})
					if err != nil && !errors.IsNotFound(err) {
						return nil, err
					}

					keys := make([]string, 0)

					merged := append(
						toPatchTaintsFromTaintSlice(node.Spec.Taints),
						toPatchTaintsFromStringSlice(additional)...,
					)

					for _, t := range merged {
						keys = append(keys, *t.Key)
					}

					sort.Strings(keys)

					// Simple sort by key.
					sorted := make([]corev1.TaintPatch, 0)
					for k := range keys {
						for _, t := range merged {
							if *t.Key == keys[k] {
								sorted = append(sorted, t)
							}
						}
					}

					return slices.CompactFunc(sorted,
						func(k, j corev1.TaintPatch) bool {
							if *k.Key == *j.Key && *k.Effect == *j.Effect {
								return true
							}
							return false
						},
					), nil
				},
			).(corev1.TaintPatchArrayOutput),
		},
	},
		// Recreate resource on any changes to delete our old fieldManager.
		pulumi.ReplaceOnChanges([]string{"*"}),
		pulumi.DeleteBeforeReplace(true),
		pulumi.Provider(m.provider),
	)
	if err != nil {
		return err
	}

	m.resources = append(m.resources, taints)

	return nil
}

func toPatchTaintsFromTaintSlice(taints []corev1api.Taint) []corev1.TaintPatch {
	t := make([]corev1.TaintPatch, 0)

	for i := range taints {
		effect := string(taints[i].Effect)
		t = append(t, corev1.TaintPatch{
			Key:    &taints[i].Key,
			Value:  &taints[i].Value,
			Effect: &effect,
		})
	}

	return t
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
