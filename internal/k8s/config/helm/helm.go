package helm

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	defaultVerFilePath = "versions/default-helm-versions.yaml"
)

type Config struct {
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
