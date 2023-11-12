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
	Name string
	Decoded *config.Config
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

type DecodedConfig struct {
	Config *config.Config `yaml:"config"`
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
