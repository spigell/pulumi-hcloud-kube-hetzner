//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
)

func TestKubeVersion(t *testing.T) {
	name := testKubeVersion
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline())
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], name) {
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
		for k, v := range i.Example.UniqConfigsByNodes() {
			if strings.HasSuffix(n.Name, k) && v.K3s != nil && v.K3s.Version != "" {
				assert.Equal(t, n.Status.NodeInfo.KubeletVersion, v.K3s.Version)
				continue
			}
		}

		// If k3s-upgrade-controller is disabled, then kubelet version must be equal to k3s version of node specified in config
		if !i.Example.Decoded.K8S.Addons.K3SSystemUpgrader.Enabled {
			assert.Equal(t, n.Status.NodeInfo.KubeletVersion, i.Example.Decoded.Defaults.Global.K3s.Version)
		}
	}
}
