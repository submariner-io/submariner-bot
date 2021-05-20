package handler

import (
	"github.com/go-playground/webhooks/github"

	"github.com/submariner-io/pr-brancher-webhook/pkg/handler/pullrequest"
)

func Handle(payload interface{}) error {
	switch payload.(type) {

	case github.PullRequestPayload:
		return pullrequest.Handle(payload.(github.PullRequestPayload))
	case github.PullRequestReviewPayload:
		return handlePullRequestReview(payload.(github.PullRequestReviewPayload))
	}
	return nil
}
