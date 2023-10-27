package modules

import (
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Module interface {
	SetOrder(int)
	Order() int
	Up(*pulumi.Context, *connection.Connection, []pulumi.Resource) (Output, error)
}

type Output interface {
	Value() interface{}
	Resources() []pulumi.Resource
}
