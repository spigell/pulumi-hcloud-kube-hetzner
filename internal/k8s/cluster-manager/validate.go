package manager

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var AllowedTaintsEffects = []string{"NoSchedule", "NoExecute", "PreferNoSchedule"}

func (m *ClusterManager) ValidateNodePatches() error {
	if err := validateTaints(m.nodes); err != nil {
		return fmt.Errorf("failed to validate taints: %w", err)
	}

	return validateLabels(m.nodes)
}

func validateLabels(nodes map[string]*Node) error {
	for _, node := range nodes {
		if len(node.Labels) > 0 {
			for _, v := range node.Labels {
				if l := len(strings.Split(v, "=")); l == 1 {
					return fmt.Errorf("label must be in `key=value` format, got: %s", v)
				}
			}
		}
	}
	return nil
}

func validateTaints(nodes map[string]*Node) error {
	for _, node := range nodes {
		if len(node.Taints) > 0 {
			for _, taint := range node.Taints {
				keyValue := strings.Split(taint, ":")[0]
				if key := strings.Split(keyValue, "=")[0]; key == "" {
					return errors.New("empty taint key")
				}

				effect := strings.Split(taint, ":")[1]
				if effect == "" {
					return fmt.Errorf("missing taint effect: %s", effect)
				}

				if !slices.Contains(AllowedTaintsEffects, effect) {
					return fmt.Errorf("invalid taint effect: %s", effect)
				}
			}
		}
	}

	return nil
}
