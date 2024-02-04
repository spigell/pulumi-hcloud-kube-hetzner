package network

import (
	"fmt"
	"strconv"
	"strings"

	hcloudapi "github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

const (
	FromEnd    = "end"
	FromStart  = "start"

	defaultZone    = "eu-central"
	defaultNetCIDR = "10.20.0.0/16"
)

type Network struct {
	ctx              *program.Context
	allocatedSubnets []*allocatedSubnet
	takenSubnets     []*TakenSubnet

	Config *Config
}

type allocatedSubnet struct {
	allocator ipaddr.PrefixBlockAllocator[*ipaddr.IPAddress]

	Free      bool
	Cidr      string
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

func New(ctx *program.Context, cfg *Config) *Network {
	if cfg.CIDR == "" {
		cfg.CIDR = defaultNetCIDR
	}

	if cfg.Zone == "" {
		cfg.Zone = defaultZone
	}

	subnets, err := loadNetworkStateFile(ctx.Context().Stack())
	if err != nil {
		subnets = ipam(cfg.CIDR)
	}


	return &Network{
		allocatedSubnets: subnets,
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
			if subnet.Free {
				taken = subnet
				n.allocatedSubnets[l].Free = false
				break
			}
		}
	case FromStart:
		// Take first subnet
		for i, subnet := range n.allocatedSubnets {
			if subnet.Free {
				taken = subnet
				n.allocatedSubnets[i].Free = false
				break
			}
		}
	default:
		return fmt.Errorf("unknown from: %s", from)
	}

	n.takenSubnets = append(n.takenSubnets, &TakenSubnet{
		ID:        id,
		allocator: taken.allocator,
		CIDR:      taken.Cidr,
	})

	return nil
}

func (n *Network) Up() (*Deployed, error) {
	net, err := hcloud.NewNetwork(n.ctx.Context(), fmt.Sprintf("%s-%s", n.ctx.Context().Project(), n.ctx.Context().Stack()), &hcloud.NetworkArgs{
		IpRange: pulumi.String(n.Config.CIDR),
		Name:    pulumi.String(fmt.Sprintf("%s-%s", n.ctx.Context().Project(), n.ctx.Context().Stack())),
	}, n.ctx.Options()...)
	if err != nil {
		return nil, err
	}
	//nolint: gocritic // this is the only way to convert string to int
	converted := net.ID().ApplyT(func(id string) (int, error) {
		return strconv.Atoi(id)
	}).(pulumi.IntOutput)

	subnets := make(map[string]*Subnet)
	for _, subnet := range n.takenSubnets {
		s, err := hcloud.NewNetworkSubnet(n.ctx.Context(), subnet.ID, &hcloud.NetworkSubnetArgs{
			NetworkId:   converted,
			Type:        pulumi.String(hcloudapi.NetworkSubnetTypeCloud),
			IpRange:     pulumi.String(subnet.CIDR),
			NetworkZone: pulumi.String(n.Config.Zone),
		}, append(n.ctx.Options(), pulumi.DeleteBeforeReplace(true))...)
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
