//go:build connectivity
// +build connectivity

// This package contains all the integration suites for the pulumi program.
// The integration suites are used to test the pulumi program against a real deployed pulumi stack.
package integration

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/integration/ssh"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

const ()

func TestSSHConnectivity(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], testSSHConnectivity) {
		t.Skip()
	}

	out, err := i.Outputs()
	assert.NoError(t, err)

	privatekey, ok := out[phkh.PrivatekeyKey].(string)
	assert.True(t, ok)

	nodes, ok := out[phkh.HetznerServersKey].([]interface{})
	assert.True(t, ok, "expected []interface{} got %T", out[phkh.HetznerServersKey])

	for _, node := range nodes {
		ip, ok := node.(map[string]interface{})["ip"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, ip)

		user, ok := node.(map[string]interface{})["user"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, user)

		// Port is 22 by hardcoded now.
		err = ssh.SimpleCheck(ip+":22", user, privatekey)
		assert.NoError(t, err)
	}
}
