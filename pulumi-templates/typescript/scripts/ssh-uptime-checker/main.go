package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	pollingInterval = 2 * time.Second
)

func main() {
	addr := os.Args[1]
	username := os.Args[2]
	privateKey := os.Getenv("CHECKER_SSH_PRIVATE_KEY")

	if len(strings.Split(addr, ":")) == 1 {
		addr = fmt.Sprintf("%s:22", addr)
	}

	//	privateKey, err := os.ReadFile(privateKeyPath)
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(pollingInterval)

		client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         1 * time.Second,
		})

		if err != nil {
			// SSH connection failed, indicating the server might still be rebooting
			fmt.Printf("SSH connection failed: %s, waiting... \n", err.Error())
			continue
		}
		// SSH connection successful, check uptime
		session, err := client.NewSession()
		if err != nil {
			fmt.Printf("SSH connection failed: %s, waiting... \n", err.Error())
			continue
		}
		defer session.Close()

		// Run the uptime command
		output, err := session.CombinedOutput("cat /proc/uptime")
		if err != nil {
			fmt.Printf("SSH command failed: %s, waiting... \n", err.Error())
			continue
		}

		// Check if the output contains information about server uptime
		if rebooted(output) {
			fmt.Println("Server rebooted and accessible via SSH")
			break

		}

		fmt.Println("Server is not yet rebooted, waiting...")
	}
}

func rebooted(output []byte) bool {
	// The uptime command returns two values, the first is the uptime in seconds
	// 84.03 153.29
	// Let's just check if the first value is less than 180 (3 min)
	sec := strings.Split(strings.Split(string(output), " ")[0], ".")[0]
	i, err := strconv.Atoi(sec)

	if err != nil {
		log.Fatal(err)
	}

	if i < 600 {
		return true
	}

	return false
}
