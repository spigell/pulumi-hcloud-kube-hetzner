package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8S) Nodes() ([]v1.Node, error) {
	n, err := k.Client.CoreV1().Nodes().List(k.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return n.Items, nil
}
