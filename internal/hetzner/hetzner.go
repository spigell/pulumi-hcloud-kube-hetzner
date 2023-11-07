package hetzner

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/firewall"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/server"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

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
			node.Server.Firewall = &config.Firewall{}
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

		if !node.Server.Firewall.Hetzner.Dedicated() && node.Server.Firewall.Hetzner.Enabled && !node.Server.Firewall.Hetzner.DedicatedPool() {
			switch node.Role {
			case variables.ServerRole:
				firewalls[variables.ServerRole] = node.Server.Firewall.Hetzner
			case variables.AgentRole:
				firewalls[variables.AgentRole] = node.Server.Firewall.Hetzner
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

// FirewallConfigByID returns the firewall config for the node.
// If not found, then search a config for nodepool.
// If not found again, then return a config for role.
func (h *Hetzner) FirewallConfigByID(id, pool string) (*firewall.Config, error) {
	node := h.Servers[id]
	fw := node.Server.Firewall.Hetzner
	if enabled := fw.Enabled; !enabled {
		return nil, ErrFirewallDisabled
	}

	// For node
	if fw.Dedicated() {
		return fw, nil
	}

	// For pool
	poolFw := h.Firewalls[pool]
	if poolFw != nil {
		return poolFw, nil
	}

	switch role := node.Role; role {
	case variables.ServerRole:
		return h.Firewalls[variables.ServerRole], nil
	case variables.AgentRole:
		return h.Firewalls[variables.AgentRole], nil
	default:
		return nil, fmt.Errorf("unknown node role %s", role)
	}
}

func (h *Hetzner) Up(info *Deployed, keys *keypair.ECDSAKeyPair) (*Deployed, error) {
	nodes := make(map[string]*Server)
	firewalls := make(map[string]*firewall.Firewall)
	firewallsByNodeRole := make(map[string]pulumi.IntArray)
	firewallsByNodepool := make(map[string]pulumi.IntArray)

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
	for name, fw := range h.Firewalls {
		// If name is not role, then it is a pool name.
		fmt.Println(h.Firewalls)
		kind := "pool"
		if name == variables.ServerRole || name == variables.AgentRole {
			kind = "role"
		}
		firewall, err := firewall.New(fw).Up(h.ctx, fmt.Sprintf("%s-%s", kind, name))
		if err != nil {
			return nil, err
		}
		firewalls[name] = firewall
	}

	fmt.Println("firewalls", firewalls)

	interFw := NewInterconnectFirewall()

	for _, id := range utils.SortedMapKeys(h.Servers) {
		srv := h.Servers[id]

		// if the passwd is given by user, use the password from the config.
		// Check if we have a password in the state as well since we may have empty state.
		// Generate a new password if we do not have it in creating stage.
		if srv.Server.UserPasswd == "" && info.Servers[id] != nil {
			srv.Server.UserPasswd = info.Servers[id].LocalPassword
		}

		internalIP, pool := "", h.FindInPools(id)
		if h.Network.Config.Enabled {
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

		//nolint: gocritic // this is the only way to convert string to int
		nodeId := node.Resource.ID().ToStringOutput().ApplyT(func(id string) (int, error) {
			return strconv.Atoi(id)
		}).(pulumi.IntOutput)

		if srv.Server.Firewall.Hetzner.Enabled {
			// All nodes with enabled FW must be added to the interconnect firewall
			interFw.Ips = append(interFw.Ips, pulumi.Sprintf("%s/32", node.Resource.Ipv4Address))
			interFw.Ids = append(interFw.Ids, nodeId)

			if srv.Server.Firewall.Hetzner.Dedicated() {
				firewall, err := firewall.New(srv.Server.Firewall.Hetzner).Up(h.ctx, id)
				if err != nil {
					return nil, fmt.Errorf("failed to create a dedicated firewall for node %s: %w", id, err)
				}
				_, err = firewall.Attach(h.ctx, id, pulumi.IntArray{nodeId})
				if err != nil {
					return nil, fmt.Errorf("failed to attach a dedicated firewall to node %s: %w", id, err)
				}

				continue
			}

			if firewalls[pool] != nil {
				firewallsByNodepool[pool] = append(firewallsByNodepool[pool], nodeId)
				continue
			}
			fmt.Println(pool)

			firewallsByNodeRole[srv.Role] = append(firewallsByNodeRole[srv.Role], nodeId)
		}
	}

	for kind, ids := range firewallsByNodeRole {
		fmt.Println(kind)
		_, err := firewalls[kind].Attach(h.ctx, fmt.Sprintf("role-%s", kind), ids)
		if err != nil {
			return nil, fmt.Errorf("failed to attach the group firewall for nodes: %w", err)
		}
	}

	for pool, ids := range firewallsByNodepool {
		_, err := firewalls[pool].Attach(h.ctx, fmt.Sprintf("pool-%s", pool), ids)
		if err != nil {
			return nil, fmt.Errorf("failed to attach the nodepool firewall for nodes: %w", err)
		}
	}

	// Create a global firewall to allow communication between all nodes
	if len(interFw.Ids) != 0 {
		if err := interFw.Up(h.ctx); err != nil {
			return nil, fmt.Errorf("failed to create a interconnect firewall for nodes: %w", err)
		}
	}

	return &Deployed{
		Servers: nodes,
	}, nil
}
