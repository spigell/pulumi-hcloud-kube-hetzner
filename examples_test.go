package phkh

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

type PulumiConfig struct {
	Config *config.Config `yaml:"config"`
}

func TestExampleWithUnknownFields(t *testing.T) {
	exampleDir := "pulumi-template/examples"

	var decoded *PulumiConfig

	files, err := os.ReadDir(exampleDir)
	assert.NoError(t, err)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(exampleDir, f.Name()))
		assert.NoError(t, err)

		// Remove namespace from file
		re := regexp.MustCompile("pulumi-hcloud-kube-hetzner:")
		res := re.ReplaceAllString(string(content), "")

		decoder := yaml.NewDecoder(strings.NewReader(res))
		decoder.KnownFields(true)

		assert.NoError(t, decoder.Decode(&decoded))
	}
}
