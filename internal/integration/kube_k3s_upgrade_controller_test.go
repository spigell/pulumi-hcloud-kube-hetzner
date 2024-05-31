//go:build kubernetes
// +build kubernetes

package integration

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/k8s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/k3supgrader"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestK3SUpgradeControllerPlan(t *testing.T) {
	targetPlans := []string{k3supgrader.ControlPlanNodesPlanName, k3supgrader.AgentNodesPlanName}

	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline())
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeK3SUpgradeControllerPlan) {
		t.Skip()
	}

	out, err := i.Outputs()
	assert.NoError(t, err)

	kubeconfig, ok := out[phkh.KubeconfigKey].(string)
	require.True(t, ok)

	k8s, err := k8s.New(ctx, kubeconfig)
	require.NoError(t, err)

	for _, planName := range targetPlans {
		plan, err := k8s.K3SUpgradePlan(k3supgrader.Namespace, planName)
		require.NoError(t, err)

		status, ok := plan["status"].(map[string]interface{})
		require.True(t, ok, fmt.Sprintf("plan status is not map[string]interface{}: %T", plan["status"]))

		conditions, ok := status["conditions"].([]interface{})
		require.True(t, ok, fmt.Sprintf("conditions is not []interface{}, but: %T", status["conditions"]))

		validated := false
		resolved := false
		for _, condition := range conditions {
			c, ok := condition.(map[string]interface{})
			require.True(t, ok, fmt.Sprintf("condition is not map[string]interface{}: %T", condition))

			reason, ok := c["reason"].(string)
			require.True(t, ok, fmt.Sprintf("Reason is not string: %T", c["reason"]))

			if reason == "PlanIsValid" {
				validated = true
				assert.Equal(t, "Validated", c["type"].(string))
				assert.Equal(t, "True", c["status"].(string))
			}

			if reason == "Channel" {
				resolved = true
				assert.Equal(t, "LatestResolved", c["type"].(string))
				assert.Equal(t, "True", c["status"].(string))
			}

			if reason == "Version" {
				resolved = true
				assert.Equal(t, "LatestResolved", c["type"].(string))
				assert.Equal(t, "True", c["status"].(string))
			}
		}
		assert.True(t, validated, "Plan is not validated")
		assert.True(t, resolved, "Channel or Version are not resolved")
	}
}

func TestK3SUpgradeControllerConfigEnv(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline())
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testKubeK3SUpgradeControllerConfigEnv) {
		t.Skip()
	}

	out, err := i.Outputs()
	require.NoError(t, err)

	kubeconfig, ok := out[phkh.KubeconfigKey].(string)
	require.True(t, ok)

	k8s, err := k8s.New(ctx, kubeconfig)
	require.NoError(t, err)

	cms, err := k8s.ConfigMaps(k3supgrader.Namespace)
	require.NoError(t, err)
	// How the name of Config Name generated?
	cm, err := k8s.ConfigMap(k3supgrader.Namespace, "test-k3s-upgrade-controller-system-upgrade-controller-env")
	require.NoError(t, err, fmt.Sprintf("error while getting config map: %s. All config maps are %+v", err.Error(), cms))

	require.True(t, ok)
	for _, env := range i.Example.Decoded.K8S.Addons.K3SSystemUpgrader.ConfigEnv {
		k, v := strings.Split(env, "=")[0], strings.Split(env, "=")[1]
		assert.Equal(t, v, cm.Data[k])
	}
}
