package hetzner

import (
	"errors"
	"fmt"
	"pulumi-hcloud-kube-hetzner/internal/config"
	"pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"pulumi-hcloud-kube-hetzner/internal/hetzner/server"
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"
	"strconv"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var ErrFirewallDisabled = errors.New("firewall is disabled")

type Hetzner struct {
	ctx       *pulumi.Context
	Servers   map[string]*config.Node
	Firewalls map[string]*firewall.Config
}

type Deployed struct {
	Servers map[string]*Server
}

type Server struct {
	ID         pulumi.IDOutput
	Connection *connection.Connection
}

func New(ctx *pulumi.Context, nodes []*config.Node) *Hetzner {
	servers := make(map[string]*config.Node)
	firewalls := make(map[string]*firewall.Config)

	for _, node := range nodes {
		servers[node.ID] = node
		if node.Server == nil {
			node.Server = &config.Server{}
		}

		if node.Server.Firewall == nil {
			node.Server.Firewall = &config.Firewall{
				Firewalld: &config.FWFirewalld{
					Enabled: false,
				},
			}
		}

		if node.Server.Firewall.Hetzner == nil {
			node.Server.Firewall.Hetzner = &firewall.Config{
				Enabled: false,
			}
		}

		if node.Server.Firewall.Hetzner.AdditionalRules == nil {
			node.Server.Firewall.Hetzner.AdditionalRules = make([]*firewall.Rule, 0)
		}

		if node.Server.Firewall.Hetzner.SSH == nil {
			node.Server.Firewall.Hetzner.SSH = &firewall.SSH{
				Allow: false,
			}
		}

		if !node.Server.Firewall.Hetzner.Dedicated() && node.Server.Firewall.Hetzner.Enabled {
			switch node.Role {
			case "server":
				firewalls["server"] = node.Server.Firewall.Hetzner
			case "agent":
				firewalls["agent"] = node.Server.Firewall.Hetzner
			}
		}
	}

	return &Hetzner{
		ctx:       ctx,
		Servers:   servers,
		Firewalls: firewalls,
	}
}

func (h *Hetzner) FirewallConfigByIDOrRole(id string) (*firewall.Config, error) {
	node := h.Servers[id]
	fw := node.Server.Firewall.Hetzner
	if enabled := fw.Enabled; !enabled {
		return nil, ErrFirewallDisabled
	}

	if fw.Dedicated() {
		return fw, nil
	}

	switch role := node.Role; role {
	case "server":
		return h.Firewalls["server"], nil
	case "agent":
		return h.Firewalls["agent"], nil
	default:
		return nil, fmt.Errorf("unknown node role %s", role)
	}
}

func (h *Hetzner) Up(keys *keypair.ECDSAKeyPair) (*Deployed, error) {
	nodes := make(map[string]*Server)
	firewalls := make(map[string]*firewall.Firewall)
	firewallsByNodeRole := make(map[string]pulumi.IntArray)

	// Create a dedicated firewall for master (servers) and agents (if exists) nodes separattely
	for kind, fw := range h.Firewalls {
		firewall, err := firewall.New(fw).Up(h.ctx, fmt.Sprintf("role-%s", kind))
		if err != nil {
			return nil, err
		}
		firewalls[kind] = firewall
	}

	for id, srv := range h.Servers {
		s := server.New(srv.Server, keys)
		if err := s.Validate(); err != nil {
			return nil, err
		}
		node, err := s.Up(h.ctx, id)
		if err != nil {
			return nil, err
		}
		nodes[id] = &Server{
			ID: node.ID(),
			Connection: &connection.Connection{
				IP:         node.Ipv4Address,
				PrivateKey: keys.PrivateKey,
				User:       srv.Server.UserName,
			},
		}

		if srv.Server.Firewall.Hetzner.Enabled && srv.Server.Firewall.Hetzner.Dedicated() {
			firewall, err := firewall.New(srv.Server.Firewall.Hetzner).
				Up(h.ctx, id)
			if err != nil {
				return nil, fmt.Errorf("failed to create dedicated firewall for node %s: %w", id, err)
			}
			_, err = firewall.Attach(h.ctx, id,
				//nolint: gocritic // this is the only way to convert string to int
				pulumi.IntArray{node.ID().ToIDOutput().ApplyT(func(id string) (int, error) {
					return strconv.Atoi(id)
				}).(pulumi.IntOutput)},
			)
			if err != nil {
				return nil, fmt.Errorf("failed to attach dedicated firewall for node %s: %w", id, err)
			}
			continue
		}

		if srv.Server.Firewall.Hetzner.Enabled {
			//nolint: gocritic // this is the only way to convert string to int
			firewallsByNodeRole[srv.Role] = append(firewallsByNodeRole[srv.Role], node.ID().ToStringOutput().ApplyT(func(id string) (int, error) {
				return strconv.Atoi(id)
			}).(pulumi.IntOutput))
		}
	}

	for kind, ids := range firewallsByNodeRole {
		_, err := firewalls[kind].Attach(h.ctx, fmt.Sprintf("role-%s", kind), ids)
		if err != nil {
			return nil, fmt.Errorf("failed to attach the group level firewall for nodes: %w", err)
		}
	}

	return &Deployed{
		Servers: nodes,
	}, nil
}
