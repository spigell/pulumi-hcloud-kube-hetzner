package ipam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddrAllocator(t *testing.T) {
	ipam := FreshIPAM("192.168.0.0/24")
	subnetID := "main"
	ipam.Subnets[0].ID = subnetID

	ip1, _ := ipam.GetFree(subnetID)
	ip2, _ := ipam.GetFree(subnetID)
	ip3, _ := ipam.GetFree(subnetID)
	ip4, _ := ipam.GetFree(subnetID)
	ip5, _ := ipam.GetFree(subnetID)
	ip6, _ := ipam.GetFree(subnetID)

	require.Equal(t, "192.168.0.2", ip1)
	require.Equal(t, "192.168.0.3", ip2)
	require.Equal(t, "192.168.0.4", ip3)
	require.Equal(t, "192.168.0.5", ip4)
	require.Equal(t, "192.168.0.6", ip5)
	require.Equal(t, "192.168.0.7", ip6)
}
