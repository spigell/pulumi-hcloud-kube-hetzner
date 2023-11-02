package microos

import (
	"sort"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type MicroOS struct {
	modules   map[string]modules.Module
	resources []pulumi.Resource

	ID           string
	RequiredPkgs []string
}

func New(id string) *MicroOS {
	return &MicroOS{
		ID:      id,
		modules: make(map[string]modules.Module),
	}
}

type Provisioned struct {
	modules map[string]modules.Output
}

func (m *MicroOS) AddAdditionalRequiredPackages(packages []string) {
	m.RequiredPkgs = append(m.RequiredPkgs, packages...)
}

func (m *MicroOS) Up(ctx *pulumi.Context, server *hetzner.Server) (os.Provisioned, error) {
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
	for _, k := range keys {
		o, err := m.modules[k].Up(ctx, server.Connection, m.resources)
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

func (m *MicroOS) SetWireguard(config *config.Wireguard) {
	m.AddAdditionalRequiredPackages(wireguard.GetRequiredPkgs("microos"))

	module := wireguard.New(m.ID, config)
	module.SetOrder(1)
	m.modules["wireguard"] = module
}

func (m *MicroOS) Wireguard() *wireguard.Wireguard {
	if m.modules["wireguard"] == nil {
		return nil
	}

	return m.modules["wireguard"].(*wireguard.Wireguard)
}

func (p *Provisioned) Modules() map[string]modules.Output {
	return p.modules
}
