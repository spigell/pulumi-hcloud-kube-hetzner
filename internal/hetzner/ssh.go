package hetzner

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"
)

func (h *Hetzner) NewSSHKey(keys *pulumi.StringOutput) (*hcloud.SshKey, error) {
	sshPublicKey, err := hcloud.NewSshKey(h.ctx, "ssh-key", &hcloud.SshKeyArgs{
		Name: pulumi.String(fmt.Sprintf("%s-%s", h.ctx.Project(), h.ctx.Stack())),
		PublicKey: keys.ApplyT(func(keys string) string {
			var keypair *keypair.ECDSAKeyPair
			_ = json.Unmarshal([]byte(keys), &keypair)

			return keypair.PublicKey
		}).(pulumi.StringOutput),
	}, h.pulumiOpts...)

	if err != nil {
		return nil, err
	}

	return sshPublicKey, nil
}
