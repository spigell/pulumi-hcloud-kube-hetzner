package network

import (
	"testing"

	"github.com/seancfoley/ipaddress-go/ipaddr"
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

	require.Equal(t, 16, len(net.allocatedSubnets))
	require.Equal(t, net.takenSubnets[0].CIDR, "192.168.15.0/24")
	require.Equal(t, net.takenSubnets[1].CIDR, "192.168.0.0/24")
}

func TestAddrAllocator(t *testing.T) {
	var subnetAllocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]
	var subnetAllocator2 ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]
	s := &Subnet{
		CIDR:      "192.168.0.0/24",
		allocator: subnetAllocator,
	}

	s.allocator.AddAvailable([]*ipaddr.IPAddress{
		ipaddr.NewIPAddressString(s.CIDR).GetAddress(),
	}...,
	)
	_, _ = s.GetFree()

	ip1, _ := s.GetFree()
	ip2, _ := s.GetFree()
	ip3, _ := s.GetFree()
	ip4, _ := s.GetFree()
	ip5, _ := s.GetFree()
	ip6, _ := s.GetFree()

	require.Equal(t, "192.168.0.1", ip1)
	require.Equal(t, "192.168.0.2", ip2)
	require.Equal(t, "192.168.0.3", ip3)
	require.Equal(t, "192.168.0.4", ip4)
	require.Equal(t, "192.168.0.5", ip5)
	require.Equal(t, "192.168.0.6", ip6)

	s2 := &Subnet{
		CIDR:      "192.168.1.0/24",
		allocator: subnetAllocator2,
	}
	s2.allocator.AddAvailable([]*ipaddr.IPAddress{
		ipaddr.NewIPAddressString(s2.CIDR).GetAddress(),
	}...,
	)

	ip1, _ = s2.GetFree()
	ip2, _ = s2.GetFree()
	ip3, _ = s2.GetFree()

	require.Equal(t, "192.168.1.0", ip1)
	require.Equal(t, "192.168.1.1", ip2)
	require.Equal(t, "192.168.1.2", ip3)
}
