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
)

func TestKubeChangeEndpoint(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeChangeEndpointType) {
		t.Skip()
	}

	i.Stack.SetConfigWithOptions(ctx, "k8s.kube-api-endpoint.type", auto.ConfigValue{
		Value: variables.WgCommunicationMethod,
	},
		&auto.ConfigOptions{Path: true},
	)

	_, err := i.Stack.Preview(ctx, optpreview.ExpectNoChanges())

	assert.NoError(t, err)
}
