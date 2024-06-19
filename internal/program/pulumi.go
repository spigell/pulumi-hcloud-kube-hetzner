package program

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	remotefile "github.com/spigell/pulumi-file/sdk/go/file/remote"
	upgradev1 "github.com/spigell/pulumi-hcloud-kube-hetzner/crds/generated/rancher/upgrade/v1"
)

type resourceArgs interface {
	*hcloud.SshKeyArgs |
		*hcloud.ServerArgs |
		*hcloud.FirewallArgs |
		*hcloud.FirewallAttachmentArgs |
		*hcloud.NetworkSubnetArgs |
		*hcloud.NetworkArgs |
		*local.CommandArgs |
		*remotefile.FileArgs |
		*remote.CommandArgs |
		*tls.PrivateKeyArgs |
		*tls.LocallySignedCertArgs |
		*tls.SelfSignedCertArgs |
		*tls.CertRequestArgs |
		*kubernetes.ProviderArgs |
		*corev1.NodePatchArgs |
		*corev1.SecretArgs |
		*corev1.NamespaceArgs |
		*helmv3.ReleaseArgs |
		helmv3.ChartArgs |
		*upgradev1.PlanArgs
}

type resources interface {
	*hcloud.SshKey |
		*hcloud.Network |
		*hcloud.NetworkSubnet |
		*hcloud.Server |
		*hcloud.Firewall |
		*hcloud.FirewallAttachment |
		*local.Command |
		*remote.Command |
		*remotefile.File |
		*tls.CertRequest |
		*tls.PrivateKey |
		*tls.SelfSignedCert |
		*tls.LocallySignedCert |
		*kubernetes.Provider |
		*corev1.NodePatch |
		*corev1.Secret |
		*corev1.Namespace |
		*helmv3.Release |
		*helmv3.Chart |
		*upgradev1.Plan
}

// PulumiRun is a wrapper for all pulumi resources in the program.
// Additional pulumi options will be added to the global option array.
func PulumiRun[rArgs resourceArgs, pRes resources](
	ctx *Context,
	res func(*pulumi.Context, string, rArgs, ...pulumi.ResourceOption) (pRes, error),
	id string,
	args rArgs,
	additionalOptions ...pulumi.ResourceOption,
) (pRes, error) {
	name := fmt.Sprintf("%s:%s", ctx.ClusterName(), id)

	return res(ctx.Context(), name, args, append(ctx.Options(), additionalOptions...)...)
}
