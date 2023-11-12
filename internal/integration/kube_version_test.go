package integration

import (
	"context"
	"slices"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
)

func TestKubeVersion(t *testing.T) {
	name := testKubeVersion
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], name) {
		t.Skip()
	}

	out, err := i.Stack.Outputs(ctx)
	assert.NoError(t, err)

	kubeconfig, ok := out[phkh.KubeconfigKey].Value.(string)
	assert.True(t, ok)
	

	ver, err := k8s.KubeletVersion(kubeconfig, "test")
	assert.NoError(t, err)

	assert.Equal(t, ver, i.Example.Decoded.Defaults.Global.K3s.Version)
}
