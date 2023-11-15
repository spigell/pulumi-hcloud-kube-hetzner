//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

// TestKubeChangeEndpoint tests that changing endpoint type from wireguard to internal and checking that kubeconfig changed.
func TestKubeChangeEndpoint(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), withPulumiDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeChangeEndpointType) {
		t.Skip()
	}

	out, err := i.Stack.Outputs(ctx)
	assert.NoError(t, err)

	old, ok := out[phkh.KubeconfigKey].Value.(string)
	assert.True(t, ok)

	val, err := i.Stack.GetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", &auto.ConfigOptions{Path: true})
	assert.NoError(t, err)
	assert.Equal(t, variables.WgCommunicationMethod, val.Value)

	i.Stack.SetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", auto.ConfigValue{
		Value: variables.InternalCommunicationMethod,
	},
		&auto.ConfigOptions{Path: true},
	)

	_, err = i.Stack.Preview(ctx, optpreview.ExpectNoChanges())
	assert.NoError(t, err)

	// Make UP for update kubeconfig output
	assert.NoError(t, i.UpWithRetry())
	assert.NoError(t, err)

	// Change endpoint type back to public to not break deletion step.
	i.Stack.SetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", auto.ConfigValue{
		Value: variables.PublicCommunicationMethod,
	},
		&auto.ConfigOptions{Path: true},
	)

	assert.NoError(t, i.UpWithRetry())

	new, ok := out[phkh.KubeconfigKey].Value.(string)
	assert.True(t, ok)

	// Check that kubeconfig changed
	assert.NotEqual(t, old, new)
}
