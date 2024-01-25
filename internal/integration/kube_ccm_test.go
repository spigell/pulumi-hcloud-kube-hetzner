//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
)

func TestKubeHcloudCCM(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeHetznerCCM) {
		t.Skip()
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
		externalIP := false
		for _, addr := range n.Status.Addresses {
			if addr.Type == "ExternalIP" {
				externalIP = true
				assert.NotEmpty(t, addr.Address)
			}
		}

		assert.True(t, externalIP, fmt.Sprintf("no externalIP found for node %s", n.Name))
		assert.Contains(t, n.Labels, "node.kubernetes.io/instance-type", fmt.Sprintf("no instance-type label found for node %s", n.Name))
		assert.NotContains(t, n.Spec.Taints, "node.cloudprovider.kubernetes.io/uninitialized", fmt.Sprintf("uninitialized taints found for node %s", n.Name))
	}
}
