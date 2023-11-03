package os

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/sshd"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type OperationSystem interface {
	Up(*pulumi.Context, *hetzner.Server) (Provisioned, error)
	SetupWireguard(*config.Wireguard)
	SetupSSHD(*sshd.Config)
	Wireguard() *wireguard.Wireguard

	AddAdditionalRequiredPackages([]string)
}

type Provisioned interface {
	Modules() map[string]modules.Output
}
