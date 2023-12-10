package network

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/seancfoley/ipaddress-go/ipaddr"
)

const (
	// 24 bit network
	// In the future, we can add an ability to change subnet size.
	subnetSize = 24
	FromEnd    = "end"
	FromStart  = "start"

	defaultZone    = "eu-central"
	defaultNetCIDR = "10.20.0.0/16"
)

type Network struct {
	ctx              *pulumi.Context
	allocatedSubnets []*allocatedSubnet
	takenSubnets     []*TakenSubnet
	allocator        ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]

	Config *Config
}

type allocatedSubnet struct {
	free      bool
	allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]
	cidr      string
}

type TakenSubnet struct {
	allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]

	ID   string
	CIDR string
}

type Config struct {
	CIDR    string
	Enabled bool
	Zone    string
}

type Deployed struct {
	ID      pulumi.IntOutput
	Subnets map[string]*Subnet
}

type Subnet struct {
	CIDR      string
	IPs       map[string]string
	allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]
	Resource  *hcloud.NetworkSubnet
}

func New(ctx *pulumi.Context, cfg *Config) *Network {
	if cfg.CIDR == "" {
		cfg.CIDR = defaultNetCIDR
	}

	if cfg.Zone == "" {
		cfg.Zone = defaultZone
	}

	var allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]

	allocator.SetReserved(2) // 2 reserved per block for network and broadcast
	allocator.AddAvailable(
		[]*ipaddr.IPAddress{
			ipaddr.NewIPAddressString(cfg.CIDR).GetAddress(),
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
			free:      true,
			cidr:      allocated.String(),
			allocator: subnetAllocator,
		})
	}

	return &Network{
		allocator:        allocator,
		allocatedSubnets: allocatedSubnets,
		ctx:              ctx,
		Config:           cfg,
	}
}

func (n *Network) PickSubnet(id string, from string) error {
	var taken *allocatedSubnet

	switch from {
	case FromEnd:
		// Take last subnet
		for i := 1; i < 254; i++ {
			l := len(n.allocatedSubnets) - i
			subnet := n.allocatedSubnets[l]
			if subnet.free {
				taken = subnet
				n.allocatedSubnets[l].free = false
				break
			}
		}
	case FromStart:
		// Take first subnet
		for i, subnet := range n.allocatedSubnets {
			if subnet.free {
				taken = subnet
				n.allocatedSubnets[i].free = false
				break
			}
		}
	default:
		return fmt.Errorf("unknown from: %s", from)
	}

	n.takenSubnets = append(n.takenSubnets, &TakenSubnet{
		ID:        id,
		allocator: taken.allocator,
		CIDR:      taken.cidr,
	})

	return nil
}

func (n *Network) Up(opts []pulumi.ResourceOption) (*Deployed, error) {
	net, err := hcloud.NewNetwork(n.ctx, fmt.Sprintf("%s-%s", n.ctx.Project(), n.ctx.Stack()), &hcloud.NetworkArgs{
		IpRange: pulumi.String(n.Config.CIDR),
		Name:    pulumi.String(fmt.Sprintf("%s-%s", n.ctx.Project(), n.ctx.Stack())),
	}, opts...)
	if err != nil {
		return nil, err
	}
	//nolint: gocritic // this is the only way to convert string to int
	converted := net.ID().ApplyT(func(id string) (int, error) {
		return strconv.Atoi(id)
	}).(pulumi.IntOutput)

	subnets := make(map[string]*Subnet)
	for _, subnet := range n.takenSubnets {
		s, err := hcloud.NewNetworkSubnet(n.ctx, subnet.ID, &hcloud.NetworkSubnetArgs{
			NetworkId:   converted,
			Type:        pulumi.String("cloud"),
			IpRange:     pulumi.String(subnet.CIDR),
			NetworkZone: pulumi.String(n.Config.Zone),
		}, append(opts, pulumi.DeleteBeforeReplace(true))...)
		if err != nil {
			return nil, fmt.Errorf("failed to create subnet %s: %w", subnet.ID, err)
		}
		// Rule: id of pool is id of the needed subnet
		subnets[subnet.ID] = &Subnet{
			CIDR:      subnet.CIDR,
			allocator: subnet.allocator,
			IPs:       make(map[string]string),
			Resource:  s,
		}
		// Allocate .0 in subnet. It is not needed.
		_, _ = subnets[subnet.ID].GetFree()
		// Allocate .1 in subnet. It is reserved for router.
		_, _ = subnets[subnet.ID].GetFree()
	}

	return &Deployed{
		ID:      converted,
		Subnets: subnets,
	}, nil
}

func (s *Subnet) GetFree() (string, error) {
	ip := s.allocator.AllocateSize(1)
	if ip == nil {
		return "", fmt.Errorf("allocator says: no more free IPs in subnet %s", s.CIDR)
	}

	return strings.Split(ip.String(), "/")[0], nil
}
