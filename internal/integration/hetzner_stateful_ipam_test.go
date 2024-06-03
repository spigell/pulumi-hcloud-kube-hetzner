//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHetznerDeleteNode(t *testing.T) {
	// t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), withPulumiDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testHetznerNodeManagement) {
		t.Skip()
	}

	require.Lenf(t, i.Example.Decoded.Nodepools.Agents, 2, "expected 2 nodepool got %d",
		len(i.Example.Decoded.Nodepools.Agents),
	)

	require.Lenf(t, i.Example.Decoded.Nodepools.Agents[0].Nodes, 2, "expected 2 nodes in 1st nodepoool, got %d",
		len(i.Example.Decoded.Nodepools.Agents[0].Nodes),
	)

	out, err := i.Outputs()
	assert.NoError(t, err)

	old, ok := out[phkh.HetznerServersKey].([]interface{})
	assert.True(t, ok, "expected []interface{} got %T", out[phkh.HetznerServersKey])

	require.NoError(t,
		i.Stack.RemoveConfigWithOptions(ctx, "cluster.nodepools.agents[0].nodes[0]",
			&auto.ConfigOptions{Path: true},
		),
	)

	require.NoError(t, i.UpWithRetry())

	out, err = i.Outputs()
	assert.NoError(t, err)

	new, ok := out[phkh.HetznerServersKey].([]interface{})
	assert.True(t, ok, "expected []interface{} got %T", out[phkh.HetznerServersKey])

	// Check if new is shorter than old by 1. It means node is deleted.
	require.Len(t, new, len(old)-1)

	checkIPS(t, new, old)

	// Remove entire nodepool
	require.NoError(t,
		i.Stack.RemoveConfigWithOptions(ctx, "cluster.nodepools.agents[0]",
			&auto.ConfigOptions{Path: true},
		),
	)
	require.NoError(t, i.UpWithRetry())

	out, err = i.Outputs()
	assert.NoError(t, err)

	new, ok = out[phkh.HetznerServersKey].([]interface{})
	assert.True(t, ok, "expected []interface{} got %T", out[phkh.HetznerServersKey])

	checkIPS(t, new, old)
}

func checkIPS(t *testing.T, new, old []interface{}) {
	for _, node := range new {
		ip, ok := node.(map[string]interface{})[phkh.ServerInternalIPKey].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, ip)

		name, ok := node.(map[string]interface{})[phkh.ServerNameKey].(string)
		assert.True(t, ok)

		for _, n := range old {
			name2, ok := n.(map[string]interface{})[phkh.ServerNameKey].(string)
			assert.True(t, ok)

			ip2, ok := node.(map[string]interface{})[phkh.ServerInternalIPKey].(string)
			assert.True(t, ok)
			assert.NotEmpty(t, ip)

			if name == name2 {
				fmt.Printf("[%s] Names are equals! new ip:%s, old ip:%s\n", name, ip, ip2)
				assert.Equal(t, ip, ip2)
			}
		}
	}
}
