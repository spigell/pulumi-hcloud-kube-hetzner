package wireguard

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Wireguard struct {
	configPath string
}

func Up(config string) (*Wireguard, error) {
	file, _ := os.OpenFile(filepath.Join(os.TempDir(), "pulumi-wg0.conf"), os.O_WRONLY|os.O_CREATE, 0600)

	_, err := file.WriteString(config)
	if err != nil {
		return nil, err
	}

	if err := exec.Command("sudo", "wg-quick", "up", file.Name()).Run(); err != nil {
		return nil, err
	}

	return &Wireguard{
		configPath: file.Name(),
	}, nil
}

func (w *Wireguard) Close() error {
	if err := exec.Command("sudo", "wg-quick", "down", w.configPath).Run(); err != nil {
		return err
	}

	if err := os.Remove(w.configPath); err != nil {
		return err
	}

	return nil
}
