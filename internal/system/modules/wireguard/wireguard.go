package wireguard

import (
	"fmt"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/info"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	remotefile "github.com/spigell/pulumi-file/sdk/go/file/remote"
)

var restartCommand = pulumi.Sprintf(strings.Join([]string{
	"sudo systemctl disable --now wg-quick@%s",
	"sudo systemctl enable --now wg-quick@%s",
	"sudo systemctl status wg-quick@%s",
}, " && "), Iface, Iface, Iface)

const (
	// Iface is the name of interface. It is not allowed to change it.
	defaultListenPort = 51822
	defaultCIDR       = "192.168.180.0/24"
)

var Iface = variables.WGIface

type Wireguard struct {
	order int
	built *CompletedConfig

	ID            string
	Self          Peer
	Neighbours    []Peer
	OS            info.OSInfo
	NeighboursIPS pulumi.StringMapOutput
	ListenPort    int
	Config        *Config
}

type Provisioned struct {
	resources []pulumi.Resource
	Config    pulumi.AnyOutput
}

var packages = map[string][]string{
	"microos": {"wireguard-tools"},
}

func GetRequiredPkgs(os string) []string {
	return packages[os]
}

func New(id string, os info.OSInfo, cfg *Config) *Wireguard {
	if cfg.CIDR == "" {
		cfg.CIDR = defaultCIDR
	}

	return &Wireguard{
		ID:         id,
		OS:         os,
		ListenPort: defaultListenPort,
		Config:     cfg,
	}
}

func (w *Wireguard) SetOrder(order int) {
	w.order = order
}

func (w *Wireguard) Order() int {
	return w.order
}

func (w *Wireguard) Up(ctx *pulumi.Context, con *connection.Connection, deps []pulumi.Resource, _ []interface{}) (modules.Output, error) {
	w.built = w.CompleteConfig()

	resources := make([]pulumi.Resource, 0)

	deployed, err := remotefile.NewFile(ctx, fmt.Sprintf("wg-cluster-%s", w.ID), &remotefile.FileArgs{
		Connection:  con.RemoteFile(),
		UseSudo:     pulumi.Bool(true),
		Content:     pulumi.ToSecret(w.built.Render()).(pulumi.StringOutput),
		Path:        pulumi.Sprintf("/etc/wireguard/%s.conf", Iface),
		SftpPath:    pulumi.String(w.OS.SFTPServerPath()),
		Permissions: pulumi.String("664"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn(deps))
	if err != nil {
		return nil, err
	}
	resources = append(resources, deployed)

	restarted, err := remote.NewCommand(ctx, fmt.Sprintf("wg-restart-%s", w.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     restartCommand,
		Triggers: pulumi.Array{
			deployed.Md5sum,
			deployed.Permissions,
			deployed.Connection,
			deployed.Path,
		},
	}, pulumi.DependsOn([]pulumi.Resource{deployed}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5m"}),
		pulumi.DeleteBeforeReplace(true),
	)
	if err != nil {
		return nil, err
	}

	resources = append(resources, restarted)

	return &Provisioned{
		resources: resources,
		Config:    w.built.config,
	}, nil
}

type CompletedConfig struct {
	config pulumi.AnyOutput
}

func (w *Wireguard) CompleteConfig() *CompletedConfig {
	return &CompletedConfig{
		config: w.NeighboursIPS.ApplyT(func(ips interface{}) *WgConfig {
			if len(w.Config.AdditionalPeers) > 0 {
				for _, p := range w.Config.AdditionalPeers {
					additionalPeer := Peer{
						PublicKey:  p.PublicKey,
						Endpoint:   p.Endpoint,
						AllowedIps: p.AllowedIps,
					}
					w.Neighbours = append(w.Neighbours, additionalPeer)
				}
			}

			peersWithoutSelf := ToPeers(w.Neighbours)

			for k, v := range peersWithoutSelf {
				peersWithoutSelf[k].PersistentKeepalive = 25
				if len(peersWithoutSelf[k].AllowedIps) == 0 {
					peersWithoutSelf[k].AllowedIps = []string{fmt.Sprintf("%s/32", v.PrivateAddr)}
				}

				ips := ips.(map[string]string)

				if ips[v.ID] != "" {
					peersWithoutSelf[k].Endpoint = fmt.Sprintf("%s:%d", ips[v.ID], w.ListenPort)
				}
			}

			config := &WgConfig{
				Peer: peersWithoutSelf.getWgPeers(),
				Interface: WgInterface{
					Address:    w.Self.PrivateAddr,
					PrivateKey: w.Self.PrivateKey,
					ListenPort: w.ListenPort,
				},
			}

			return config
		}).(pulumi.AnyOutput),
	}
}

func (c *CompletedConfig) Render() pulumi.StringOutput {
	return c.config.ApplyT(func(config interface{}) string {
		wgConfig, err := renderConfig(config.(*WgConfig))
		if err != nil {
			panic(fmt.Sprintf("Error while render Wireguard config %e", err))
		}

		return wgConfig
	}).(pulumi.StringOutput)
}

func (c *CompletedConfig) Content() interface{} {
	return c.config
}

func (c *Provisioned) Value() interface{} {
	return c.Config
}

func (c *Provisioned) Resources() []pulumi.Resource {
	return c.resources
}
