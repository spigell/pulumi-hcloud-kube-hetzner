package microos

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (m *MicroOS) Reboot(ctx *program.Context, con *connection.Connection) error {
	rebooted, err := remote.NewCommand(ctx.Context(), fmt.Sprintf("reboot-%s", m.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		// Use very primitive way to reboot node.
		Create:   pulumi.String("(sleep 1 && sudo /sbin/shutdown -r now) &"),
		Triggers: utils.ExtractRemoteCommandResources(m.resources),
	}, append(ctx.Options(), pulumi.DependsOn(m.resources),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "2m"}),
	)...)
	if err != nil {
		return fmt.Errorf("error reboot node: %w", err)
	}

	m.resources = append(m.resources, rebooted)

	rebootCheckerDir := "tmp/reboot-checker"
	rebootCheckerBinaryPath := filepath.Join(rebootCheckerDir, fmt.Sprintf("reboot-checker-for-%s", m.ID))

	waitCommand := pulumi.Sprintf(strings.Join([]string{
		"mkdir -p %s",
		"curl -L -v -o %s https://github.com/spigell/pulumi-hcloud-kube-hetzner/releases/download/v0.0.3/reboot-checker-v0.0.3-%s-%s",
		"chmod +x %s",
		"%s %s %s",
	}, " && "),
		rebootCheckerDir,
		rebootCheckerBinaryPath,
		runtime.GOOS,
		runtime.GOARCH,
		rebootCheckerBinaryPath,
		rebootCheckerBinaryPath,
		con.RemoteCommand().Host,
		con.User,
	)

	waited, err := local.NewCommand(ctx.Context(), fmt.Sprintf("local-wait-for-%s", m.ID), &local.CommandArgs{
		Create: waitCommand,
		Update: pulumi.String("echo 'checker already used before. Skipping...'"),
		Environment: pulumi.StringMap{
			"CHECKER_SSH_PRIVATE_KEY": pulumi.ToSecret(con.PrivateKey).(pulumi.StringOutput),
		},
		Triggers: utils.ExtractRemoteCommandResources(m.resources),
	}, append(ctx.Options(), pulumi.DependsOn([]pulumi.Resource{rebooted}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m"}),
	)...)
	if err != nil {
		return fmt.Errorf("error waiting for node: %w", err)
	}
	m.resources = append(m.resources, waited)

	return nil
}
