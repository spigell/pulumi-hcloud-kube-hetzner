package hetzner

import (
	"errors"
	"fmt"
	"pulumi-hcloud-kube-hetzner/internal/config"
	"pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"pulumi-hcloud-kube-hetzner/internal/hetzner/server"
	"pulumi-hcloud-kube-hetzner/internal/utils"
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
	Pools     map[string][]string
	Network   *network.Network
}

type Deployed struct {
	Servers map[string]*Server
}

type Server struct {
	ID            pulumi.IDOutput
	LocalPassword string
	InternalIP    string
	Connection    *connection.Connection
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
		Pools:     make(map[string][]string),
	}
}

func (h *Hetzner) WithNetwork(cfg *network.Config) *Hetzner {
	h.Network = network.New(h.ctx, cfg)
	return h
}

// AddToPool adds a node to the pool.
// Pool in hetzner stage is a simple slice with id of nodes.
// It is used to identify subnet for the node.
func (h *Hetzner) AddToPool(pool, node string) {
	h.Pools[pool] = append(h.Pools[pool], node)
}

func (h *Hetzner) FindInPools(node string) string {
	for pool := range h.Pools {
		for _, n := range h.Pools[pool] {
			if n == node {
				return pool
			}
		}
	}
	return ""
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

func (h *Hetzner) Up(info *Deployed, keys *keypair.ECDSAKeyPair) (*Deployed, error) {
	nodes := make(map[string]*Server)
	firewalls := make(map[string]*firewall.Firewall)
	firewallsByNodeRole := make(map[string]pulumi.IntArray)

	key, err := h.NewSSHKey(keys.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh key: %w", err)
	}

	var net *network.Deployed
	if h.Network.Config.Enabled {
		net, err = h.Network.Up()
		if err != nil {
			return nil, fmt.Errorf("failed to configure the network: %w", err)
		}
	}

	// Create a dedicated firewall for master (servers) and agents (if exists) nodes separattely
	for kind, fw := range h.Firewalls {
		firewall, err := firewall.New(fw).Up(h.ctx, fmt.Sprintf("role-%s", kind))
		if err != nil {
			return nil, err
		}
		firewalls[kind] = firewall
	}

	for _, id := range utils.SortedMapKeys(h.Servers) {
		srv := h.Servers[id]

		// if the passwd is given by user, use the password from the config.
		// Check if we have a password in the state as well since we may have empty state.
		// Generate a new password if we do not have it in creating stage.
		if srv.Server.UserPasswd == "" && info.Servers[id] != nil {
			srv.Server.UserPasswd = info.Servers[id].LocalPassword
		}

		internalIP, pool := "none", ""
		if h.Network.Config.Enabled {
			pool = h.FindInPools(id)
			internalIP, err = net.Subnets[pool].GetFree()
			if err != nil {
				return nil, fmt.Errorf("failed to get free ip for node %s: %w", id, err)
			}
			net.Subnets[pool].IPs[id] = internalIP
		}

		s := server.New(srv.Server, key)
		if err := s.Validate(); err != nil {
			return nil, err
		}
		node, err := s.Up(h.ctx, id, net, pool)
		if err != nil {
			return nil, err
		}
		nodes[id] = &Server{
			ID:            node.Resource.ID(),
			LocalPassword: node.Password,
			InternalIP:    internalIP,
			Connection: &connection.Connection{
				IP:         node.Resource.Ipv4Address,
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
				pulumi.IntArray{node.Resource.ID().ToIDOutput().ApplyT(func(id string) (int, error) {
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
			firewallsByNodeRole[srv.Role] = append(firewallsByNodeRole[srv.Role], node.Resource.ID().ToStringOutput().ApplyT(func(id string) (int, error) {
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
