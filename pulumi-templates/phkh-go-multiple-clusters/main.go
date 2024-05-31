package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	phkhlib "github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	phkh "github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner/cluster"
)

type clusters map[string]map[string]any

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		small, err := phkh.NewCluster(ctx, "small", &phkh.ClusterArgs{
			Config: cluster.ConfigConfigArgs{
				Network: &cluster.ConfigNetworkConfigArgs{
					Hetzner: &cluster.NetworkConfigArgs{
						Enabled: pulumi.Bool(true),
					},
				},
				Nodepools: &cluster.ConfigNodepoolsConfigArgs{
					Servers: &cluster.ConfigNodepoolConfigArray{
						&cluster.ConfigNodepoolConfigArgs{
							PoolID: pulumi.String("servers"),
							Nodes: cluster.ConfigNodeConfigArray{
								&cluster.ConfigNodeConfigArgs{
									NodeID: pulumi.String("server-01"),
								},
							},
						},
					},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("error while the small cluster initialization: %w", err)
		}

		outputs := make(pulumi.MapMap, 0)

		name := "cluster2"

		cluster2, err := phkh.NewCluster(ctx, name, &phkh.ClusterArgs{
			Config: cluster.ConfigConfigArgs{
				Network: &cluster.ConfigNetworkConfigArgs{
					Hetzner: &cluster.NetworkConfigArgs{
						Enabled: pulumi.Bool(true),
					},
				},
				Nodepools: &cluster.ConfigNodepoolsConfigArgs{
					Servers: &cluster.ConfigNodepoolConfigArray{
						&cluster.ConfigNodepoolConfigArgs{
							PoolID: small.Config.Nodepools().Servers().Index(pulumi.Int(0)).PoolID().Elem(),
							Nodes: cluster.ConfigNodeConfigArray{
								&cluster.ConfigNodeConfigArgs{
									NodeID: pulumi.String("server-01"),
									Server: &cluster.ConfigServerConfigArgs{
										Firewall: &cluster.ConfigFirewallConfigArgs{
											Hetzner: &cluster.FirewallConfigArgs{
												Enabled: pulumi.Bool(true),
												AdditionalRules: &cluster.FirewallRuleConfigArray{
													&cluster.FirewallRuleConfigArgs{
														Port:        pulumi.String("50082"),
														Description: pulumi.String("test"),
														SourceIps: pulumi.StringArray{
															pulumi.Sprintf("%s/32", small.Servers.Index(pulumi.Int(0)).Ip().Elem()),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("error while cluster (id:%s) initialization: %w", name, err)
		}

		outputs[name] = pulumi.Map{
			phkhlib.PrivatekeyKey:     cluster2.Privatekey,
			phkhlib.HetznerServersKey: cluster2.Servers,
			phkhlib.KubeconfigKey:     cluster2.Kubeconfig,
			"config":                  cluster2.Config,
		}

		outputs["small"] = pulumi.Map{
			phkhlib.PrivatekeyKey:     small.Privatekey,
			phkhlib.HetznerServersKey: small.Servers,
			phkhlib.KubeconfigKey:     small.Kubeconfig,
			"config":                  small.Config,
		}
		ctx.Export(phkhlib.PhkhKey, outputs)

		return nil
	})
}
