package k3s

import (
	"fmt"
	"net"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/info"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
)

type K3S struct {
	order    int
	role     string
	leaderIP pulumi.StringOutput

	ID  string
	OS  info.OSInfo
	Sys *info.Info

	Config *Config
}

type Provisioned struct {
	resources []pulumi.Resource
}

func New(id string, role string, os info.OSInfo, config *Config) *K3S {
	if role == variables.ServerRole {
		config.K3S = config.K3S.WithServerDefaults()
	}

	config.K3S = config.K3S.WithoutDuplicates()

	return &K3S{
		role:   role,
		ID:     id,
		OS:     os,
		Config: config,
	}
}

func chooseDNSIP(s string) (string, error) {
	ip, _, err := net.ParseCIDR(s)
	if err != nil {
		return "", err
	}

	// Set the last octet to 10
	ip.To4()[3] = 10

	return net.IP(ip).To4().String(), nil
}

func (k *K3S) WithSysInfo(info *info.Info) *K3S {
	k.Sys = info

	return k
}

func (k *K3S) WithLeaderIp(ip pulumi.StringOutput) *K3S {
	k.leaderIP = ip

	return k
}

func (k *K3S) SetOrder(order int) {
	k.order = order
}

func (k *K3S) Order() int {
	return k.order
}

func (k *K3S) Up(ctx *pulumi.Context, con *connection.Connection, deps []pulumi.Resource, payload []interface{}) (modules.Output, error) {
	if k.role == variables.ServerRole {
		if k.Config.K3S.ClusterDNS == "" {
			ip, err := chooseDNSIP(k.Config.K3S.ServiceCidr)
			if err != nil {
				return nil, fmt.Errorf("error while choosing DNS IP: %w", err)
			}
			k.Config.K3S.ClusterDNS = ip
		}
	}

	res := make([]pulumi.Resource, 0)
	// Add dependencies to the resources from another module.
	// It is needed for restart k3s if wireguard restarted.
	res = append(res, deps...)

	install, err := k.install(ctx, con, deps)
	if err != nil {
		return nil, fmt.Errorf("error while installing: %w", err)
	}
	res = append(res, install)

	// payload[0] is wireguard config
	// payload[1] is internal IP
	var config pulumi.StringOutput
	switch k.Sys.CommunicationMethod() {
	case variables.WgCommunicationMethod:
		config, _ = pulumi.All(payload[0].(pulumi.AnyOutput), k.leaderIP, con.IP).ApplyT(
			func(args []interface{}) (string, error) {
				wg := args[0].(*wireguard.WgConfig)

				rendered, err := k.CompleteConfig(wg.Interface.Address, args[1].(string), args[2].(string)).render()

				return string(rendered), err
			},
		).(pulumi.StringOutput)
	case variables.DefaultCommunicationMethod:
		config, _ = pulumi.All(k.leaderIP, con.IP).ApplyT(
			func(args []interface{}) (string, error) {
				rendered, err := k.CompleteConfig(args[1].(string), args[0].(string), args[1].(string)).render()

				return string(rendered), err
			},
		).(pulumi.StringOutput)

	// payload[0] is internal IP
	case variables.InternalCommunicationMethod:
		config, _ = pulumi.All(k.leaderIP, con.IP).ApplyT(
			func(args []interface{}) (string, error) {
				rendered, err := k.CompleteConfig(payload[0].(string), args[0].(string), args[1].(string)).render()

				return string(rendered), err
			},
		).(pulumi.StringOutput)
	}

	configure, err := k.configure(ctx, con, config, res)

	if err != nil {
		return nil, fmt.Errorf("error while configuring: %w", err)
	}

	res = append(res, configure...)

	return &Provisioned{
		resources: res,
	}, nil
}

func (p *Provisioned) Value() interface{} {
	return nil
}

func (p *Provisioned) Resources() []pulumi.Resource {
	return p.resources
}

// CompleteConfig completes k3s config with pulumi.Outputs values.
func (k *K3S) CompleteConfig(ip, leaderIP, externalIP string) *CompletedConfig {
	k.Config.K3S.FlannelIface = k.Sys.CommunicationIface()
	k.Config.K3S.NodeIP = ip
	k.Config.K3S.ExternalNodeIP = externalIP
	k.Config.K3S.Server = fmt.Sprintf("https://%s:6443", leaderIP)

	if k.role == variables.ServerRole {
		k.Config.K3S.BindAddress = ip
		k.Config.K3S.AdvertiseAddr = ip
	}

	// if we are the leader node
	if ip == leaderIP {
		k.Config.K3S.ClusterInit = true
		k.Config.K3S.Server = ""
	}

	return &CompletedConfig{
		k.Config.K3S,
	}
}
