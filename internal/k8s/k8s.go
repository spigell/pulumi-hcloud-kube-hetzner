package k8s

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
)

type K8S struct {
	ctx *pulumi.Context
}

func New(ctx *pulumi.Context) *K8S{
	return &K8S{
		ctx: ctx,
	}
}

func (k *K8S) Up(kubeconfig pulumi.AnyOutput) error {

	prov, err := kubernetes.NewProvider(k.ctx, "main", &kubernetes.ProviderArgs{
		// TO DO: Make it configurable
		DeleteUnreachable: pulumi.Bool(false),
		Kubeconfig: kubeconfig.ApplyT(func(s interface{}) string {
			kubeconfig := s.(*api.Config)

			k, _ := clientcmd.Write(*kubeconfig)

			return string(k)
		}).(pulumi.StringOutput),
	})

	if err != nil {
		return err
	}

	_, err = corev1.NewPod(k.ctx, "pod", &corev1.PodArgs{
		Spec: corev1.PodSpecArgs{
			Containers: corev1.ContainerArray{
				corev1.ContainerArgs{
					Name:  pulumi.String("nginx"),
					Image: pulumi.String("nginx"),
				},
			},
		}}, pulumi.Provider(prov))
	if err != nil {
		return err
	}

	return nil
}
