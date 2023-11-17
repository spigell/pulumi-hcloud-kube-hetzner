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

	out, err := i.Stack.Outputs(ctx)
	assert.NoError(t, err)

	kubeconfig, ok := out[phkh.KubeconfigKey].Value.(string)
	assert.True(t, ok)

	k8s, err := k8s.New(ctx, kubeconfig)
	assert.NoError(t, err)

	nodes, err := k8s.Nodes()
	assert.NoError(t, err)

	for _, n := range nodes {
		externalIP := false
		ready := false
		for _, addr := range n.Status.Addresses {
			if addr.Type == "ExternalIP" {
				externalIP = true
				assert.NotEmpty(t, addr.Address)
			}
		}

		for _, c := range n.Status.Conditions {
			if c.Type == "Ready" {
				ready = true
			}
		}

		assert.True(t, externalIP, fmt.Sprintf("no externalIP found for node %s", n.Name))
		assert.True(t, ready, fmt.Sprintf("node %s is not ready", n.Name))
		assert.Contains(t, n.Labels, "node.kubernetes.io/instance-type", fmt.Sprintf("no instance-type label found for node %s", n.Name))
	}
}
