package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	phkhlib "github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	phkh "github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner"
	"gopkg.in/yaml.v3"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		dir := "./clusters"

		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		outputs := make(pulumi.MapMap)

		for _, file := range files {
			if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
				filePath := filepath.Join(dir, file.Name())
				content, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}

				var cfg map[string]interface{}
				if err := yaml.Unmarshal(content, &cfg); err != nil {
					return err
				}

				name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

				cluster, err := phkh.NewCluster(ctx, name, &phkh.ClusterArgs{
					Config: pulumi.ToMap(cfg),
				})
				if err != nil {
					return fmt.Errorf("error while cluster (id:%s) initialization: %w", name, err)
				}

				outputs[name] = pulumi.Map{
					phkhlib.PrivatekeyKey:     cluster.Privatekey,
					phkhlib.HetznerServersKey: cluster.Servers,
					phkhlib.KubeconfigKey:     cluster.Kubeconfig,
				}
			}
		}

		ctx.Export(phkhlib.PhkhKey, outputs)

		return nil
	})

}
