package microos

import (
	"fmt"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (m *MicroOS) Reboot(ctx *pulumi.Context, con *connection.Connection) error {
	rebooted, err := remote.NewCommand(ctx, fmt.Sprintf("reboot-%s", m.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		// Use very primitive way to reboot node.
		Create:   pulumi.String("(sleep 1 && sudo /sbin/shutdown -r now) &"),
		Triggers: utils.ExtractRemoteCommandResources(m.resources),
	}, pulumi.DependsOn(m.resources),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "2m"}),
	)
	if err != nil {
		return fmt.Errorf("error reboot node: %w", err)
	}

	m.resources = append(m.resources, rebooted)

	waitCommand := pulumi.Sprintf(strings.Join([]string{
		"go run ./scripts/ssh-uptime-checker/main.go %s %s",
	}, " && "), con.RemoteCommand().Host, con.User)

	waited, err := local.NewCommand(ctx, fmt.Sprintf("local-wait-%s", m.ID), &local.CommandArgs{
		Create: waitCommand,
		Environment: pulumi.StringMap{
			"CHECKER_SSH_PRIVATE_KEY": pulumi.ToSecret(pulumi.String(con.PrivateKey)).(pulumi.StringOutput),
		},
		Triggers: utils.ExtractRemoteCommandResources(m.resources),
	}, pulumi.DependsOn([]pulumi.Resource{rebooted}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m"}),
	)
	if err != nil {
		return fmt.Errorf("error waiting for node: %w", err)
	}
	m.resources = append(m.resources, waited)

	return nil
}
