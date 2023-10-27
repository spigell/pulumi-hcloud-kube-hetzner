package os

import (
	"pulumi-hcloud-kube-hetzner/internal/config"
	"pulumi-hcloud-kube-hetzner/internal/hetzner"
	"pulumi-hcloud-kube-hetzner/internal/system/modules"
	"pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type OperationSystem interface {
	Up(*pulumi.Context, *hetzner.Server) (Provisioned, error)
	SetWireguard(*config.Wireguard)
	Wireguard() *wireguard.Wireguard

	AddAdditionalRequiredPackages([]string)
}

type Provisioned interface {
	Modules() map[string]modules.Output
}
