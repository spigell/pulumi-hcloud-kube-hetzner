package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"gopkg.in/yaml.v3"
)

type Example struct {
	Name    string
	Decoded *config.Config
}

func DiscoverExample(path string) (*Example, error) {
	var decoded *config.Config
	example := &Example{
		Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err = yaml.Unmarshal(content, &decoded); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	example.Decoded = decoded

	return example, nil
}

func (e *Example) NodesIDs() []string {
	var ids []string

	for _, pool := range e.Decoded.Nodepools.Agents {
		for _, n := range pool.Nodes {
			ids = append(ids, n.NodeID)
		}
	}

	for _, pool := range e.Decoded.Nodepools.Servers {
		for _, n := range pool.Nodes {
			ids = append(ids, n.NodeID)
		}
	}

	return ids
}

func (e *Example) UniqConfigsByNodes() map[string]*config.NodeConfig {
	configs := make(map[string]*config.NodeConfig)

	for _, pool := range e.Decoded.Nodepools.Agents {
		for _, n := range pool.Nodes {
			if n.Server != nil || n.K3s != nil {
				configs[n.NodeID] = n
			}
		}
	}

	for _, pool := range e.Decoded.Nodepools.Servers {
		for _, n := range pool.Nodes {
			if n.Server != nil || n.K3s != nil {
				configs[n.NodeID] = n
			}
		}
	}

	return configs
}
