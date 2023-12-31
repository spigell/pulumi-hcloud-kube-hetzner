package os

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/sshd"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type OperationSystem interface {
	Up(*pulumi.Context, *hetzner.Server, map[string][]pulumi.Resource) (Provisioned, error)
	SetupWireguard(*wireguard.Config)
	SetupSSHD(*sshd.Config)
	AddK3SModule(string, *k3s.Config)
	Wireguard() *wireguard.Wireguard

	AddAdditionalRequiredPackages([]string)
	Modules() map[string]modules.Module
}

type Provisioned interface {
	Modules() map[string]modules.Output
}
