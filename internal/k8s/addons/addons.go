package addons

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/ccm"
	k3supgrader "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/addons/k3s-upgrade-controller"
	manager "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/cluster-manager"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/config/helm"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Addons struct {
	CCM               *ccm.Config
	K3SSystemUpgrader *k3supgrader.Config `json:"k3s-upgrade-controller" yaml:"k3s-upgrade-controller"`
}

type Addon interface {
	Name() string
	Enabled() bool
	Manage(*program.Context, *kubernetes.Provider, *manager.ClusterManager) error
	Supported(string) bool
	Helm() *helm.Config
	SetHelm(*helm.Config)
}

func New(addons *Addons) []Addon {
	a := []Addon{
		WithHelmInited(ccm.New(addons.CCM)),
		WithHelmInited(k3supgrader.New(addons.K3SSystemUpgrader)),
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

	if h == nil {
		h = &helm.Config{}
	}

	if h.Version == "" {
		defVer, _ := helm.GetDefaultVersion(addon.Name())

		h.Version = defVer
	}

	if len(h.ValuesFilePath) > 0 {
		var assets pulumi.AssetOrArchiveArray
		for _, asset := range h.ValuesFilePath {
			assets = append(assets, pulumi.NewFileAsset(asset))
		}
		h.ValuesFiles = assets
	}

	addon.SetHelm(h)

	return addon
}
