package wireguard

import (
	"fmt"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	remotefile "github.com/spigell/pulumi-file/sdk/go/file/remote"
)

const (
	defaultListenPort = 51822
	defaultIface      = "kubewg0"
	defaultCIDR       = "192.168.180.0/24"
)

type Wireguard struct {
	order int
	built *Config

	ID            string
	Self          Peer
	Neighbours    []Peer
	NeighboursIPS pulumi.StringMapOutput
	ListenPort    int
	Iface         string
	Mgmt          bool
	Config        *config.Wireguard
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

func New(id string, cfg *config.Wireguard) *Wireguard {
	if cfg.CIDR == "" {
		cfg.CIDR = defaultCIDR
	}

	if cfg.Firewall == nil {
		cfg.Firewall = &config.WGFirewall{
			Hetzner: &config.ServiceFirewall{
				AllowedIps: FWAllowedIps,
			},
		}
	}

	return &Wireguard{
		ID:         id,
		ListenPort: defaultListenPort,
		Iface:      defaultIface,
		Config:     cfg,
	}
}

func (w *Wireguard) SetOrder(order int) {
	w.order = order
}

func (w *Wireguard) Order() int {
	return w.order
}

func (w *Wireguard) Up(ctx *pulumi.Context, con *connection.Connection, deps []pulumi.Resource) (modules.Output, error) {
	w.built = w.NewConfig()

	resources := make([]pulumi.Resource, 0)

	deployed, err := remotefile.NewFile(ctx, fmt.Sprintf("wg-cluster-%s", w.ID), &remotefile.FileArgs{
		Connection: con.RemoteFile(),
		UseSudo:    pulumi.Bool(true),
		Content:    w.built.Render(),
		Path:       pulumi.Sprintf("/etc/wireguard/%s.conf", w.Iface),
		SftpPath:   pulumi.String("/usr/libexec/ssh/sftp-server"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn(deps))
	if err != nil {
		return nil, err
	}
	resources = append(resources, deployed)

	restartCommand := pulumi.Sprintf(strings.Join([]string{
		"sudo systemctl disable --now wg-quick@%s",
		"sudo systemctl enable --now wg-quick@%s",
	}, " && "), w.Iface, w.Iface)

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

type Config struct {
	config pulumi.AnyOutput
}

func (w *Wireguard) NewConfig() *Config {
	return &Config{
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

func (c *Config) Render() pulumi.StringOutput {
	return c.config.ApplyT(func(config interface{}) string {
		wgConfig, err := renderConfig(config.(*WgConfig))
		if err != nil {
			panic(fmt.Sprintf("Error while render Wireguard config %e", err))
		}

		return wgConfig
	}).(pulumi.StringOutput)
}

func (c *Config) Content() interface{} {
	return c.config
}

func (c *Provisioned) Value() interface{} {
	return c.Config
}

func (c *Provisioned) Resources() []pulumi.Resource {
	return c.resources
}
