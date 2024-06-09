package os

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/audit"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/journald"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/sshd"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type OperatingSystem interface {
	Up(*program.Context, *hetzner.Server, map[string][]pulumi.Resource) (Provisioned, error)
	SetupSSHD(*sshd.Params)
	SetupJournalD(*journald.Config)
	AddK3SModule(string, *k3s.Config, *audit.AuditLog)

	AddAdditionalRequiredPackages([]string)
	Modules() map[string]modules.Module
}

type Provisioned interface {
	Modules() map[string]modules.Output
}
