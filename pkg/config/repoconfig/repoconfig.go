package repoconfig

import (
	"fmt"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

type BotConfig struct {
	yaml BotConfigYAML
}

type BotConfigLabelApproved struct {
	yaml LabelApprovedYAML
}

func (c *BotConfig) LabelApproved() *BotConfigLabelApproved {
	if c.yaml.LabelApproved == nil {
		return nil
	}

	return &BotConfigLabelApproved{yaml: *c.yaml.LabelApproved}
}

func (c *BotConfigLabelApproved) Approvals() int {
	if c.yaml.Approvals == nil {
		return defaultApprovals
	}

	return *c.yaml.Approvals
}

func (c *BotConfigLabelApproved) Label() string {
	if c.yaml.Label == nil {
		return defaultLabel
	}

	return *c.yaml.Label
}

func Read(gitRepo *git.Git, sha string) (*BotConfig, error) {
	err := gitRepo.CheckoutHash(sha)
	if err != nil {
		return nil, err
	}

	buf, err := gitRepo.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config, err := parseConfigYAML(buf)
	if err != nil {
		klog.Errorf("Error reading the following config for %s:\n%s", sha, string(buf))
		return nil, fmt.Errorf("in file %q from %s: %v", filename, sha, err)
	}

	return config, nil
}

func parseConfigYAML(buf []byte) (*BotConfig, error) {
	configYAML := BotConfigYAML{}
	err := yaml.Unmarshal(buf, &configYAML)
	return &BotConfig{yaml: configYAML}, err
}
