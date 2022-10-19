package repoconfig

import (
	"fmt"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

const (
	defaultApprovals = 2
	defaultLabel     = "ready-to-test"
	filename         = ".submarinerbot.yaml"
)

type BotConfig struct {
	LabelApproved *struct {
		Approvals *int
		Label     *string
	} `yaml:"label-approved"`
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

	klog.Infof("read the following config for %s: %s", git.Origin, string(buf))
	config := &BotConfig{}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	if config.LabelApproved != nil {
		if config.LabelApproved.Approvals == nil {
			v := defaultApprovals
			config.LabelApproved.Approvals = &v
		}

		if config.LabelApproved.Label == nil {
			v := defaultLabel
			config.LabelApproved.Label = &v
		}
	}

	return config, nil
}
