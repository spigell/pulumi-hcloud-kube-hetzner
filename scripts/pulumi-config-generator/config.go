package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Content string
}

type PulumiProject struct {
	// We need only project name
	Name string
}

func NewConfig(path string) *Config {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("error while reading file: %v", err)
	}

	return &Config{
		Content: string(file),
	}
}

func (c *Config) WithReplaceImageValue(image string) *Config {
	target := "image: '<put-your-image-here>'"
	new := fmt.Sprintf("image: '%s'", image)

	c.Content = strings.ReplaceAll(c.Content, target, new)

	return c
}

func (c *Config) WithReplaceProjectName() * Config {
	target := "pulumi-hcloud-kube-hetzner"
	new := getProjectName()
	c.Content = strings.ReplaceAll(c.Content, target, new)

	return c
}

func getProjectName() string {
	var proj *PulumiProject
	file, err := os.ReadFile("Pulumi.yaml")
	if err != nil {
		log.Fatalf("error while reading project file: %v", err)
	}

	err = yaml.Unmarshal(file, &proj)
	if err != nil {
		log.Fatalf("error while unmarshaling file: %v", err)
	}

	return proj.Name
}
