package pullrequest

import (
	"context"

	"github.com/go-playground/webhooks/github"
	goGithub "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/config"
)

func commentOnPR(pr *github.PullRequestPayload, ghClient *goGithub.Client, comment string) {

	// don't ask me why github uses issue comments for PR comments, PR comments seem to be
	// for code reviews only ¯\_(ツ)_/¯
	comment = "🤖 " + comment
	prComment := goGithub.IssueComment{Body: &comment}
	_, resp, err := ghClient.Issues.CreateComment(
		context.Background(),
		pr.PullRequest.Base.User.Login,
		pr.PullRequest.Base.Repo.Name,
		int(pr.PullRequest.Number),
		&prComment)

	if err != nil {
		klog.Errorf("Error commeting on pr%d: %s, response: %v", pr.Number, err, resp)
	}
}

func getGithubClient() (*goGithub.Client, error) {
	ctx := context.Background()
	token, err := config.GetGithubToken()
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return goGithub.NewClient(tc), nil
}
