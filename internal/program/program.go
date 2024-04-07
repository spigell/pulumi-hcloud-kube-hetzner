package program

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

type Context struct {
	ctx   *pulumi.Context
	opts  []pulumi.ResourceOption
	state *State
}

func NewContext(ctx *pulumi.Context, state *State, opts ...pulumi.ResourceOption) *Context {
	return &Context{
		ctx:   ctx,
		opts:  opts,
		state: state,
	}
}

func (c *Context) Context() *pulumi.Context {
	return c.ctx
}

func (c *Context) Options() []pulumi.ResourceOption {
	return c.opts
}

func LoadStateFile(ctx *pulumi.Context) (*State, error) {
	var state *State
	path := fmt.Sprintf("%s-%s.yaml", stateFilePrefix, ctx.Stack())
	file, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{}, ErrNoStateFile
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if err := yaml.Unmarshal(file, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %w", err)
	}

	return state, nil
}

func (c *Context) DumpStateToFile(deps []pulumi.Resource) error {
	// Dump file with pulumi
	path := fmt.Sprintf("%s-%s.yaml", stateFilePrefix, c.Context().Stack())

	_, err := local.NewCommand(c.Context(), "store-state", &local.CommandArgs{
		Create: c.state.IPAM.InternalIPS.ToArrayMapOutput().ApplyT(
			func(m map[string][]any) (string, error) {
				for _, subnet := range c.state.IPAM.Subnets {
					for _, internalIP := range m[subnet.ID] {
						subnet.TakenIPS = append(subnet.TakenIPS, internalIP.(string))
					}
				}

				for _, subnet := range c.state.IPAM.Subnets {
					slices.Sort(subnet.TakenIPS)
					subnet.TakenIPS = slices.Compact(subnet.TakenIPS)
				}

				encoded, err := yaml.Marshal(c.state)
				if err != nil {
					return "", fmt.Errorf("failed to marshal state: %w", err)
				}

				return fmt.Sprintf("echo '%s' > %s", encoded, path), nil
			},
		).(pulumi.StringOutput),
	}, append(c.Options(), pulumi.DependsOn(deps))...)

	return err
}

func (c *Context) State() *State {
	return c.state
}
