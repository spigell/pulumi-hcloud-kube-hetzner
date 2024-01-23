//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelsTaintsChange(t *testing.T) {
	targetLabelKey, targetLabelValue, desiredLabelValue := "example.io/test-label2", changeMe, "good"

	desiredTaint := "example.io/important-node=true:NoSchedule"

	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testNodeChangeLabelsTaints) {
		t.Skip()
	}

	out, err := i.Outputs()
	assert.NoError(t, err)

	kubeconfig, ok := out[phkh.KubeconfigKey].(string)
	require.True(t, ok)

	k8s, err := k8s.New(ctx, kubeconfig)
	require.NoError(t, err)

	old, err := k8s.Nodes()
	require.NoError(t, err)

	// Get first server node id
	nodeID, err := i.Stack.GetConfigWithOptions(ctx, "nodepools.servers[0].nodes[0].id", &auto.ConfigOptions{Path: true})
	require.NoError(t, err)

	for _, n := range old {
		if strings.HasSuffix(n.Name, nodeID.Value) {
			targetLabelValue, ok = n.Labels[targetLabelKey]
			assert.False(t, ok, fmt.Sprintf("label must not exist for %s", n.Name))
		}
	}
	assert.NotEqual(t, targetLabelValue, fmt.Sprintf("target label is not changed for node %s. Is node exist?", nodeID.Value))

	// Try to patch node labels and taints with new values
	i.Stack.SetConfigWithOptions(ctx, "nodepools.servers[0].nodes[0].k8s.node-label[0]", auto.ConfigValue{
		Value: fmt.Sprintf("%s=%s", targetLabelKey, desiredLabelValue),
	}, &auto.ConfigOptions{Path: true})

	i.Stack.SetConfigWithOptions(ctx, "nodepools.servers[0].nodes[0].k8s.node-taint[0]", auto.ConfigValue{
		Value: desiredTaint,
	}, &auto.ConfigOptions{Path: true})

	require.NoError(t, i.UpWithRetry())
	require.NoError(t, err)

	withLabelsTaints, err := k8s.Nodes()
	require.NoError(t, err)

	for _, n := range withLabelsTaints {
		if nodeID.Value == n.Name {
			require.Contains(t, n.Labels, targetLabelKey, fmt.Sprintf("no target label found for node %s", n.Name))
			require.Equal(t, desiredLabelValue, n.Labels[targetLabelKey])
			// It is enough to check only key existence for now
			exist := false
			for _, taint := range n.Spec.Taints {
				if taint.Key == strings.Split(desiredTaint, "=")[0] {
					exist = true
					break
				}
			}
			require.True(t, exist)
		}
	}

	i.Stack.RemoveConfigWithOptions(ctx, "nodepools.servers[0].nodes[0].k8s.node-label[0]", &auto.ConfigOptions{Path: true})
	i.Stack.RemoveConfigWithOptions(ctx, "nodepools.servers[0].nodes[0].k8s.node-taint[0]", &auto.ConfigOptions{Path: true})
	require.NoError(t, i.UpWithRetry())
	require.NoError(t, err)

	withoutLabelsTaints, err := k8s.Nodes()
	require.NoError(t, err)

	for _, n := range withoutLabelsTaints {
		if strings.HasSuffix(n.Name, nodeID.Value) {
			// TO DO: check that label is removed
			assert.NotContains(t, n.Labels, targetLabelKey, fmt.Sprintf("The label found for node %s", n.Name))
			// It is enough to check only key existence for now
			exist := false
			for _, taint := range n.Spec.Taints {
				if taint.Key == strings.Split(desiredTaint, "=")[0] {
					exist = true
					break
				}
			}
			assert.False(t, exist)
		}
	}
}
