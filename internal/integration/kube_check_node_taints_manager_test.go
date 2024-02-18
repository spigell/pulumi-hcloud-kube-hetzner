//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
)

func TestKubeTaintsManager(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline())
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeCheckTaintsManager) {
		t.Skip()
	}

	if slices.Contains(TestsByExampleName[i.Example.Name], testNodeChangeLabelsTaints) {
		t.Error("conflict: testNodeChangeLabelsTaints must not be enabled for this example")
	}

	out, err := i.Outputs()
	assert.NoError(t, err)

	kubeconfig, ok := out[phkh.KubeconfigKey].(string)
	assert.True(t, ok)

	k8s, err := k8s.New(ctx, kubeconfig)
	assert.NoError(t, err)

	nodes, err := k8s.Nodes()
	assert.NoError(t, err)

	for _, n := range nodes {
		if len(n.Spec.Taints) > 0 {
			for _, f := range n.ManagedFields {
				if strings.Contains(f.FieldsV1.String(), "f:taints") {
					assert.True(t, strings.HasPrefix(f.Manager, "pulumi-kubernetes"), fmt.Sprintf("field manager is not pulumi as expected: %s", f.Manager))
				}
			}
		}
	}
}
