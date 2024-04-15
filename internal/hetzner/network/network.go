package network

import (
	"fmt"
	"strconv"

	hcloudapi "github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network/ipam"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

const (
	FromEnd   = "end"
	FromStart = "start"

	defaultZone    = "eu-central"
	defaultNetCIDR = "10.20.0.0/16"
)

// Params can be used to configure hetzner network with given CIDR and zone.
type Params struct {
	// CIDR of private network. Default is 10.20.0.0/16
	CIDR    string
	Enabled bool
	Zone    string
}

type Network struct {
	ctx  *program.Context
	ipam *ipam.IPAM

	Config *Params
}

type Deployed struct {
	ID      pulumi.IntOutput
	Subnets map[string]*Subnet
}

type Subnet struct {
	Resource *hcloud.NetworkSubnet
}

func New(ctx *program.Context, cfg *Params) *Network {
	if cfg.CIDR == "" {
		cfg.CIDR = defaultNetCIDR
	}

	if cfg.Zone == "" {
		cfg.Zone = defaultZone
	}

	return &Network{
		ipam:   ipam.FreshIPAM(cfg.CIDR),
		ctx:    ctx,
		Config: cfg,
	}
}

func (n *Network) WithIPAM(ipam *ipam.IPAM) *Network {
	n.ipam = ipam
	return n
}

func (n *Network) PickSubnet(id string, from string) error {
	for _, subnet := range n.ipam.Subnets {
		if subnet.ID == id {
			return nil
		}
	}

	switch from {
	case FromEnd:
		// Take last subnet
		for i := 1; i < 254; i++ {
			l := len(n.ipam.Subnets) - i
			subnet := n.ipam.Subnets[l]
			if !subnet.Used {
				n.ipam.Subnets[l].Used = true
				n.ipam.Subnets[l].ID = id
				break
			}
		}
	case FromStart:
		// Take first subnet
		for i, subnet := range n.ipam.Subnets {
			if !subnet.Used {
				n.ipam.Subnets[i].Used = true
				n.ipam.Subnets[i].ID = id
				// Add blocklist for 0 and 1 ip
				break
			}
		}
	default:
		return fmt.Errorf("unknown from: %s", from)
	}

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
	for _, subnet := range n.ipam.Subnets {
		if subnet.Used {
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
				Resource: s,
			}
		}
	}

	return &Deployed{
		ID:      converted,
		Subnets: subnets,
	}, nil
}

func (n *Network) GetFree(subnetID string) (string, error) {
	return n.ipam.GetFree(subnetID)
}

func (n *Network) IPAM() *ipam.IPAM {
	return n.ipam
}
