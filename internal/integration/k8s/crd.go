package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func (k *K8S) K3SUpgradePlan(ns, name string) (map[string]interface{}, error) {
	cliSet := dynamic.New(k.Client.RESTClient())

	plan, err := cliSet.Resource(schema.GroupVersionResource{
		Group:    "upgrade.cattle.io",
		Version:  "v1",
		Resource: "plans",
	}).Namespace(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get plan for %s: %w", name, err)
	}

	return plan.Object, nil
}
