package ccm

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

const (
	namespace = "kube-system"
	name      = "hcloud-cloud-controller-manager"
)

func (m *CCM) Manage(ctx *program.Context, prov *kubernetes.Provider, mgmt *manager.ClusterManager) error {
	token, err := m.discoverHcloudToken(ctx.Context())
	if err != nil {
		return fmt.Errorf("unable to discover hcloud token: %w", err)
	}

	secret, err := corev1.NewSecret(ctx.Context(), name, &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			// hcloud is hardcoded secretn name in ccm helm chart.
			Name:      pulumi.String("hcloud"),
			Namespace: pulumi.String(namespace),
		},
		StringData: pulumi.StringMap{
			"token": pulumi.String(token),
			// If networking is disabled it is doesn't used.
			// But it will be created anyway.
			"network": pulumi.Sprintf("%s-%s", ctx.Context().Project(), ctx.Context().Stack()),
		},
		//	}, append(ctx.Options(), pulumi.Provider(prov), pulumi.DependsOn(mgmt.Resources()))...)
	}, append(ctx.Options(), pulumi.Provider(prov))...)
	if err != nil {
		return fmt.Errorf("unable to create secret: %w", err)
	}

	_, err = helmv3.NewRelease(ctx.Context(), name, &helmv3.ReleaseArgs{
		Chart:     pulumi.String(name),
		Namespace: pulumi.String(namespace),
		Version:   pulumi.String(m.helm.Version),
		Name:      pulumi.String(name),
		RepositoryOpts: helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(HelmRepo),
		},
		ValueYamlFiles: m.helm.ValuesFiles(),
		Values: pulumi.Map{
			"args": pulumi.Map{
				"cloud-provider":       pulumi.String("hcloud"),
				"allow-untagged-cloud": pulumi.String(""),
				"controllers":          pulumi.String(strings.Join(m.controllers, ",")),
			},
			"networking": pulumi.Map{
				"enabled":     pulumi.Bool(m.networking),
				"clusterCIDR": pulumi.String(m.clusterCIDR),
			},
			"env": pulumi.Map{
				"HCLOUD_LOAD_BALANCERS_ENABLED": pulumi.Map{
					"value": pulumi.String(strconv.FormatBool(m.loadbalancersEnabled)),
				},
				"HCLOUD_LOAD_BALANCERS_LOCATION": pulumi.Map{
					"value": pulumi.String(m.defaultLoadbalancersLocation),
				},
				"HCLOUD_LOAD_BALANCERS_USE_PRIVATE_IP": pulumi.Map{
					"value": pulumi.String(strconv.FormatBool(m.loadbalancersUsePrivateIP)),
				},
			},
		},
	}, append(
		ctx.Options(),
		pulumi.Provider(prov),
		pulumi.DeleteBeforeReplace(true),
		pulumi.DependsOn([]pulumi.Resource{secret}),
	)...)
	if err != nil {
		return fmt.Errorf("unable to create helm release: %w", err)
	}

	return nil
}

func (m *CCM) discoverHcloudToken(ctx *pulumi.Context) (string, error) {
	cfg := config.New(ctx, "hcloud")
	tokenFromConfig := cfg.Get("token")

	switch {
	case m.token != "":
		return m.token, nil
	case tokenFromConfig != "":
		return tokenFromConfig, nil
	case os.Getenv("HCLOUD_TOKEN") != "":
		return os.Getenv("HCLOUD_TOKEN"), nil
	default:
		return "", fmt.Errorf("can't discover hcloud token via env or configs")
	}
}
