package pullrequest

import (
	"context"
	"fmt"

	"github.com/go-playground/webhooks/github"
	goGithub "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/config"
)

func commentOnPR(pr *github.PullRequestPayload, ghClient *goGithub.Client, comment string) {

	commentOnPRNum(pr, int(pr.Number), ghClient, comment)
}

func commentOnPRNum(pr *github.PullRequestPayload, prNum int, ghClient *goGithub.Client, comment string) {

	// don't ask me why github uses issue comments for PR comments, PR comments seem to be
	// for code reviews only ¯\_(ツ)_/¯
	comment = "🤖 " + comment
	prComment := goGithub.IssueComment{Body: &comment}
	_, resp, err := ghClient.Issues.CreateComment(
		context.Background(),
		pr.PullRequest.Base.User.Login,
		pr.PullRequest.Base.Repo.Name,
		prNum,
		&prComment)

	// We don't propagate and just log the error
	if err != nil {
		klog.Errorf("Error commeting on pr%d: %s, response: %v", prNum, err, resp)
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

// fetchPRsWithBase: gets a list of pull requests which have an specific branch as base
func fetchPRsWithBase(owner, repo, baseBranch string, ghClient *goGithub.Client) ([]*goGithub.PullRequest, error) {
	list, _, err := ghClient.PullRequests.List(context.Background(), owner, repo, &goGithub.PullRequestListOptions{
		Base: baseBranch,
	})

	if err != nil {
		klog.Errorf("An error happened while trying to find PRs dependent on branch: %s on repo %s/%s", baseBranch,
			owner, repo)
	}

	return list, err
}

func updateDependingPRs(pr *github.PullRequestPayload, ghClient *goGithub.Client, branchesToDelete []string) error {
	for _, branchName := range branchesToDelete {
		prs, err := fetchPRsWithBase(pr.Repository.Owner.Login, pr.Repository.Name, branchName, ghClient)
		if err != nil {
			klog.Errorf("Error fetching dependent PRs for %s: %s", branchName, err)
			commentOnPR(pr, ghClient,
				fmt.Sprintf("Error fetching dependent PRs for %s: %s", branchName, err))
			return err
		}

		for _, dependentPr := range prs {

			commentOnPR(pr, ghClient, fmt.Sprintf("Updating dependent PRs: %s", *dependentPr.HTMLURL))

			// The PR payloadPR has been merged to pr.PullRequest.Base.Ref, so that should be the new base
			// of the dependent PRs
			dependentPr.Base.Ref = &pr.PullRequest.Base.Ref
			ghClient.PullRequests.Edit(context.Background(), pr.Repository.Owner.Login, pr.Repository.Name,
				*dependentPr.Number, dependentPr)

			if err != nil {
				klog.Error("updating dependent PR: %s : %s", *dependentPr.HTMLURL, err)
				commentOnPR(pr, ghClient,
					fmt.Sprintf("Error updating dependent PRs: %s : %s", *dependentPr.HTMLURL, err))
				return err
			}

			commentOnPRNum(pr, *dependentPr.Number, ghClient, fmt.Sprintf(
				"the base of this PR has been updated to %s\n"+
					"please rebase this branch and remove %s related commits", pr.PullRequest.Base.Ref,
				*dependentPr.HTMLURL))
		}

	}
	return nil
}
