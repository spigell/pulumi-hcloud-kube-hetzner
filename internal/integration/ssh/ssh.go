package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// SimpleCheck try to check connection via ssh.
// With simple echo command.
func SimpleCheck(addr, username, privateKey string) error {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("creation of ssh client failed: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("ssh connection failed: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput("echo Hello from $(hostname)")
	if err != nil {
		return fmt.Errorf("ssh command failed. error: %w. Stdout: %s",
			err, output)
	}

	return nil
}
