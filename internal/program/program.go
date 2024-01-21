package program

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Context struct {
	ctx  *pulumi.Context
	opts []pulumi.ResourceOption
}

func NewContext(ctx *pulumi.Context, opts ...pulumi.ResourceOption) *Context {
	return &Context{
		ctx:  ctx,
		opts: opts,
	}
}

func (c *Context) Context() *pulumi.Context {
	return c.ctx
}

func (c *Context) Options() []pulumi.ResourceOption {
	return c.opts
}
