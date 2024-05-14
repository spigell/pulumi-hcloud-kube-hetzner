package phkh

import (
	"fmt"
	"os"
	"path/filepath"
//	"regexp"
	"strings"
	"testing"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

type PulumiConfig struct {
	Config *Config `yaml:"config"`
}

type Config struct {
	Clusters map[string]*config.Config `yaml:"pulumi-hcloud-kube-hetzner:clusters"`
}

func TestExampleWithUnknownFields(t *testing.T) {
	exampleDir := "examples"

	var decoded PulumiConfig

	files, err := os.ReadDir(exampleDir)
	assert.NoError(t, err)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), "k3s-private-non-ha-simple.yaml") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(exampleDir, f.Name()))
		assert.NoError(t, err)

		// Remove namespace from file
		// re := regexp.MustCompile("pulumi-hcloud-kube-hetzner:")
		// res := re.ReplaceAllString(string(content), "")

		decoder := yaml.NewDecoder(strings.NewReader(string(content)))
		decoder.KnownFields(true)
		fmt.Println(string(content))

		assert.NoError(t, decoder.Decode(&decoded), fmt.Sprintf("%s failed to decode", f.Name()))

//		cluster, ok := decoded.Config.Clusters["main"]
//		if !ok {
//			t.Skip()
//		}
	}
}
