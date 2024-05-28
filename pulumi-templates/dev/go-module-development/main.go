// This template is used mostly for local development.
// It is not possible to change configuration with pulumi.Output values when using the project as golang module.
// All dependencies must be resolved!

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	"gopkg.in/yaml.v3"
)

type clusters map[string]map[string]any

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

				cluster, err := phkh.NewCluster(ctx, name, cfg, []pulumi.ResourceOption{})
				if err != nil {
					return fmt.Errorf("error while cluster (id:%s) initialization: %w", name, err)
				}

				deployed, err := cluster.Up()
				if err != nil {
					return fmt.Errorf("error while cluster (id:%s) creation: %w", name, err)
				}

				outputs[name] = pulumi.Map{
					phkh.PrivatekeyKey:     deployed.Privatekey,
					phkh.HetznerServersKey: deployed.Servers,
					phkh.KubeconfigKey:     deployed.Kubeconfig,
				}
			}
		}
		ctx.Export(phkh.PhkhKey, outputs)

		return nil
	})
}
