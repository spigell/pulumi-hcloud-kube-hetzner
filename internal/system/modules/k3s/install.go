package k3s

import (
	"fmt"
	"path"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
)

const (
	k3sPath = "/usr/local/bin/k3s"
)

var installCommand = fmt.Sprintf(strings.Join([]string{
	// Check if initial install or upgrade.
	"sudo mkdir -p %s && if [[ -e %s ]]; then restart=true; fi",
	"curl -sfL https://get.k3s.io | sudo -E sh -x - 2>&1 >> /tmp/k3s-install-pulumi.log",
	"sudo systemctl daemon-reload",
	// Fix selinux context
	"sudo /sbin/restorecon -v %s",
	// If the old binary is installed then restart after upgrade.
	// Since the main installer will not restart it.
	"if [[ $restart ]]; then sudo systemctl restart k3s*; fi",
}, " && "), path.Dir(cfgPath), k3sPath, k3sPath)

func (k *K3S) install(ctx *program.Context, con *connection.Connection, deps []pulumi.Resource) (pulumi.Resource, error) {
	k3sExec := k.role

	installed, err := remote.NewCommand(ctx.Context(), fmt.Sprintf("install-k3s-binary-for-%s", k.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Environment: pulumi.StringMap{
			"INSTALL_K3S_SKIP_START":       pulumi.String("true"),
			"INSTALL_K3S_SKIP_ENABLE":      pulumi.String("true"),
			"INSTALL_K3S_SKIP_SELINUX_RPM": pulumi.String("true"),
			"INSTALL_K3S_VERSION":          pulumi.String(k.Config.Version),
			"INSTALL_K3S_EXEC":             pulumi.String(k3sExec),
		},
		Create: pulumi.String(installCommand),
		Delete: pulumi.String("/usr/local/bin/k3s-killall.sh"),
	},
		append(ctx.Options(),
			pulumi.DependsOn(deps),
			pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Delete: "10m"}),
			pulumi.RetainOnDelete(!k.Config.CleanDataOnUpgrade),
		)...)
	if err != nil {
		return nil, fmt.Errorf("error install a k3s cluster via script: %w", err)
	}

	return installed, nil
}
