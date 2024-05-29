package phkh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleDir = "cluster-examples"
)

func TestExampleWithUnknownFields(t *testing.T) {
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		suitable, err := hasExamplesDir(path)
		require.NoError(t, err)

		if !suitable {
			return nil
		}

		files, err := os.ReadDir(filepath.Join(path, exampleDir))
		require.NoError(t, err)

		for _, file := range files {
			var c *config.Config
			if !(strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
				return nil
			}

			fullPath := filepath.Join(path, exampleDir, file.Name())

			content, err := os.ReadFile(fullPath)
			require.NoError(t, err)

			decoder := yaml.NewDecoder(strings.NewReader(string(content)))
			decoder.KnownFields(true)

			assert.NoError(t, decoder.Decode(&c), fmt.Sprintf("%s: failed to decode", fullPath))
		}
		return nil
	})
}

func hasExamplesDir(target string) (bool, error) {

	dirs, err := os.ReadDir(target)
	if err != nil {
		return false, err
	}

	for _, file := range dirs {
		if file.IsDir() && file.Name() == exampleDir {
			return true, nil
		}
	}

	return false, nil
}
