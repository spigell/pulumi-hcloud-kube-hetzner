package wireguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Wireguard struct {
	configPath string
}

func Up(config string) (*Wireguard, error) {
	file, _ := os.OpenFile(filepath.Join(os.TempDir(), "pulumi-wg0.conf"), os.O_WRONLY|os.O_CREATE, 0o600)

	_, err := file.WriteString(config)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("sudo", "wg-quick", "up", file.Name())
	output, err := cmd.CombinedOutput()

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run wg-quick: %w: %s", err, output)
	}

	return &Wireguard{
		configPath: file.Name(),
	}, nil
}

func (w *Wireguard) Close() error {
	if err := exec.Command("sudo", "wg-quick", "down", w.configPath).Run(); err != nil {
		return err
	}

	return os.Remove(w.configPath)
}
