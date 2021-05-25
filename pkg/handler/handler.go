package handler

import (
	"sync"

	"github.com/go-playground/webhooks/github"

	"github.com/submariner-io/pr-brancher-webhook/pkg/handler/pullrequest"
)

func EventsToHandle() []github.Event {
	return []github.Event{
		github.ReleaseEvent,
		github.PullRequestEvent,
		github.PullRequestReviewEvent,
	}
}

func Handle(payload interface{}) error {
	switch payload.(type) {

	case github.PullRequestPayload:
		pr := payload.(github.PullRequestPayload)
		mutex := getMutex(pr.PullRequest.Base.Repo.FullName)
		mutex.Lock()
		defer mutex.Unlock()
		return pullrequest.Handle(pr)

	case github.PullRequestReviewPayload:
		prReview := payload.(github.PullRequestReviewPayload)
		mutex := getMutex(prReview.PullRequest.Base.Repo.FullName)
		mutex.Lock()
		defer mutex.Unlock()
		return handlePullRequestReview(prReview)
	}
	return nil
}

var repoMutexMap map[string]sync.Mutex
var repoMutex sync.Mutex

func getMutex(repoName string) sync.Mutex {
	repoMutex.Lock()
	defer repoMutex.Unlock()

	if mutex, ok := repoMutexMap[repoName]; ok {
		return mutex
	}

	repoMutexMap[repoName] = sync.Mutex{}
	return repoMutexMap[repoName]
}
