package sshkeypair

import (
	"encoding/json"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"
)

type KeyPair struct {
	storage *storage.Storage
}

func New(ctx *program.Context) (*KeyPair, error) {
	created, err := keypair.NewECDSA()
	if err != nil {
		return nil, err
	}

	storage := storage.New("store-generated-ssh-keypair", created).WithOneShot().WithPulumiOpts(ctx.Options())
	storage.Store(ctx)

	return &KeyPair{
		storage: storage,
	}, nil
}

func (s *KeyPair) PublicKey() pulumi.StringOutput {
	return s.storage.Get().ApplyT(func(value string) (string, error) {
		pair, err := unmarshal(value)

		if err != nil {
			return "", err
		}

		return pair.PublicKey, nil
	}).(pulumi.StringOutput)
}

func (s *KeyPair) PrivateKey() pulumi.StringOutput {
	return s.storage.Get().ApplyT(func(value string) (string, error) {
		pair, err := unmarshal(value)

		if err != nil {
			return "", err
		}

		return pair.PrivateKey, nil
	}).(pulumi.StringOutput)
}

func unmarshal(value string) (*keypair.ECDSAKeyPair, error) {
	var keypair *keypair.ECDSAKeyPair

	err := json.Unmarshal([]byte(value), &keypair)
	if err != nil {
		return nil, err
	}

	return keypair, nil
}
