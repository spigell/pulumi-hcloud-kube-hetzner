package wireguard

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Enabled         bool
	IP              string
	Firewall        *Firewall
	CIDR            string           `json:"cidr"`
	AdditionalPeers []AdditionalPeer `json:"additional-peers" yaml:"additional-peers"`
}

type Firewall struct {
	Hetzner *HetznerFirewall
}

type AdditionalPeer struct {
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips"`
	Endpoint   string
	PublicKey  string
}

type HetznerFirewall struct {
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips"`
}

type Peer struct {
	ID                  string
	PrivateAddr         string
	PublicAddr          string
	PublicKey           string
	PrivateKey          string
	AllowedIps          []string
	Endpoint            string
	PersistentKeepalive int
}

type Peers []Peer

type WgPeer struct {
	Endpoint            string `toml:"Endpoint,omitempty"`
	PublicKey           string
	AllowedIps          []string
	PersistentKeepalive int
}

type WgInterface struct {
	ListenPort int
	PrivateKey string
	Address    string
}

type WgConfig struct {
	Interface WgInterface
	Peer      []WgPeer
}

func ToPeers(peers []Peer) Peers {
	p := make(Peers, 0)
	for _, s := range peers {
		p = append(p, s)
	}
	return p
}

func (t Peers) without(id string) Peers {
	without := make(Peers, 0)
	for _, s := range t {
		if s.ID == id {
			continue
		}
		without = append(without, s)
	}
	return without
}

func (t Peers) Get(id string) Peer {
	var found Peer
	for _, s := range t {
		if s.ID == id {
			return s
		}
	}
	return found
}

func (t Peers) getWgPeers() []WgPeer {
	peers := make([]WgPeer, 0)
	for _, p := range t {
		peer := &WgPeer{
			PublicKey:           p.PublicKey,
			Endpoint:            p.Endpoint,
			AllowedIps:          p.AllowedIps,
			PersistentKeepalive: p.PersistentKeepalive,
		}
		peers = append(peers, *peer)
	}
	return peers
}

var re = regexp.MustCompile(`([\d])]`)

func renderConfig(cfg *WgConfig) (string, error) {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(cfg); err != nil {
		return "", err
	}
	return re.ReplaceAllString(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(buf.String(),
						"= [", "= "),
					"[[", "["),
				"]]", "]"),
			"\"", ""),
		`$1`), nil
}
