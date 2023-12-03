package k3s

import (
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
)

const (
	defaultServiceCIDR = "10.140.0.0/16"
	defaultClusterCIDR = "10.141.0.0/16"
)

// This is very opinionated values and it is based on my expirience with k3s.
var (
	defaultKubeControllerManagerArgs = map[string]string{
		// Increase time for a grace period for failed nodes.
		// With this increased value cluster discovers failed nodes longer.
		// K3s are mostly used in small environments with very tight amounts of resources.
		// So, it is better to wait a bit longer for a node to come back than to lose it.
		"node-monitor-grace-period": "2m",
	}
	defaultsKubeCloudControllerManagerArgs = map[string]string{
		// https://github.com/k3s-io/k3s/discussions/6452#discussioncomment-4080240
		// It can conflict with hetzner CCM
		"secure-port": "0",
	}
	defaultKubeAPIServerArgs = map[string]string{
		// If the node is down there is no need to wait more than 60s.
		"default-not-ready-toleration-seconds":   "60",
		"default-unreachable-toleration-seconds": "60",
	}
	defaultsKubeletArgs = map[string]map[string]string{
		variables.ServerRole: {
			// every 5s is too much for small clusters.
			"node-status-update-frequency": "20s",
			// We need to be sure that server has needed resources for k3s binary service.
			"system-reserved": "cpu=1,memory=1Gi",
		},
		variables.AgentRole: {
			"node-status-update-frequency": "20s",
			// Agent consumes less resources than server.
			"system-reserved": "cpu=100m,memory=100Mi",
		},
	}
	DefaultTaints = map[string][]string{
		variables.ServerRole: {
			// This taints are needed to prevent pods from being scheduled on the server node.
			// Used in situations when agent nodes exists.
			"CriticalAddonsOnly=true:NoExecute",
			"node-role.kubernetes.io/control-plane:NoSchedule",
		},
	}
)

func (k *K3sConfig) WithServerDefaults() *K3sConfig {
	k.WriteKubeconfigMode = "0644"
	k.TLSSanSecurity = true

	if k.ClusterCidr == "" {
		k.ClusterCidr = defaultClusterCIDR
	}

	if k.ServiceCidr == "" {
		k.ServiceCidr = defaultServiceCIDR
	}

	for _, key := range utils.SortedMapKeys(defaultKubeControllerManagerArgs) {
		value := defaultKubeControllerManagerArgs[key]
		if !containsKey(k.KubeControllerManagerArgs, key) {
			k.KubeControllerManagerArgs = append(k.KubeControllerManagerArgs,
				strings.Join([]string{key, value}, "="),
			)
		}
	}

	for _, key := range utils.SortedMapKeys(defaultsKubeCloudControllerManagerArgs) {
		value := defaultsKubeCloudControllerManagerArgs[key]
		if !containsKey(k.KubeCloudControllerManagerArgs, key) {
			k.KubeCloudControllerManagerArgs = append(k.KubeCloudControllerManagerArgs,
				strings.Join([]string{key, value}, "="),
			)
		}
	}

	for _, key := range utils.SortedMapKeys(defaultKubeAPIServerArgs) {
		value := defaultKubeAPIServerArgs[key]
		if !containsKey(k.KubeAPIServerArgs, key) {
			k.KubeAPIServerArgs = append(k.KubeAPIServerArgs,
				strings.Join([]string{key, value}, "="),
			)
		}
	}

	for _, key := range utils.SortedMapKeys(defaultsKubeletArgs[variables.ServerRole]) {
		value := defaultsKubeletArgs[variables.ServerRole][key]
		if !containsKey(k.KubeletArgs, key) {
			k.KubeletArgs = append(k.KubeletArgs,
				strings.Join([]string{key, value}, "="),
			)
		}
	}
	return k
}

// containsKey checks if a key exists in a slice.
func containsKey(slice []string, key string) bool {
	for _, s := range slice {
		if strings.Split(s, "=")[0] == key {
			return true
		}
	}
	return false
}
