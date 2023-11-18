//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

// TestKubeChangeEndpoint tests that changing endpoint type from public to internal and checking that kubeconfig changed.
func TestKubeChangeEndpoint(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), withPulumiDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeChangeEndpointType) {
		t.Skip()
	}

	old, err := i.Stack.Outputs(ctx)
	assert.NoError(t, err)

	publicKubeconfig, ok := old[phkh.KubeconfigKey].Value.(string)
	assert.True(t, ok)

	val, err := i.Stack.GetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", &auto.ConfigOptions{Path: true})
	require.NoError(t, err)
	require.Equal(t, variables.PublicCommunicationMethod.String(), val.Value)

	// Change to internal
	i.Stack.SetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", auto.ConfigValue{
		Value: variables.InternalCommunicationMethod.String(),
	},
		&auto.ConfigOptions{Path: true},
	)

	// Make UP for update kubeconfig output
	// Also allowed rules will be removed
	assert.NoError(t, i.UpWithRetry())
	assert.NoError(t, err)

	new, err := i.Stack.Outputs(ctx)
	assert.NoError(t, err)
	internalKubeconfig, ok := new[phkh.KubeconfigKey].Value.(string)
	assert.True(t, ok)

	// Change to wireguard
	i.Stack.SetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", auto.ConfigValue{
		Value: variables.WgCommunicationMethod.String(),
	},
		&auto.ConfigOptions{Path: true},
	)

	// Check that preview is ok and no changes
	_, err = i.Stack.Preview(ctx, optpreview.ExpectNoChanges())
	assert.NoError(t, err)

	// Change endpoint type back to public to not break deletion step.
	i.Stack.SetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", auto.ConfigValue{
		Value: variables.PublicCommunicationMethod.String(),
	},
		&auto.ConfigOptions{Path: true},
	)
	assert.NoError(t, i.UpWithRetry())

	// Check that kubeconfig changed
	assert.NotEqual(t, publicKubeconfig, internalKubeconfig)
}
