package k3stoken

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
)

type Token struct {
	storage *storage.Storage
}

func New(ctx *program.Context) (*Token, error) {
	token := utils.GenerateRandomString(48)

	storage := storage.New("store-generated-k3s-token", token).WithOneShot().WithPulumiOpts(ctx.Options())
	storage.Store(ctx)

	return &Token{
		storage: storage,
	}, nil
}

func (s *Token) Value() pulumi.StringOutput {
	return s.storage.Get()
}
