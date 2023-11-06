package k3s

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	remotefile "github.com/spigell/pulumi-file/sdk/go/file/remote"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
)

var ()

func (k *K3S) configure(ctx *pulumi.Context, con *connection.Connection, config pulumi.StringOutput, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	svcName := "k3s"
	if k.role == variables.AgentRole {
		svcName = "k3s-agent"
	}

	result := make([]pulumi.Resource, 0)

	deployed, err := remotefile.NewFile(ctx, fmt.Sprintf("k3s-configure-%s", k.ID), &remotefile.FileArgs{
		Connection: con.RemoteFile(),
		UseSudo:    pulumi.Bool(true),
		Path:       pulumi.String(cfgPath),
		Content:    config,
		SftpPath:   pulumi.String(k.OS.SFTPServerPath()),
	}, pulumi.DependsOn(deps), pulumi.RetainOnDelete(true))
	if err != nil {
		return nil, err
	}

	result = append(result, deployed)

	// Flannel iface is based on kubewg0 iface directly (wg mode), so flannel.0 disapered after wg restart.
	// We need to maintain k3s restart with wireguard network interface.
	// K3S config based on WG config, but sometimes it is not enough to restart k3s service because config can be the same.
	// So, we need to restart it manually somehow.
	// The main reason of this is find our dependencies and build trigger array for only them.
	// But set dependencies for both our deps and leader.
	triggers := pulumi.Array{
		deployed.Content,
		deployed.Md5sum,
	}
	for _, dep := range deps {
		if dep == nil {
			continue
		}
		c, ok := dep.(*remote.Command)
		if !ok {
			continue
		}
		t := pulumi.All(con.RemoteCommand().Host, c.Connection).ApplyT(
			func(args []interface{}) interface{} {
				// If it is our deps then add to the trigger slice.
				if args[1].(remote.Connection).Host == args[0].(string) {
					return c.Connection
				}
				// We need to return smth to make trigger work.
				return args[0].(string)
			}).(pulumi.AnyOutput)

		triggers = append(triggers, t)
	}

	// Restart k3s service.
	restartCommand := pulumi.Sprintf(strings.Join([]string{
		"sudo systemctl disable --now %s",
		"sudo systemctl enable --now %s",
		"sudo systemctl status %s",
		"echo 'systemctl status command returned' $? exit code",
	}, " && "), svcName, svcName, svcName)

	restared, err := remote.NewCommand(ctx, fmt.Sprintf("k3s-restart-%s", k.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     restartCommand,
		Triggers:   triggers,
	}, pulumi.DependsOn(result),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m"}),
		pulumi.DeleteBeforeReplace(false),
	)
	if err != nil {
		return nil, err
	}

	result = append(result, restared)

	return result, nil
}
