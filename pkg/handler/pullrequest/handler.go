package pullrequest

import (
	"github.com/go-playground/webhooks/github"
	"k8s.io/klog"
)

func Handle(payload github.PullRequestPayload) error {

	klog.Infof("Received pull request payload: %v", payload)
	return nil
}
