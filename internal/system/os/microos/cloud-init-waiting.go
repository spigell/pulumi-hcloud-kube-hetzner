package microos

import (
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (m *MicroOS) WaitForCloudInit(ctx *pulumi.Context, con *connection.Connection) error {
	// There is always error
	cmd := "cloud-init status -l --wait || true"

	installed, err := remote.NewCommand(ctx, fmt.Sprintf("wait-for-cloudinit-%s", m.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.String(cmd),
	},
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5m", Update: "5m"}),
		pulumi.DependsOn(m.resources),
	)
	if err != nil {
		return err
	}

	m.resources = append(m.resources, installed)

	return nil
}
