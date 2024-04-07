package k3s

import (
	"fmt"
	"net"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/audit"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/info"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
)

const (
	// ManagedLabel is a label for node Label. Used for internal purposes.
	NodeManagedLabel = "phkh.io/managed=true"
	logDir           = "/var/lib/rancher/k3s/server/logs"
	auditPolicyFIle  = "/var/lib/audit.yaml"
)

type K3S struct {
	order              int
	role               string
	leaderIP           pulumi.StringOutput
	token              pulumi.StringOutput
	auditPolicyEnabled bool
	auditPolicyContent *string

	ID  string
	OS  info.OSInfo
	Sys *info.Info

	Config *Config
}

type Provisioned struct {
	resources []pulumi.Resource
	Outputs   *Outputs
}

type Outputs struct {
	KubeconfigForExport pulumi.AnyOutput
	KubeconfigForUsage  pulumi.AnyOutput
}

var packages = map[string][]string{
	"microos": {"k3s-selinux"},
}

func GetRequiredPkgs(os string) []string {
	return packages[os]
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

	return ip.To4().String(), nil
}

func (k *K3S) WithSysInfo(info *info.Info) *K3S {
	k.Sys = info

	return k
}

func (k *K3S) WithLeaderIP(ip pulumi.StringOutput) *K3S {
	k.leaderIP = ip

	return k
}

func (k *K3S) WithToken(token pulumi.StringOutput) *K3S {
	k.token = token

	return k
}

func (k *K3S) WithK8SAuditLog(log *audit.AuditLog) *K3S {
	k.auditPolicyEnabled = log.Enabled()

	if k.auditPolicyEnabled {
		k.auditPolicyContent = log.PolicyContent()
		k.Config.K3S.KubeAPIServerArgs = append(k.Config.K3S.KubeAPIServerArgs,
			fmt.Sprintf("audit-policy-file=%s", auditPolicyFIle),
			fmt.Sprintf("audit-log-path=%s/audit.log", logDir),
			fmt.Sprintf("audit-log-maxage=%d", log.AuditLogMaxAge()),
			fmt.Sprintf("audit-log-maxbackup=%d", log.AuditLogMaxBackup()),
			fmt.Sprintf("audit-log-maxsize=%d", log.AuditLogMaxSize()),
		)
	}

	return k
}

func (k *K3S) SetOrder(order int) {
	k.order = order
}

func (k *K3S) Order() int {
	return k.order
}

func (k *K3S) Up(ctx *program.Context, con *connection.Connection, deps []pulumi.Resource, payload []interface{}) (modules.Output, error) {
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

	var config pulumi.StringOutput
	switch k.Sys.CommunicationMethod() {
	case variables.PublicCommunicationMethod:
		config, _ = pulumi.All(k.leaderIP, con.IP, k.token).ApplyT(
			func(args []interface{}) (string, error) {
				rendered, err := k.CompleteConfig(args[2].(string), args[1].(string), args[0].(string), args[1].(string)).render()

				return string(rendered), err
			},
		).(pulumi.StringOutput)

	// payload[0] is internal IP
	case variables.InternalCommunicationMethod:
		internalIP := payload[0].(pulumi.StringOutput)
		config, _ = pulumi.All(k.leaderIP, con.IP, k.token, internalIP).ApplyT(
			func(args []interface{}) (string, error) {
				rendered, err := k.CompleteConfig(args[2].(string), args[3].(string), args[0].(string), args[1].(string)).render()

				return string(rendered), err
			},
		).(pulumi.StringOutput)
	}

	configure, err := k.configure(ctx, con, config, res)
	if err != nil {
		return nil, fmt.Errorf("error while configuring: %w", err)
	}

	res = append(res, configure...)

	var kubeconfig pulumi.AnyOutput
	if k.Sys.Leader() {
		kubeconfig, err = k.kubeconfig(ctx, con, res)
		if err != nil {
			return nil, fmt.Errorf("error while grabbing kubeconfig: %w", err)
		}
	}

	return &Provisioned{
		resources: res,
		Outputs: &Outputs{
			KubeconfigForUsage:  kubeconfig,
			KubeconfigForExport: kubeconfig,
		},
	}, nil
}

func (p *Provisioned) Value() interface{} {
	return p.Outputs
}

func (p *Provisioned) Resources() []pulumi.Resource {
	return p.resources
}

// CompleteConfig completes k3s config with pulumi.Outputs values.
func (k *K3S) CompleteConfig(token, ip, leaderIP, externalIP string) *CompletedConfig {
	k.Config.K3S.Token = token
	k.Config.K3S.FlannelIface = k.Sys.CommunicationIface()
	k.Config.K3S.NodeIP = ip
	k.Config.K3S.ExternalNodeIP = externalIP
	k.Config.K3S.Server = fmt.Sprintf("https://%s:6443", leaderIP)

	if k.role == variables.ServerRole {
		// Do not bind server API to specific ip.
		// Security is handled by firewall.
		// If user decides to use k3s without firewall or allow public access, then it is his own responsibility.
		// k.Config.K3S.BindAddress = ip
		k.Config.K3S.AdvertiseAddr = ip
		k.Config.K3S.TLSSan = externalIP
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
