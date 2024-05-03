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
	kubeApiMetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"

	// cloudproviderapi "k8s.io/cloud-provider/api".
	corev1api "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// whitelistedTaints is a list of taints that should not be treated as user-defined.
var whitelistedTaints = []string{
	// Next taints are used by kubernetes itself and can be added by controllers
	corev1api.TaintNodeUnreachable,
	corev1api.TaintNodeNetworkUnavailable,
	corev1api.TaintNodeDiskPressure,
	corev1api.TaintNodeMemoryPressure,
	corev1api.TaintNodePIDPressure,
	corev1api.TaintNodeNotReady,
}

func (m *ClusterManager) ManageTaints(node *Node) error {
	// Create NodePatch
	taints, err := corev1.NewNodePatch(m.ctx.Context(), fmt.Sprintf("taints-%s", node.ID), &corev1.NodePatchArgs{
		Metadata: &metav1.ObjectMetaPatchArgs{
			Name: pulumi.String(node.ID),
			Annotations: pulumi.StringMap{
				"pulumi.com/patchForce": pulumi.String("true"),
			},
		},
		Spec: &corev1.NodeSpecPatchArgs{
			Taints: m.kubeconfig.ApplyT(
				func(cfg interface{}) []corev1.TaintPatch {
					kubeconfig := cfg.(*api.Config)
					additional := node.Taints
					// current := args[0].([]corev1.Taint)

					d, _ := clientcmd.Write(*kubeconfig)
					restConfig, err := clientcmd.RESTConfigFromKubeConfig(d)
					if err != nil {
						fmt.Errorf("Error creating Kubernetes REST config: %s", err)
					}
					clientSet, err := kubernetes.NewForConfig(restConfig)
					if err != nil {
						fmt.Errorf("error creating the clientset: %v", err)
					}

					current := make([]corev1api.Taint, 0)

					// K3S tries to take ownership of the node taints.
					// K3S does it after the node is created, so we need to wait for it.
					// We need to wait for the node to be initialized, wait little more and then apply our taints with pulumi Manager.
					// Very naive way to do it, but it should work for now.
					for _, i := range []int64{1, 2, 3, 5, 10, 20, 30} {
						node, err := clientSet.CoreV1().Nodes().Get(context.Background(), node.ID, kubeApiMetav1.GetOptions{})
						if err != nil {
							fmt.Errorf("failed to get node: %s", err)
						}
						for _, f := range node.ManagedFields {
							if strings.Contains(f.FieldsV1.String(), "f:taints") && strings.HasPrefix(f.Manager, "k3s") {
								current = append(current, node.Spec.Taints...)
							}
						}
						m.ctx.Context().Log.Debug(fmt.Sprintf("taints: sleep for %d...", i), nil)
						time.Sleep(time.Duration(i) * time.Second)
					}

					keys := make([]string, 0)

					merged := append(
						toPatchTaintsFromTaintSlice(removeMartianTaints(current)),
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
					)
				},
			).(corev1.TaintPatchArrayOutput),
		},
	}, append(m.ctx.Options(), pulumi.Provider(m.provider))...)
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

func removeMartianTaints(taints []corev1api.Taint) []corev1api.Taint {
	t := make([]corev1api.Taint, 0)

	for _, taint := range taints {
		if slices.Contains(whitelistedTaints, taint.Key) {
			t = append(t, taint)
		}
	}

	return t
}
