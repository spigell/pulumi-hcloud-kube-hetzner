package system

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os/microos"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"inet.af/netaddr"
)

const (
	// The name of the master peer (connection from operator).
	wgMasterID   = "master"
	wgMasterCIDR = "10.120.150.0/24"
)

type WgCluster struct {
	Peers            pulumi.Map
	MasterConnection pulumi.StringOutput
}

type WGPeers struct {
	Peers []wireguard.Peer
	IPS   pulumi.StringMapOutput
}

func (c *Cluster) NewWgCluster(wgInfo map[string]*wireguard.WgConfig, servers map[string]*hetzner.Server) *WgCluster {
	provisionedWGPeers := make(pulumi.Map)

	peers := c.BuildWgPeers(wgInfo, servers)
	master := c.NewWgMaster()
	// Always generate a new wireguard config.
	// Even if it disabled for all nodes.
	master.Neighbours = peers.Without(wgMasterID)
	master.Self = peers.Peer(wgMasterID)
	master.NeighboursIPS = peers.MasterIPS
	cfg := master.CompleteConfig()

	provisionedWGPeers[wgMasterID] = cfg.Content().(pulumi.AnyOutput)

	for _, v := range *c {
		if v.OS.Wireguard() != nil {
			v.OS.Wireguard().Neighbours = peers.Without(v.ID)
			v.OS.Wireguard().Self = peers.Peer(v.ID)
			v.OS.Wireguard().NeighboursIPS = peers.IPS
		}
	}

	return &WgCluster{
		Peers:            provisionedWGPeers,
		MasterConnection: cfg.Render(),
	}
}

func (c *Cluster) NewWgMaster() *wireguard.Wireguard {
	// MicroOS is the only supported OS right now.
	// Let's choose it.
	return wireguard.New(wgMasterID, &microos.MicroOS{}, &wireguard.Config{
		Enabled: true,
	})
}

func (c *Cluster) BuildWgPeers(info map[string]*wireguard.WgConfig, servers map[string]*hetzner.Server) *WGPeers {
	connectionIPS := make(pulumi.StringMap)
	internalIPS := make(pulumi.StringMap)

	start, _, _ := net.ParseCIDR(wgMasterCIDR)
	masterIP := netaddr.MustParseIP(start.String())
	// Make a random amount of cicles to get a different IP for every cluster
	//nolint:gosec // There is no need to use crypto/rand here.
	for i := 0; i < rand.Intn(100+1); i++ {
		masterIP = masterIP.Next()
	}
	pk, _ := wgtypes.GeneratePrivateKey()
	// Always generate a new key for the master node.
	peers := []wireguard.Peer{
		{
			ID:          wgMasterID,
			PrivateAddr: masterIP.String(),
			PublicKey:   pk.PublicKey().String(),
			PrivateKey:  pk.String(),
		},
	}
	// But if the state has provided the information, use it.
	if info[wgMasterID] != nil {
		k := info[wgMasterID]
		pk, _ := wgtypes.ParseKey(k.Interface.PrivateKey)
		peers[0].PrivateKey = k.Interface.PrivateKey
		peers[0].PublicKey = pk.PublicKey().String()
		peers[0].PrivateAddr = k.Interface.Address
	}

	key, pub, ip := "", "", ""
	m := make(map[string]netaddr.IP)

	for _, sys := range *c {
		if sys.OS.Wireguard() == nil {
			continue
		}

		// Always generate a new keypair for every peer.
		generated, _ := wgtypes.GeneratePrivateKey()
		key = generated.String()
		pub = generated.PublicKey().String()

		// If the state has provided the information, use it.
		if info[sys.ID] != nil {
			k := info[sys.ID]
			pk, _ := wgtypes.ParseKey(k.Interface.PrivateKey)
			key = pk.String()
			pub = pk.PublicKey().String()
			ip = k.Interface.Address
			m[sys.ID] = netaddr.MustParseIP(ip)
		}

		// If the client has provided the IP of Peer, use it.
		// Overwriting the state information.
		if sys.OS.Wireguard().Config.IP != "" {
			ip = sys.OS.Wireguard().Config.IP
			m[sys.ID] = netaddr.MustParseIP(ip)
		}

		if m[sys.ID].IsZero() {
			ipo, _, err := net.ParseCIDR(sys.OS.Wireguard().Config.CIDR)
			if err != nil {
				// Rewrite this panic please
				panic(fmt.Sprintf("Can not parse CIDR for Wireguard! Is it a valid network? (%s)", err.Error()))
			}

			start := netaddr.MustParseIP(ipo.String())
			i := start.Next()
			for !free(m, i) {
				i = i.Next()
			}
			m[sys.ID] = i
		}

		peer := wireguard.Peer{
			ID:          sys.ID,
			PrivateKey:  key,
			PublicKey:   pub,
			PrivateAddr: m[sys.ID].String(),
		}

		connectionIPS[sys.ID] = servers[sys.ID].Connection.IP

		if servers[sys.ID].InternalIP != "" {
			internalIPS[sys.ID] = pulumi.String(servers[sys.ID].InternalIP).ToStringOutput()
		}

		peers = append(peers, peer)
	}

	return &WGPeers{
		Peers: peers,
		ConnectionIPS:   connnectionIPS.ToStringMapOutput(),
	}
}

func (w *WGPeers) Without(id string) []wireguard.Peer {
	peers := make([]wireguard.Peer, 0)
	for _, v := range w.Peers {
		if v.ID == id {
			continue
		}
		peer := wireguard.Peer{
			ID:          v.ID,
			PublicKey:   v.PublicKey,
			PrivateAddr: v.PrivateAddr,
			PublicAddr:  v.PublicAddr,
		}
		peers = append(peers, peer)
	}
	return peers
}

func (w *WGPeers) Peer(id string) wireguard.Peer {
	for _, v := range w.Peers {
		if v.ID == id {
			return v
		}
	}
	return wireguard.Peer{}
}

func free(m map[string]netaddr.IP, match netaddr.IP) bool {
	for _, n := range m {
		if n == match {
			return false
		}
	}
	return true
}
