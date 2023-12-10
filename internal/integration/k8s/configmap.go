package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMap returns configmap by name and namespace.
func (k *K8S) ConfigMap(ns, name string) (*v1.ConfigMap, error) {
	return k.Client.CoreV1().ConfigMaps(ns).Get(k.ctx, name, metav1.GetOptions{})
}
