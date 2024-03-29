package handler

import (
	"github.com/go-playground/webhooks/v6/github"

	"github.com/submariner-io/submariner-bot/pkg/handler/pullrequest"
)

func EventsToHandle() []github.Event {
	return []github.Event{
		github.ReleaseEvent,
		github.PullRequestEvent,
		github.PullRequestReviewEvent,
	}
}

func Handle(payload interface{}) error {
	switch payload := payload.(type) {

	case github.PullRequestPayload:
		return pullrequest.Handle(payload)
	case github.PullRequestReviewPayload:
		return handlePullRequestReview(payload)
	}
	return nil
}
