package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"gopkg.in/yaml.v3"
)

type Example struct {
	Name    string
	Decoded *config.Config
}

type DecodedConfig struct {
	Config *config.Config `yaml:"config"`
}

func DiscoverExample(path string) (*Example, error) {
	example := &Example{
		Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}

	decoded, err := decodeConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	example.Decoded = decoded

	return example, nil
}

func (e *Example) NodesIDs() []string {
	var ids []string

	for _, pool := range e.Decoded.Nodepools.Agents {
		for _, n := range pool.Nodes {
			ids = append(ids, n.ID)
		}
	}

	for _, pool := range e.Decoded.Nodepools.Servers {
		for _, n := range pool.Nodes {
			ids = append(ids, n.ID)
		}
	}

	return ids
}

func (e *Example) UniqConfigsByNodes() map[string]*config.Node {
	configs := make(map[string]*config.Node)

	for _, pool := range e.Decoded.Nodepools.Agents {
		for _, n := range pool.Nodes {
			if n.Server != nil || n.K3s != nil {
				configs[n.ID] = n
			}
		}
	}

	for _, pool := range e.Decoded.Nodepools.Servers {
		for _, n := range pool.Nodes {
			if n.Server != nil || n.K3s != nil {
				configs[n.ID] = n
			}
		}
	}

	return configs
}

func decodeConfig(path string) (*config.Config, error) {
	var decoded *DecodedConfig
	content, err := os.ReadFile(path)

	// Remove namespace from file
	re := regexp.MustCompile("pulumi-hcloud-kube-hetzner:")
	res := re.ReplaceAllString(string(content), "")

	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err = yaml.Unmarshal([]byte(res), &decoded); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return decoded.Config, nil
}
