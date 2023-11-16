package addons

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/mcc"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Addons struct {
	MCC *mcc.Config
}

type Addon interface {
	Name() string
	IsEnabled() bool
	Manage(*pulumi.Context, *kubernetes.Provider) error
	IsSupported(string) bool
}

func New(addons *Addons) []Addon {
	return []Addon{
		mcc.New(addons.MCC),
	}
}
