package manager

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
)

type ClusterManager struct {
	ctx        *program.Context
	provider   *kubernetes.Provider
	kubeconfig pulumi.AnyOutput

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

func New(ctx *program.Context, nodes map[string]*Node) *ClusterManager {
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

func (m *ClusterManager) Up(kubeconfig pulumi.AnyOutput, provider *kubernetes.Provider) error {
	m.provider = provider
	m.kubeconfig = kubeconfig

	for _, node := range m.nodes {
		if len(node.Taints) > 0 {
			if err := m.ManageTaints(node); err != nil {
				return err
			}
		}

		labels, err := corev1.NewNodePatch(m.ctx.Context(), fmt.Sprintf("labels-%s", node.ID), &corev1.NodePatchArgs{
			Metadata: &metav1.ObjectMetaPatchArgs{
				Name: pulumi.String(node.ID),
				Annotations: pulumi.StringMap{
					"pulumi.com/patchForce": pulumi.String("true"),
				},
				Labels: utils.ToPulumiMap(node.Labels, "="),
			},
		}, append(m.ctx.Options(), pulumi.Provider(provider))...)
		if err != nil {
			return err
		}

		m.resources = append(m.resources, labels)
	}

	return nil
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
