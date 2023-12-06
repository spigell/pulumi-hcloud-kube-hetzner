//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"slices"
	"testing"


	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestK3SUpgradeControllerPlanValid(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeK3SUpgradeControllerPlanValid) {
		t.Skip()
	}

	out, err := i.Stack.Outputs(ctx)
	assert.NoError(t, err)

	kubeconfig, ok := out[phkh.KubeconfigKey].Value.(string)
	require.True(t, ok)

	_, err = k8s.New(ctx, kubeconfig)
	require.NoError(t, err)

}
