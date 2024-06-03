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
	clusterName string
	ctx         *pulumi.Context
	opts        []pulumi.ResourceOption
	state       *State
}

func NewContext(ctx *pulumi.Context, name string, opts ...pulumi.ResourceOption) (*Context, error) {
	c := &Context{
		ctx:         ctx,
		clusterName: name,
		opts:        opts,
	}

	state, err := c.loadStateFile()
	if err != nil {
		if !errors.Is(err, ErrNoStateFile) {
			return nil, err
		}
	}

	c.state = state

	return c, nil
}

func (c *Context) Context() *pulumi.Context {
	return c.ctx
}

func (c *Context) Options() []pulumi.ResourceOption {
	return c.opts
}

func (c *Context) loadStateFile() (*State, error) {
	var state *State
	path := c.StatePath()
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

	_, err := PulumiRun(c, local.NewCommand, "store-state", &local.CommandArgs{
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

				return fmt.Sprintf("mkdir -p %s && echo '%s' > %s", stateDirectory, encoded, c.StatePath()), nil
			},
		).(pulumi.StringOutput),
	}, pulumi.DependsOn(deps))

	return err
}

func (c *Context) State() *State {
	return c.state
}

func (c *Context) ClusterName() string {
	return c.clusterName
}

func (c *Context) FullName() string {
	return fmt.Sprintf("%s-%s-%s", c.ctx.Project(), c.ctx.Stack(), c.ClusterName())
}

func (c *Context) StatePath() string {
	return fmt.Sprintf("%s/%s-%s.yaml", stateDirectory, stateFilePrefix, c.ClusterName())
}
