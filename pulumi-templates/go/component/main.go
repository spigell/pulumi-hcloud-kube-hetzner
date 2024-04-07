package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	phkhlib "github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	phkh "github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cluster, err := phkh.NewCluster(ctx, "my-cluster", &phkh.ClusterArgs{})
		if err != nil {
			return err
		}

		outputs := pulumi.Map{
			phkhlib.PrivatekeyKey: cluster.Privatekey,
			phkhlib.HetznerServersKey: cluster.Servers,
			phkhlib.KubeconfigKey: cluster.Kubeconfig,
		}

		ctx.Export(phkhlib.PhkhKey, outputs)

		return nil
	})
}
