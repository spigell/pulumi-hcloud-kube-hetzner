package ipam

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/seancfoley/ipaddress-go/ipaddr"
)

const (
	// 24 bit network
	// In the future, we can add an ability to change subnet size.
	subnetSize = 24
)

type IPAM struct {
	internalIPS pulumi.ArrayMap
	Subnets     []*allocatedSubnet
}

type IPAMData struct {
	InternalIPS pulumi.ArrayMap `yaml:"-"`
	Subnets     []*allocatedSubnet
}

type allocatedSubnet struct {
	allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]

	Used     bool
	CIDR     string
	ID       string   `yaml:",omitempty"`
	TakenIPS []string `yaml:"taken-ips,omitempty"`
}

// Load loads an IPAM from a given IPAMData
func Load(data *IPAMData) *IPAM {
	return &IPAM{
		Subnets: withAllocators(data.Subnets),
	}
}

// FreshIPAM creates a new IPAM with a given CIDR
func FreshIPAM(cidr string) *IPAM {
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
			Used:      false,
			CIDR:      allocated.String(),
			allocator: subnetAllocator,
		})
	}

	return &IPAM{
		Subnets: allocatedSubnets,
	}
}

// GetFreeIP returns a free IP from the IPAM.
func (i *IPAM) GetFree(subnetID string) (string, error) {
	var ip string
	err := fmt.Errorf("network with id %s not found", subnetID)

	for _, subnet := range i.Subnets {
		if subnet.ID == subnetID {
			allocated := subnet.allocator.AllocateSize(1)

			for _, i := range []int64{0, 1} {
				if allocated.GetSegment(3).GetValue().Int64() == i {
					allocated = subnet.allocator.AllocateSize(1)
				}
			}

			if allocated == nil {
				return "", fmt.Errorf("allocator says: no more free IPs in subnet %s", subnet.CIDR)
			}

			ip = strings.Split(allocated.String(), "/")[0]
			err = nil
		}
	}

	return ip, err
}

func (i *IPAM) ToData() *IPAMData {
	return &IPAMData{
		InternalIPS: i.internalIPS,
		Subnets:     i.Subnets,
	}
}

func (i *IPAM) WithInternalIPS(ips pulumi.ArrayMap) *IPAM {
	i.internalIPS = ips

	return i
}

// withAllocators is a helper function to create a list of allocatedSubnet with the correct allocator
func withAllocators(subnets []*allocatedSubnet) []*allocatedSubnet {
	allocatedSubnets := make([]*allocatedSubnet, 0)
	for _, subnet := range subnets {
		var allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]
		allocator.SetReserved(2) // 2 reserved per block for network and broadcast

		if !subnet.Used {
			allocator.AddAvailable([]*ipaddr.IPAddress{
				ipaddr.NewIPAddressString(subnet.CIDR).GetAddress(),
			}...)

			subnet.allocator = allocator
			allocatedSubnets = append(allocatedSubnets, subnet)

			continue
		}

		sub := ipaddr.NewIPAddressString(subnet.CIDR).GetAddress()
		subnets := make([]*ipaddr.IPAddress, 0)

		iterator := sub.Iterator()
		for i := 0; i < 254 && iterator.HasNext(); i++ {
			ip := iterator.Next().WithoutPrefixLen()

			excluded := false
			for _, takenIP := range subnet.TakenIPS {
				if ip.String() == takenIP {
					excluded = true
					break
				}
			}

			if !excluded {
				subnets = append(subnets, ip)
			}
		}

		allocator.AddAvailable(subnets...)

		subnet.allocator = allocator
		allocatedSubnets = append(allocatedSubnets, subnet)
	}

	return allocatedSubnets
}
