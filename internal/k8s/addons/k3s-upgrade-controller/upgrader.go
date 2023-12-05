package upgrader

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/config/helm"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions"
)

const (
	enabledByDefault = false
	defaultChannel   = "stable"

	name             = "k3s-upgrade-controller"
	// ControlLabelKey is a label key to use for enabling k3s-upgrade-controller for specific node.
	ControlLabelKey  = "k3s-upgrade"
)

type Config struct {
	Enabled bool
	Helm    *helm.Config
	// Version is a version to use for the upgrade. Conflicts with Channel.
	TargetVersion string `json:"target-version" yaml:"target-version"`
	// Channel is a channel to use for the upgrade. Conflicts with Version.
	TargetChannel string `json:"target-channel" yaml:"target-channel"`
}

type Upgrader struct {
	enabled                      bool
	helm                         *helm.Config
	channel                      string
	version                      string
	serviceAccountName           string
}


func New(cfg *Config) *Upgrader {
	u := &Upgrader{}

	if cfg == nil {
		cfg = &Config{
			Enabled: enabledByDefault,
		}
	}

	if cfg.TargetChannel == "" && cfg.TargetVersion == "" {
		cfg.TargetChannel = defaultChannel
	}

	u.enabled = cfg.Enabled
	u.channel = cfg.TargetChannel
	u.version = cfg.TargetVersion
	// Hardcoded in the helm chart.
	u.serviceAccountName = "system-upgrade"

	return u
}

func (u *Upgrader) Helm() *helm.Config {
	return u.helm
}

func (u *Upgrader) SetHelm(h *helm.Config) {
	u.helm = h
}

func (u *Upgrader) Name() string {
	return name
}

func (u *Upgrader) Enabled() bool {
	return u.enabled
}

func (u *Upgrader) Supported(distr string) bool {
	switch distr {
	case distributions.K3SDistrName:
		return true
	default:
		return false
	}
}
