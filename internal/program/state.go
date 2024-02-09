package program

import (
	"errors"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network/ipam"
)

const (
	stateFilePrefix = "state"
)

// ErrNoStateFile is returned when the state file is not found.
var ErrNoStateFile = errors.New("no state file")

type State struct {
	IPAM *ipam.Data `yaml:",omitempty"`
}
