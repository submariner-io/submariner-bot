package ghclient

import (
	"context"
	"fmt"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"k8s.io/klog"

	"github.com/submariner-io/submariner-bot/pkg/config"
)

type GH interface {
	AddLabel(issueOrPRNum int, label string) error
	CommentOnPR(prNum int, comment string, args ...interface{})
	ListReviews(prNum int) ([]*github.PullRequestReview, error)
	UpdateDependingPRs(prNum int, baseRef string, branchesToDelete []string) error
}

func New(owner, repo string) (GH, error) {
	ctx := context.Background()
	token, err := config.GetGithubToken()
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	gh := ghClient{
		client: github.NewClient(tc),
		owner:  owner,
		repo:   repo,
	}
	return &gh, nil
}

type ghClient struct {
	client *github.Client
	owner  string
	repo   string
}

func (gh ghClient) AddLabel(issueOrPRNum int, label string) error {
	_, _, err := gh.client.Issues.AddLabelsToIssue(
		context.Background(),
		gh.owner,
		gh.repo,
		issueOrPRNum,
		[]string{label})
	return err
}

func (gh ghClient) CommentOnPR(prNum int, comment string, args ...interface{}) {
	// In GitHub PRs are a sort of issue, so some operations need to be done on the Issues API
	comment = "🤖 " + fmt.Sprintf(comment, args...)
	prComment := github.IssueComment{Body: &comment}
	_, resp, err := gh.client.Issues.CreateComment(
		context.Background(),
		gh.owner,
		gh.repo,
		prNum,
		&prComment)
	// We don't propagate and just log the error
	if err != nil {
		klog.Errorf("Error commenting on pr %d: %s, response: %v", prNum, err, resp)
	}
}

func (gh ghClient) ListReviews(prNum int) ([]*github.PullRequestReview, error) {
	reviews, _, err := gh.client.PullRequests.ListReviews(
		context.Background(),
		gh.owner,
		gh.repo,
		prNum,
		&github.ListOptions{PerPage: 100})

	return reviews, err
}

// fetchPRsWithBase: gets a list of pull requests which have an specific branch as base
func (gh ghClient) fetchPRsWithBase(baseBranch string) ([]*github.PullRequest, error) {
	list, _, err := gh.client.PullRequests.List(context.Background(), gh.owner, gh.repo, &github.PullRequestListOptions{
		Base: baseBranch,
	})
	if err != nil {
		klog.Errorf("An error happened while trying to find PRs dependent on branch: %s on repo %s/%s", baseBranch,
			gh.owner, gh.repo)
	}

	return list, err
}

func (gh ghClient) UpdateDependingPRs(prNum int, baseRef string, branchesToDelete []string) error {
	for _, branchName := range branchesToDelete {
		prs, err := gh.fetchPRsWithBase(branchName)
		if err != nil {
			klog.Errorf("Error fetching dependent PRs for %s: %s", branchName, err)
			gh.CommentOnPR(prNum, "Error fetching dependent PRs for %s: %s", branchName, err)
			return err
		}

		for _, dependentPr := range prs {
			gh.CommentOnPR(prNum, "Updating dependent PRs: %s", *dependentPr.HTMLURL)

			// The PR payloadPR has been merged to pr.PullRequest.Base.Ref, so that should be the new base
			// of the dependent PRs
			dependentPr.Base.Ref = &baseRef
			_, _, err := gh.client.PullRequests.Edit(context.Background(), gh.owner, gh.repo,
				*dependentPr.Number, dependentPr)
			if err != nil {
				klog.Errorf("updating dependent PR: %s : %s", *dependentPr.HTMLURL, err)
				gh.CommentOnPR(prNum, "Error updating dependent PRs: %s : %s", *dependentPr.HTMLURL, err)
				return err
			}

			gh.CommentOnPR(*dependentPr.Number,
				"The base of this PR has been updated to %s\nPlease rebase this branch and remove %s related commits",
				baseRef, *dependentPr.HTMLURL)
		}
	}

	return nil
}
