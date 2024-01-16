package hetzner

import (
	"encoding/json"
	"encoding/base64"
	"fmt"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"
)

func (h *Hetzner) NewSSHKey(keys *pulumi.StringOutput) (*hcloud.SshKey, error) {
	sshPublicKey, err := hcloud.NewSshKey(h.ctx, "ssh-key", &hcloud.SshKeyArgs{
		Name: pulumi.String(fmt.Sprintf("%s-%s", h.ctx.Project(), h.ctx.Stack())),
		PublicKey: keys.ApplyT(func(keys string) (string, error) {
			var keypair *keypair.ECDSAKeyPair

			decoded, err := base64.StdEncoding.DecodeString(keys)
			if err != nil {
				return "", nil
			}

			err = json.Unmarshal([]byte(decoded), &keypair)
			if err != nil {
				return "", nil
			}

			return keypair.PublicKey, nil
		}).(pulumi.StringOutput),
	}, h.pulumiOpts...)

	if err != nil {
		return nil, err
	}

	return sshPublicKey, nil
}
