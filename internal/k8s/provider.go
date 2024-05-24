package k8s

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

func (k *K8S) Provider(kubeconfig pulumi.AnyOutput, deps []pulumi.Resource) (*kubernetes.Provider, error) {
	return program.PulumiRun(k.ctx, kubernetes.NewProvider, "control-kubeconfig", &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig.ApplyT(func(s interface{}) string {
			kubeconfig := s.(*api.Config)

			k, _ := clientcmd.Write(*kubeconfig)

			return string(k)
		}).(pulumi.StringOutput),
	},
		pulumi.AdditionalSecretOutputs([]string{"stdout"}),
		pulumi.DependsOn(deps),
		// Ignore kubeconfig changes because it leads to recreation of all k8s resources.
		pulumi.IgnoreChanges([]string{"kubeconfig"}),
	)
}
