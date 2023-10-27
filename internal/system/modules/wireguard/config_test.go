package wireguard

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const ListenPort = 51820

var TestPeers = []Peer{
	{
		ID:          "test1",
		PrivateAddr: "192.168.25.1",
		PublicAddr:  "192.186.0.25",
		PublicKey:   "PublicKey1",
		PrivateKey:  "PrivateKey1",
		AllowedIps:  []string{"192.168.25.1/32", "10.0.0.0/8"},
	},
	{
		ID:          "test2",
		PrivateAddr: "192.168.25.2",
		PublicAddr:  "192.186.0.26",
		PublicKey:   "PublicKey2",
		PrivateKey:  "PrivateKey2",
		AllowedIps:  []string{"192.168.25.2/32"},
	},
	{
		ID:          "test3",
		PrivateAddr: "192.168.25.3",
		PublicAddr:  "192.186.0.27",
		PublicKey:   "PublicKey3",
		PrivateKey:  "PrivateKey3",
		AllowedIps:  []string{"192.168.25.3/32"},
	},
}

func TestRenderConfig(t *testing.T) {
	for _, peer := range TestPeers {
		peersWithoutSelf := ToPeers(TestPeers).without(peer.ID)
		for k, v := range peersWithoutSelf {
			peersWithoutSelf[k].Endpoint = fmt.Sprintf("%s:%d", v.PublicAddr, ListenPort)
		}

		config := &WgConfig{
			Peer:      peersWithoutSelf.getWgPeers(),
			Interface: WgInterface{Address: peer.PrivateAddr, PrivateKey: peer.PrivateKey, ListenPort: ListenPort},
		}
		got, err := renderConfig(config)
		require.NoError(t, err)

		fmt.Println(got)
		fmt.Println("---")
	}
}
