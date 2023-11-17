package addons

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/ccm"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/config/helm"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Addons struct {
	CCM *ccm.Config
}

type Addon interface {
	Name() string
	IsEnabled() bool
	Manage(*pulumi.Context, *kubernetes.Provider) error
	IsSupported(string) bool
	Helm() *helm.Config
	SetHelm(*helm.Config)
}

func New(addons *Addons) []Addon {
	a := []Addon{
		WithHelmInited(ccm.New(addons.CCM)),
	}

	return a
}

func Validate(a []Addon) error {
	for _, addon := range a {
		_, err := helm.GetDefaultVersion(addon.Name())
		if err != nil {
			return err
		}
	}

	return nil

}

func WithHelmInited(addon Addon) Addon {
	h := addon.Helm()

	if addon.Helm() == nil {
		h = &helm.Config{}
	}

	if h.Version == "" {
		defVer, _ := helm.GetDefaultVersion(addon.Name())

		h.Version = defVer

	}

	addon.SetHelm(h)

	return addon
}
