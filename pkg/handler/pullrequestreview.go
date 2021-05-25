package handler

import (
	"fmt"

	"github.com/go-playground/webhooks/github"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/ghclient"
	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
)

const (
	defaultApprovals = 2
	defaultLabel     = "ready-to-test"
)

type botConfig struct {
	LabelApproved *struct {
		Approvals *int
		Label     *string
	} `yaml:"label-approved"`
}

func handlePullRequestReview(prr github.PullRequestReviewPayload) error {
	prNum := int(prr.PullRequest.Number)
	klog.Infof("handling PR review on %s PR #%d", prr.Repository.FullName, prNum)
	gh, err := ghclient.New(prr.Repository.Owner.Login, prr.Repository.Name)
	if err != nil {
		klog.Errorf("creating github client: %s", err)
		return err
	}

	config, err := readConfig(prr)
	if err != nil {
		klog.Errorf("reading bot config: %s", err)
		return err
	}

	if config.LabelApproved == nil {
		klog.Infof("label when approved not enabled in bot config for PR %s/#d", prr.Repository.Owner.Login, prNum)
		return nil
	}

	reviews, err := gh.ListReviews(prNum)
	if err != nil {
		klog.Errorf("listing reviews: %s", err)
		return err
	}

	approvals := 0
	for _, review := range reviews {
		if *review.State == "APPROVED" {
			approvals++
		}
	}

	minApprovals := *config.LabelApproved.Approvals
	if approvals < minApprovals {
		klog.Infof("%d not enough approvals for PR #%d, need at least %d", approvals, prNum, minApprovals)
		return nil
	}

	label := *config.LabelApproved.Label
	klog.Infof("adding label %s to PR #%d", label, prNum)
	err = gh.AddLabel(prNum, label)
	if err != nil {
		klog.Errorf("error while adding label %s to PR #%d: %s", label, prNum, err)
		return err
	}

	return nil
}

func readConfig(prr github.PullRequestReviewPayload) (*botConfig, error) {
	repoFullName := prr.PullRequest.Base.Repo.FullName
	gitRepo, err := git.New(repoFullName, prr.PullRequest.Base.Repo.SSHURL)
	if err != nil {
		return nil, err
	}

	gitRepo.Lock()
	defer gitRepo.Unlock()

	err = gitRepo.EnsureRemote(prr.PullRequest.User.Login, prr.PullRequest.Head.Repo.SSHURL)
	if err != nil {
		return nil, err
	}

	gitRepo.CheckoutHash(prr.PullRequest.Head.Sha)
	if err != nil {
		return nil, err
	}

	filename := ".submarinerbot.yaml"
	buf, err := gitRepo.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	klog.Infof("read the following config for %s: %s", repoFullName, string(buf))
	config := &botConfig{}
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
