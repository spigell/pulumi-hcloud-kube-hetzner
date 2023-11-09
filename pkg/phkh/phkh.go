package phkh

import (
	"github.com/sanity-io/litter"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
)

type PHKH struct {
	config   *config.Config
	compiled *Compiled
	state    *State
}

func New(ctx *pulumi.Context) (*PHKH, error) {
	cfg := config.New(ctx)
	state, err := state(ctx)
	if err != nil {
		return nil, err
	}

	keys, err := state.sshKeyPair()
	if err != nil {
		return nil, err
	}

	token, err := state.k3sToken()
	if err != nil {
		return nil, err
	}

	compiled, err := compile(ctx, token, cfg, keys)
	if err != nil {
		return nil, err
	}

	return &PHKH{
		config:   cfg,
		compiled: compiled,
		state:    state,
	}, nil
}

func (c *PHKH) Up(ctx *pulumi.Context) error {
	hetznerInfo, err := c.state.hetznerInfra()
	if err != nil {
		return err
	}

	keys, err := c.state.sshKeyPair()
	if err != nil {
		return err
	}

	wgInfo, err := c.state.wgInfo()
	if err != nil {
		return err
	}

	cloud, err := c.compiled.Hetzner.Up(hetznerInfo, keys)
	if err != nil {
		return err
	}
	sys, err := c.compiled.SysCluster.Up(wgInfo, cloud)
	if err != nil {
		return err
	}

	//

	prov, err := kubernetes.NewProvider(ctx, "test", &kubernetes.ProviderArgs{
		// Make it configurable
		DeleteUnreachable: pulumi.Bool(false),
		Kubeconfig: sys.K3s.Kubeconfig.ApplyT(func(s interface{}) string {
			kubeconfig := s.(*api.Config)

			k, _ := clientcmd.Write(*kubeconfig)

			return string(k)
		}).(pulumi.StringOutput),
	})

	if err != nil {
		return err
	}

	_, err = corev1.NewPod(ctx, "pod", &corev1.PodArgs{
		Spec: corev1.PodSpecArgs{
			Containers: corev1.ContainerArray{
				corev1.ContainerArgs{
					Name:  pulumi.String("nginx"),
					Image: pulumi.String("nginx"),
				},
			},
		}}, pulumi.Provider(prov), pulumi.RetainOnDelete(true))
	if err != nil {
		return err
	}

	c.state.exportHetznerInfra(cloud)
	c.state.exportSSHKeyPair(keys)
	c.state.exportWGInfo(sys.Wireguard)
	c.state.exportK3SToken(sys.K3s.Token)
	c.state.exportK3SKubeconfig(sys.K3s.Kubeconfig)

	return nil
}

// DumpConfig returns a string representation of the parsed config.
// This is useful for debugging.
func (c *PHKH) DumpConfig() string {
	return litter.Sdump(c.config)
}
