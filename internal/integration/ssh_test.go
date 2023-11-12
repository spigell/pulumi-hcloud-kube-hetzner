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

func TestSSHConnectivity(t *testing.T) {
	t.Parallel()
	name := testSSHConnectivity

	ctx, cancel := context.WithDeadline(context.Background(), defaultDeadline)
	defer cancel()

	i, _ := New(ctx)

	if !slices.Contains(TestsByExampleName[i.Example.Name], name) {
		t.Skip()
	}

	out, err := i.Stack.Outputs(ctx)

	assert.NoError(t, err)

	keyPair, ok := out[phkh.KeyPairKey]
	assert.True(t, ok)

	privatekey, ok := keyPair.Value.(map[string]interface{})[phkh.PrivateKey].(string)
	assert.True(t, ok)

	nodes, ok := out[phkh.HetznerServersKey]
	assert.True(t, ok)

	for _, node := range nodes.Value.(map[string]interface{}) {
		n := node.(map[string]interface{})

		ip, ok := n["ip"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, ip)

		user, ok := n["user"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, user)

		// Port is 22 by hardcoded now.
		err := ssh.SimpleCheck(ip+":22", user, privatekey)
		assert.NoError(t, err)
	}
}
