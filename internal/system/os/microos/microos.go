package microos

import (
	"sort"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/audit"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/sshd"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// Name of OS.
	Name = "microos"
	// sftp-server is preinstalled in microos based images.
	SFTPServerPath = "/usr/libexec/ssh/sftp-server"

	AfterReboot int = iota
	AfterNetwork
	SystemServices
)

type MicroOS struct {
	modules map[string]modules.Module
	// Temporary storage for resources from previous module.
	resources []pulumi.Resource

	ID           string
	RequiredPkgs []string
}

type Provisioned struct {
	modules map[string]modules.Output
}

func New(id string) *MicroOS {
	return &MicroOS{
		ID:      id,
		modules: make(map[string]modules.Module),
	}
}

func (m *MicroOS) AddAdditionalRequiredPackages(packages []string) {
	m.RequiredPkgs = append(m.RequiredPkgs, packages...)
}

func (m *MicroOS) Up(ctx *program.Context, server *hetzner.Server, kubeDependecies map[string][]pulumi.Resource) (os.Provisioned, error) {
	if err := m.WaitForCloudInit(ctx, server.Connection); err != nil {
		return nil, err
	}

	if err := m.Packages(ctx, server.Connection); err != nil {
		return nil, err
	}

	if err := m.Reboot(ctx, server.Connection); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(m.modules))
	for key := range m.modules {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return m.modules[keys[i]].Order() < m.modules[keys[j]].Order()
	})

	outputs := make(map[string]modules.Output)
	k3sPayload := make([]interface{}, 0)

	for _, k := range keys {
		var o modules.Output
		var err error

		// Always recreate deps because some modules require additional dependencies.
		// But other doesn't.
		deps := m.resources

		switch k {
		case variables.K3s:
			k3sPayload = append(k3sPayload, server.InternalIP)

			// All nodes must wait for leader to be ready.
			deps = append(deps, kubeDependecies["leader"]...)
			// Add own modules to dependencies.
			deps = append(deps, kubeDependecies[m.ID]...)

			o, err = m.modules[k].Up(ctx, server.Connection, deps, k3sPayload)
		default:
			o, err = m.modules[k].Up(ctx, server.Connection, deps, nil)
		}

		if err != nil {
			return nil, err
		}

		m.resources = o.Resources()
		outputs[k] = o
	}

	return &Provisioned{
		modules: outputs,
	}, nil
}

func (m *MicroOS) SFTPServerPath() string {
	return SFTPServerPath
}

func (m *MicroOS) SetupSSHD(config *sshd.Config) {
	module := sshd.New(m.ID, &MicroOS{}, config)

	module.SetOrder(AfterNetwork)
	m.modules[variables.SSHD] = module
}

func (m *MicroOS) SetupUpdate(config *updates.Config) {
	module := transUpdate.New(m.ID, &MicroOS{}, &Config{
		Schedule: config.Schedule,
		RebootMethod: config.Method,
	})

	module.SetOrder(AfterNetwork)
	m.modules[variables.SSHD] = module
}

func (m *MicroOS) AddK3SModule(role string, config *k3s.Config, auditLog *audit.AuditLog) {
	m.AddAdditionalRequiredPackages(k3s.GetRequiredPkgs(Name))
	module := k3s.New(m.ID, role, &MicroOS{}, config)

	if role == variables.ServerRole {
		module = module.WithK8SAuditLog(auditLog)
	}

	module.SetOrder(SystemServices)
	m.modules[variables.K3s] = module
}

func (m *MicroOS) Modules() map[string]modules.Module {
	return m.modules
}

func (p *Provisioned) Modules() map[string]modules.Output {
	return p.modules
}
