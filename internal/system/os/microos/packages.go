package microos

import (
	"fmt"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (m *MicroOS) Packages(ctx *pulumi.Context, con *connection.Connection) error {
	zypper := "zypper up -y"
	if len(m.RequiredPkgs) > 0 {
		zypper = fmt.Sprintf("%s && zypper install -y %s", zypper, strings.Join(m.RequiredPkgs, " "))
	}
	cmd := fmt.Sprintf(`sudo transactional-update -n run bash -c '%s'`, zypper)

	installed, err := remote.NewCommand(ctx, fmt.Sprintf("packages-%s", m.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.String(cmd),
	})
	if err != nil {
		return err
	}

	m.resources = append(m.resources, installed)

	return nil
}
