//go:build connectivity && linux
// +build connectivity,linux

// This package contains all the integration suites for the pulumi program.
// The integration suites are used to test the pulumi program against a real deployed pulumi stack.
package integration

import (
	"context"
	"os/exec"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

var (
	up *wireguard.Wireguard
)

func init() {
	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	out, err := i.Stack.Outputs(ctx)
	if err != nil {
		panic(err)
	}

	config, ok := out[phkh.WGMasterConKey].Value.(string)
	if !ok {
		panic("failed to get wireguard master connection string")
	}

	up, err = wireguard.Up(config)
	if err != nil {
		panic(err)
	}
}

func TestWireguradConnectivity(t *testing.T) {
	t.Parallel()

	defer up.Close()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testWGConnectivity) {
		t.Skip()
	}

	out, err := i.Stack.Outputs(ctx)
	assert.NoError(t, err)

	info, ok := out[phkh.WGInfoKey]
	assert.True(t, ok)

	for _, peer := range info.Value.(map[string]interface{}) {
		n := peer.(map[string]interface{})

		ip, ok := n["ip"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, ip)

		ping := exec.Command("ping", "-c", "2", ip).Run()
		assert.NoError(t, ping)
	}
}
