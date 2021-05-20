package handler

import (
	"github.com/go-playground/webhooks/github"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/ghclient"
)

func handlePullRequestReview(prr github.PullRequestReviewPayload) error {
	prNum := int(prr.PullRequest.Number)
	gh, err := ghclient.New(prr.Repository.Owner.Login, prr.Repository.Name)
	if err != nil {
		klog.Errorf("creating github client: %s", err)
		return err
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

	if approvals > 1 {
		err := gh.AddLabel(prNum, "ready-to-test")
		if err != nil {
			klog.Errorf("adding label to pull request: %s", err)
			return err
		}
	}

	return nil
}
