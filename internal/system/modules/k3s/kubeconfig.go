package k3s

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func (k *K3S) kubeconfig(ctx *pulumi.Context, con *connection.Connection, deps []pulumi.Resource) (pulumi.AnyOutput, error) {
	grabbed, err := remote.NewCommand(ctx, fmt.Sprintf("kubeconfig-%s", k.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.String("sudo cat /etc/rancher/k3s/k3s.yaml"),
	}, pulumi.DependsOn(deps))

	if err != nil {
		return pulumi.AnyOutput{}, fmt.Errorf("error grab kubeconfig: %w", err)
	}

	kube := grabbed.Stdout.ApplyT(func(v interface{}) (*api.Config, error) {
		stdout := v.(string)

		kubeconfig, err := clientcmd.Load([]byte(stdout))
		if err != nil {
			return nil, fmt.Errorf("error parse kubeconfig: %w", err)
		}

		ctxName := fmt.Sprintf("%s-direct", ctx.Stack())

		kubeconfig.Contexts[ctxName] = kubeconfig.Contexts["default"]
		delete(kubeconfig.Contexts, "default")
		kubeconfig.CurrentContext = ctxName

		return kubeconfig, nil

	}).(pulumi.AnyOutput)

	return kube, nil
}
