package handler

import (
	"github.com/go-playground/webhooks/github"
	"github.com/submariner-io/pr-brancher-webhook/pkg/config/repoconfig"
	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/ghclient"
)

func handlePullRequestReview(prr github.PullRequestReviewPayload) error {
	prNum := int(prr.PullRequest.Number)
	klog.Infof("handling PR review on %s PR #%d", prr.Repository.FullName, prNum)
	gh, err := ghclient.New(prr.Repository.Owner.Login, prr.Repository.Name)
	if err != nil {
		klog.Errorf("creating github client: %s", err)
		return err
	}

	gitRepo, err := git.New(prr.PullRequest.Base.Repo.FullName, prr.PullRequest.Base.Repo.SSHURL)
	if err != nil {
		klog.Errorf("creating git object: %s", err)
		return err
	}

	gitRepo.Lock()
	defer gitRepo.Unlock()

	config, err := repoconfig.Read(gitRepo, prr.PullRequest.Base.Sha)
	if err != nil {
		klog.Infof("Error reading bot config: %s", err)
		return err
	}

	if config.LabelApproved() == nil {
		klog.Infof("label when approved not enabled in bot config for PR %s/#%d", prr.Repository.Owner.Login, prNum)
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

	minApprovals := config.LabelApproved().Approvals()
	if approvals < minApprovals {
		klog.Infof("%d not enough approvals for PR #%d, need at least %d", approvals, prNum, minApprovals)
		return nil
	}

	label := config.LabelApproved().Label()
	klog.Infof("adding label %s to PR #%d", label, prNum)
	err = gh.AddLabel(prNum, label)
	if err != nil {
		klog.Errorf("error while adding label %s to PR #%d: %s", label, prNum, err)
		return err
	}

	return nil
}
