package k3supgrader

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/k8sconfig/helm"
)

const (
	enabledByDefault = false
	defaultChannel   = "stable"

	Name = "k3s-upgrade-controller"
	// ControlLabelKey is a label key to use for enabling k3s-upgrade-controller for specific node.
	ControlLabelKey = "k3s-upgrade"
)

type Config struct {
	Enabled bool
	Helm    *helm.Config
	// Version is a version to use for the upgrade. Conflicts with Channel.
	TargetVersion string `json:"target-version" yaml:"target-version"`
	// Channel is a channel to use for the upgrade. Conflicts with Version.
	TargetChannel string `json:"target-channel" yaml:"target-channel"`
	// ConfigEnv is a map of environment variables to pass to the controller.
	ConfigEnv []string `json:"config-env" yaml:"config-env"`
}

type Upgrader struct {
	enabled            bool
	helm               *helm.Config
	channel            string
	version            string
	serviceAccountName string
	configEnv          []string
}

func New(cfg *Config) *Upgrader {
	u := &Upgrader{}

	if cfg == nil {
		cfg = &Config{
			Enabled:   enabledByDefault,
			ConfigEnv: make([]string, 0),
		}
	}

	if cfg.TargetChannel == "" && cfg.TargetVersion == "" {
		cfg.TargetChannel = defaultChannel
	}

	u.helm = cfg.Helm
	u.enabled = cfg.Enabled
	u.channel = cfg.TargetChannel
	u.version = cfg.TargetVersion
	u.configEnv = cfg.ConfigEnv
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
	return Name
}

func (u *Upgrader) Enabled() bool {
	return u.enabled
}

func (u *Upgrader) Version() string {
	return u.version
}

func (u *Upgrader) Supported(distr string) bool {
	switch distr {
	case distributions.K3SDistrName:
		return true
	default:
		return false
	}
}
