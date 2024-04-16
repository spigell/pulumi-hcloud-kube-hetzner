package helm

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

const (
	defaultVerFilePath = "versions/default-helm-versions.yaml"
)

type Config struct {
	valuesFiles pulumi.AssetOrArchiveArray `json:"-"`

	// ValuesFilePaths is a list of path/to/file to values files.
	// See https://www.pulumi.com/registry/packages/kubernetes/api-docs/helm/v3/release/#valueyamlfiles_nodejs for details.
	ValuesFilePath []string `json:"values-files" yaml:"values-files"`
	// Version is version of helm chart.
	// Default is taken from default-helm-versions.yaml in template's versions directory.
	Version string
}

func GetDefaultVersion(addon string) (string, error) {
	versions, err := parseDefaultVersionsFile()
	if err != nil {
		return "", fmt.Errorf("unable to parse default versions file: %w", err)
	}

	v, ok := versions[addon]
	if !ok {
		return "", fmt.Errorf("no default version found for %s", addon)
	}

	return v.(string), nil
}

func parseDefaultVersionsFile() (map[string]interface{}, error) {
	m := make(map[string]interface{})

	file, err := os.ReadFile(defaultVerFilePath)
	if err != nil {
		return m, fmt.Errorf("unable to read default versions file: %w", err)
	}
	if err := yaml.Unmarshal(file, m); err != nil {
		return m, fmt.Errorf("unable to parse default versions file: %w", err)
	}

	return m, nil
}

func (c *Config) ValuesFiles() pulumi.AssetOrArchiveArray {
	return c.valuesFiles
}

func (c *Config) SetValuesFiles(assets pulumi.AssetOrArchiveArray) {
	c.valuesFiles = assets
}
