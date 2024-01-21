package microos

import (
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (m *MicroOS) WaitForCloudInit(ctx *program.Context, con *connection.Connection) error {
	// There is always error
	cmd := "cloud-init status -l --wait 1>/dev/null || echo 'skip error since cloud-init status always returns error now. TO DO: see https://github.com/lima-vm/lima/issues/1496"

	opts := []pulumi.ResourceOption{
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5m", Update: "5m"}),
		pulumi.DependsOn(m.resources),
	}

	opts = append(opts, ctx.Options()...)

	installed, err := remote.NewCommand(ctx.Context(), fmt.Sprintf("wait-for-cloudinit-%s", m.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.String(cmd),
	}, opts...)
	if err != nil {
		return err
	}

	m.resources = append(m.resources, installed)

	return nil
}
