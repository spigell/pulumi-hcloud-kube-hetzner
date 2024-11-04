package talos

import (
	"encoding/json"
	"fmt"
	"strings"

	//"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"gopkg.in/yaml.v3"
)

var (
	masterIP = "65.109.4.153"
)

//
// TALOS_IMAGE="https://factory.talos.dev/image/1da3394e6229e507d4e3d166b718cacff86435a61c4765feedd66b43ac237558/v1.8.2/hcloud-amd64.raw.xz"
//  WGET="wget --timeout=5 --waitretry=5 --tries=5 --retry-connrefused --inet4-only"
//
//  apt-get install -y wget
//  $WGET -O /tmp/talos.raw.xz ${TALOS_IMAGE}
//  xz -d -c /tmp/talos.raw.xz | dd of=/dev/sda && sync
//  # Reboot
//  echo b > /proc/sysrq-trigger

func Provision(ctx *program.Context, servers map[string]*hetzner.Server) error {
	wget := "wget --timeout=5 --waitretry=5 --tries=5 --retry-connrefused --inet4-only"
	talosImage := "https://factory.talos.dev/image/613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245/v1.8.2/metal-amd64.raw.zst"

	command := pulumi.Sprintf(strings.Join([]string{
		"%s -O /tmp/talos.raw.xz %s",
		"zstd -d -c /tmp/talos.raw.xz | dd of=/dev/sda && sync",
		"reboot",
	}, " && "), wget, talosImage)

	for _, srv := range servers {
		_, err := program.PulumiRun(ctx, remote.NewCommand, fmt.Sprintf("deploy-talos:%s", "test"), &remote.CommandArgs{
			Connection: &remote.ConnectionArgs{
				User:       pulumi.String("root"),
				Host:       srv.Connection.IP,
				PrivateKey: srv.Connection.PrivateKey,
			},
			Create: command,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func Bootstrap(ctx *program.Context, servers map[string]*hetzner.Server) error {

	// talosVersion := "v1.8.2"
	secrets, err := machine.NewSecrets(ctx.Context(), "secrets", &machine.SecretsArgs{})
	if err != nil {
		return err
	}

	configuration := machine.GetConfigurationOutput(ctx.Context(), machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String("exampleCluster"),
		MachineType:     pulumi.String("controlplane"),
		ClusterEndpoint: pulumi.Sprintf("https://%s:6443", masterIP),
		MachineSecrets:  secrets.ToSecretsOutput().MachineSecrets(),
	}, nil)

	t := true

	config0 := v1alpha1.Config{
		MachineConfig: &v1alpha1.MachineConfig{
			MachineInstall: &v1alpha1.InstallConfig{
				InstallDisk: "/dev/sda",
			},
		},
		ClusterConfig: &v1alpha1.ClusterConfig{
			AllowSchedulingOnControlPlanes: &t,
		},
	}

	tmpJSON0, err := json.Marshal(map[string]interface{}{
		"machine": map[string]interface{}{
			"install": map[string]interface{}{
				"disk": "/dev/sda",
			},
		},
		"cluster": map[string]interface{}{
			"allowSchedulingOnControlPlanes": true,
		},
	})

	json1, _ := yaml.Marshal(config0)
	json0, _ := json.Marshal(tmpJSON0)

	fmt.Println(string(json0))
	fmt.Println(string(json1))

	for _, srv := range servers {

		json0 := string(json1)
		configurationApply, err := machine.NewConfigurationApply(ctx.Context(), "configurationApply", &machine.ConfigurationApplyArgs{
			ClientConfiguration:       secrets.ClientConfiguration,
			MachineConfigurationInput: configuration.MachineConfiguration(),
			Node:                      srv.Connection.IP,
			ConfigPatches: pulumi.StringArray{
				pulumi.String(json0),
			},
		})
		if err != nil {
			return err
		}
		_, err = machine.NewBootstrap(ctx.Context(), "bootstrap", &machine.BootstrapArgs{
			Node:                srv.Connection.IP,
			ClientConfiguration: secrets.ClientConfiguration,
		}, pulumi.DependsOn([]pulumi.Resource{
			configurationApply,
		}))

		_ = cluster.GetHealthOutput(ctx.Context(), cluster.GetHealthOutputArgs{
			ClientConfiguration: cluster.GetHealthClientConfigurationArgs{
				CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
				ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
				ClientKey:         secrets.ClientConfiguration.ClientKey(),
			},
			ControlPlaneNodes: pulumi.StringArray{
				srv.Connection.IP,
			},
		})
		if err != nil {
			return err
		}

		kube := cluster.GetKubeconfigOutput(ctx.Context(), cluster.GetKubeconfigOutputArgs{
			ClientConfiguration: cluster.GetKubeconfigClientConfigurationArgs{
				CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
				ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
				ClientKey:         secrets.ClientConfiguration.ClientKey(),
			},
			Node: srv.Connection.IP,
		})

		ctx.Context().Export("kube", kube.KubeconfigRaw())
	}

	return nil
}

/*mport * as talos from "@pulumiverse/talos";

const masterIP = "192.168.1.10"
const talosVersion = "v1.7.5"

const secrets = new talos.machine.Secrets("secrets", {
	talosVersion: talosVersion,
});

const configuration = talos.machine.getConfigurationOutput({
    clusterName: "main",
    talosVersion: talosVersion,
    machineType: "controlplane",
    clusterEndpoint: `https://${masterIP}:6443`,
    machineSecrets: secrets.machineSecrets,
});

const configurationApply = new talos.machine.ConfigurationApply("configurationApply", {
    clientConfiguration: secrets.clientConfiguration,
    machineConfigurationInput: configuration.machineConfiguration,
    node: masterIP,
    configPatches: [JSON.stringify({
        machine: {
        	network: {
        		hostname: "my-machine-01"
        	},
            install: {
            	image: `ghcr.io/siderolabs/installer:${talosVersion}`,
                disk: "/dev/nvme0n1",
//	            extraKernelArgs: [
//	            	"ipv6.disable=1",
//		        ],
            },
            kubelet: {
//            	extraArgs: {
//	            	"rotate-server-certificates": true
//            	}
            },
        },
        cluster: {
          allowSchedulingOnControlPlanes: true,
		  network: {
		    cni: {
		      name: 'flannel'
		    }
		  }
        }
    })],
});

const bootstrap = new talos.machine.Bootstrap("bootstrap", {
    node: masterIP,
    clientConfiguration: secrets.clientConfiguration,
}, {
    dependsOn: [configurationApply],
});

const client = talos.client.getConfigurationOutput({
    clusterName: "example-cluster",
    clientConfiguration: secrets.clientConfiguration,
    nodes: [masterIP],
    endpoints: [masterIP]
});

const cluster = talos.cluster.getKubeconfigOutput({
    clientConfiguration: secrets.clientConfiguration,
    node: masterIP,
});

export const talosconfig = client.talosConfig
export const kubeconfig = cluster.kubeconfigRaw
*/
