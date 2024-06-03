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
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelsTaintsManagement(t *testing.T) {
	targetLabelKey, targetLabelValue, desiredLabelValue := "example.io/test-label2", changeMe, "good"

	desiredTaint := "example.io/important-node=true:NoSchedule"

	taintConfigKey := "cluster.nodepools.servers[0].nodes[0].k8s.node-taint.taints[0]"
	labelConfigKey := "cluster.nodepools.servers[0].nodes[0].k8s.node-label[0]"

	// t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), withPulumiDeadline)
	defer cancel()

	i, err := New(ctx)
	require.NoError(t, err)

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
	nodeID, err := i.Stack.GetConfigWithOptions(ctx, "cluster.nodepools.servers[0].nodes[0].node-id", &auto.ConfigOptions{Path: true})
	require.NoError(t, err)

	for _, n := range old {
		if strings.HasSuffix(n.Name, nodeID.Value) {
			targetLabelValue, ok = n.Labels[targetLabelKey]
			assert.False(t, ok, fmt.Sprintf("label must not exist for %s", n.Name))
		}
	}
	assert.NotEqual(t, targetLabelValue, fmt.Sprintf("target label is not changed for node %s. Is node exist?", nodeID.Value))

	// Try to patch node labels and taints with new values
	i.Stack.SetConfigWithOptions(ctx, labelConfigKey, auto.ConfigValue{
		Value: fmt.Sprintf("%s=%s", targetLabelKey, desiredLabelValue),
	}, &auto.ConfigOptions{Path: true})
	// Taint management must be enabled first
	i.Stack.SetConfigWithOptions(ctx, "cluster.nodepools.servers[0].nodes[0].k8s.node-taint.enabled", auto.ConfigValue{
		Value: "true",
	}, &auto.ConfigOptions{Path: true})

	i.Stack.SetConfigWithOptions(ctx, taintConfigKey, auto.ConfigValue{
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

			// Len of taints must be equal to 1. No default or marsian taints.
			assert.Len(t, n.Spec.Taints, 1, fmt.Sprintf("node must has only 1 taint. node taints: %+v", n.Name))

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

	i.Stack.RemoveConfigWithOptions(ctx, labelConfigKey, &auto.ConfigOptions{Path: true})
	i.Stack.RemoveConfigWithOptions(ctx, taintConfigKey, &auto.ConfigOptions{Path: true})
	require.NoError(t, i.UpWithRetry())
	require.NoError(t, err)

	withoutLabelsTaints, err := k8s.Nodes()
	require.NoError(t, err)

	for _, n := range withoutLabelsTaints {
		if strings.HasSuffix(n.Name, nodeID.Value) {
			// TO DO: check that label is removed
			assert.NotContains(t, n.Labels, targetLabelKey, fmt.Sprintf("The label found for node %s", n.Name))

			// Default taints must be added.
			// Disable default taints must be set to false.
			if len(i.Example.Decoded.Nodepools.Agents) > 0 {
				for _, ta := range manager.DefaultTaints[variables.ServerRole] {
					d := strings.Split(ta, "=")[0]
					exist := false
					for _, taint := range n.Spec.Taints {
						if taint.Key == strings.Split(d, "=")[0] {
							exist = true
							break
						}
					}
					require.True(t, exist)
				}
			}
			// Even if we delete taint in configuration it should be present on node.
			// It is enough to check only key existence for now.
			exist := false
			for _, taint := range n.Spec.Taints {
				if taint.Key == strings.Split(desiredTaint, "=")[0] {
					exist = true
					break
				}
			}
			assert.True(t, exist)
		}
	}
}
