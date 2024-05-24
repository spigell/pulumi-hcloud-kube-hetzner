package storage

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

type Storage struct {
	oneTime bool
	opts    []pulumi.ResourceOption
	output  *pulumi.StringOutput

	Name    string
	Payload any
}

func New(name string, payload any) *Storage {
	return &Storage{
		Name:    name,
		Payload: payload,
		opts:    make([]pulumi.ResourceOption, 0),
	}
}

func (s *Storage) WithOneShot() *Storage {
	s.oneTime = true

	return s
}

func (s *Storage) WithPulumiOpts(opts []pulumi.ResourceOption) *Storage {
	s.opts = append(s.opts, opts...)

	return s
}

func (s *Storage) Store(ctx *program.Context) error {
	cmd, _ := json.MarshalIndent(s.Payload, "  ", "  ")
	encoded := base64.StdEncoding.EncodeToString(cmd)

	s.opts = append(s.opts, pulumi.AdditionalSecretOutputs([]string{"stdout"}))

	if s.oneTime {
		s.opts = append(s.opts, pulumi.IgnoreChanges([]string{"create"}))
	}

	out, err := program.PulumiRun(ctx, local.NewCommand, s.Name, &local.CommandArgs{
		Create: pulumi.ToSecret(pulumi.Sprintf("echo %s", encoded)).(pulumi.StringOutput),
	}, s.opts...)
	if err != nil {
		return err
	}

	s.output = &out.Stdout

	return nil
}

func (s *Storage) Get() pulumi.StringOutput {
	return s.output.ApplyT(func(keys string) (string, error) {
		decoded, err := base64.StdEncoding.DecodeString(keys)
		if err != nil {
			return "", nil
		}

		return string(decoded), nil
	}).(pulumi.StringOutput)
}
