package network

import (
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/stretchr/testify/require"
)

func TestPrefixAllocator(t *testing.T) {
	var ctx *program.Context
	net := New(ctx, &Config{
		CIDR: "192.168.0.0/20",
	})

	err := net.PickSubnet("from-start", FromEnd)
	require.NoError(t, err)

	err = net.PickSubnet("from-end", FromStart)
	require.NoError(t, err)

	l := len(net.ipam.Subnets)

	require.Equal(t, 16, l)
	require.Equal(t, "192.168.0.0/24", net.IPAM().Subnets[0].CIDR)
	require.True(t, net.IPAM().Subnets[0].Used)
	require.Equal(t, "192.168.15.0/24", net.IPAM().Subnets[l-1].CIDR)
	require.True(t, net.IPAM().Subnets[l-1].Used)
}
