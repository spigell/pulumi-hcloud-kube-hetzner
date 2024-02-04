package network

import (
	"github.com/seancfoley/ipaddress-go/ipaddr"
)

const (
	// 24 bit network
	// In the future, we can add an ability to change subnet size.
	subnetSize = 24
)

func ipam(cidr string) []*allocatedSubnet {
	var allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]

	allocator.SetReserved(2) // 2 reserved per block for network and broadcast
	allocator.AddAvailable(
		[]*ipaddr.IPAddress{
			ipaddr.NewIPAddressString(cidr).GetAddress(),
		}...,
	)

	allocatedSubnets := make([]*allocatedSubnet, 0)

	// 256 /24 network in /16 network
	// I think it will be enough for us
	for i := 0; i < 256; i++ {
		allocated := allocator.AllocateBitLen(32 - subnetSize)
		// Allocator is clever.
		// If no more space in /16 network, stop
		if allocated == nil {
			break
		}
		var subnetAllocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]
		subnetAllocator.AddAvailable(
			[]*ipaddr.IPAddress{
				ipaddr.NewIPAddressString(allocated.String()).GetAddress(),
			}...,
		)

		allocatedSubnets = append(allocatedSubnets, &allocatedSubnet{
			Free:      true,
			Cidr:      allocated.String(),
			allocator: subnetAllocator,
		})
	}

	return allocatedSubnets
}

