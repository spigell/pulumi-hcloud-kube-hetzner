package k8s

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (k *K8S) Provider(kubeconfig pulumi.AnyOutput, deps []pulumi.Resource) (*kubernetes.Provider, error) {
	return kubernetes.NewProvider(k.ctx.Context(), "main", &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig.ApplyT(func(s interface{}) string {
			kubeconfig := s.(*api.Config)

			k, _ := clientcmd.Write(*kubeconfig)

			return string(k)
		}).(pulumi.StringOutput),
	}, append(
		k.ctx.Options(),
		pulumi.AdditionalSecretOutputs([]string{"stdout"}),
		pulumi.DependsOn(deps),
		// Ignore kubeconfig changes because it leads to recreation of all k8s resources.
		pulumi.IgnoreChanges([]string{"kubeconfig"}),
	)...)
}
