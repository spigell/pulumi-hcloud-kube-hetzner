package hetzner

import (
	"fmt"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (h *Hetzner) NewSSHKey(key pulumi.StringOutput) (*hcloud.SshKey, error) {
	sshPublicKey, err := hcloud.NewSshKey(h.ctx.Context(), "ssh-key", &hcloud.SshKeyArgs{
		Name:      pulumi.String(fmt.Sprintf("%s-%s", h.ctx.Context().Project(), h.ctx.Context().Stack())),
		PublicKey: key,
	}, h.ctx.Options()...)
	if err != nil {
		return nil, err
	}

	return sshPublicKey, nil
}
