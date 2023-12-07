//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	upgrader "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/k3s-upgrade-controller"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestK3SUpgradeControllerPlanValid(t *testing.T) {
	targetPlans := []string{upgrader.ControlPlanNodesPlanName, upgrader.AgentNodesPlanName}

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

	k8s, err := k8s.New(ctx, kubeconfig)
	require.NoError(t, err)

	for _, planName := range targetPlans {
		plan, err := k8s.K3SUpgradePlan(upgrader.Namespace, planName)
		require.NoError(t, err)

		fmt.Printf("%+v\n", plan)

		status, ok := plan["status"].(map[string]interface{})
		require.True(t, ok, fmt.Sprintf("plan status is not map[string]interface{}: %T", plan["status"]))

		conditions, ok := status["conditions"].([]map[string]string)
		require.True(t, ok)

		for _, condition := range conditions {
			assert.Equal(t, "Validated", condition["type"])
		}
	}
}
