package hetzner

import (
	"fmt"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

func (h *Hetzner) NewSSHKey(key pulumi.StringOutput) (*hcloud.SshKey, error) {
	cloudName := pulumi.String(fmt.Sprintf("%s-%s-%s",
		h.ctx.Context().Project(),
		h.ctx.Context().Stack(),
		h.ctx.ClusterName(),
	))

	sshPublicKey, err := program.PulumiRun(h.ctx, hcloud.NewSshKey, "ssh-key", &hcloud.SshKeyArgs{
		Name:      cloudName,
		PublicKey: key,
	})
	if err != nil {
		return nil, err
	}

	return sshPublicKey, nil
}
