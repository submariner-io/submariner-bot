package pullrequest

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/webhooks/github"
	"github.com/submariner-io/pr-brancher-webhook/pkg/config/repoconfig"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/ghclient"
	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
)

//NOTE: this has been disabled in code for just in case we think it'd valuable to enable later
const enableVersionBranches = false

func Handle(pr github.PullRequestPayload) error {

	logPullRequestInfo(&pr)
	gh, err := ghclient.New(pr.Repository.Owner.Login, pr.Repository.Name)
	if err != nil {
		klog.Errorf("creating github client: %s", err)
		return err
	}

	gitRepo, err := git.New(pr.PullRequest.Base.Repo.FullName, pr.PullRequest.Base.Repo.SSHURL)
	if err != nil {
		klog.Errorf("creating git object: %s", err)
		return err
	}

	gitRepo.Lock()
	defer gitRepo.Unlock()

	switch pr.Action {
	case "opened":
		return openOrSync(gitRepo, &pr, gh)
	case "synchronize":
		return openOrSync(gitRepo, &pr, gh)
	case "closed":
		//TODO: if closed and pr.PullRequest.Merged == true, look for existing PR's pointing to the
		// merged version and change the base to "master" or pr.PullRequest.Base.Ref
		return closeBranches(gitRepo, &pr, gh)
	case "reopened":
		//TODO: when re-opened it would be ideal to recover the previous branches, how?
		return openOrSync(gitRepo, &pr, gh)
	}

	return nil
}

func logPullRequestInfo(pr *github.PullRequestPayload) {

	klog.Infof("PR %d %s: %s", pr.Number, pr.Action, pr.PullRequest.Title)
	klog.Infof("  user: %s", pr.PullRequest.User.Login)
	klog.Infof("   head      ssh: %s", pr.PullRequest.Head.Repo.SSHURL)
	klog.Infof("          branch: %s", pr.PullRequest.Head.Ref)
	klog.Infof("            name: %s", pr.PullRequest.Head.Repo.FullName)
	klog.Infof("   base      ssh: %s", pr.PullRequest.Base.Repo.SSHURL)
	klog.Infof("          branch: %s", pr.PullRequest.Base.Repo.FullName)
	klog.Infof("            name: %s", pr.PullRequest.Base.Ref)
}

func openOrSync(gitRepo *git.Git, pr *github.PullRequestPayload, gh ghclient.GH) error {
	prNum := int(pr.Number)

	config, err := repoconfig.Read(gitRepo, pr.PullRequest.Base.Sha)
	if err != nil {
		klog.Infof("Error reading bot config: %s", err)
	}

	readyToReviewMsg := ""
	if config != nil && config.LabelApproved != nil {
		readyToReviewMsg += fmt.Sprintf("\n🚀 Full E2E won't run until the %q label is applied. "+
			"I will add it automatically once the PR has %d approvals, or you can add it manually.",
			*(config.LabelApproved.Label), *(config.LabelApproved.Approvals))
	}

	// If the pull request is coming from a local branch
	if pr.PullRequest.Base.Repo.FullName == pr.PullRequest.Head.Repo.FullName {
		// We only comment if the PR isn't from a bot, to avoid affecting their behaviour
		// (e.g. dependabot stops maintaining PRs automatically if they're commented)
		if pr.Action == "opened" && pr.PullRequest.User.Type != "Bot" {
			gh.CommentOnPR(prNum, "I see this PR is using the local branch workflow, ignoring it on my side, have fun!"+readyToReviewMsg)
		}
		return nil
	}

	err = gitRepo.EnsureRemote(pr.PullRequest.User.Login, pr.PullRequest.Head.Repo.SSHURL)
	if err != nil {
		klog.Errorf("git remote setup failed: %s", err)
		return err
	}

	branches, err := gitRepo.GetBranches()
	if err != nil {
		klog.Errorf("Error getting branches for origin repo")
		return nil
	}

	versionBranch := getVersionBranch(pr, branches)

	err = gitRepo.CreateBranch(versionBranch, pr.PullRequest.Head.Sha)
	if err != nil {
		return err
	}

	var infoMsg string
	if branches[versionBranch] == nil {
		infoMsg = fmt.Sprintf("Created branch: %s %s", versionBranch, readyToReviewMsg)
	}

	klog.Infof(infoMsg)

	if err = gitRepo.Push(versionBranch); err != nil {
		klog.Errorf("Error pushing origin with the new branch: %s", err)
		gh.CommentOnPR(prNum, "I had an issue pushing the updated branch: %s", err)
		return err
	}

	if infoMsg != "" {
		gh.CommentOnPR(prNum, infoMsg)
	}

	klog.Infof("Pushed branch: %s", versionBranch)
	return err
}

func getVersionBranch(pr *github.PullRequestPayload, branches git.Branches) string {
	if enableVersionBranches {
		return getNextVersionBranch(pr, branches)
	} else {
		return versionedBranch(pr)
	}
}

func getNextVersionBranch(pr *github.PullRequestPayload, branches git.Branches) string {
	existing := filterVersionBranches(pr, branches)

	num := 0
	for _, branch := range existing {
		parts := strings.Split(branch, "/")
		if numBr, _ := strconv.Atoi(parts[len(parts)-1]); numBr > num {
			num = numBr
		}
	}
	return fmt.Sprintf(versionedBranchFmt(pr), num+1)
}

func closeBranches(gitRepo *git.Git, prPayload *github.PullRequestPayload, gh ghclient.GH) error {
	prNum := int(prPayload.Number)
	err := gitRepo.EnsureRemote(prPayload.PullRequest.User.Login, prPayload.PullRequest.Head.Repo.SSHURL)
	if err != nil {
		klog.Errorf("git remote setup failed: %s", err)
		return err
	}

	branches, err := gitRepo.GetBranches()
	if err != nil {
		klog.Errorf("Error getting branches for origin repo")
		return nil
	}

	branchesToDelete := filterVersionBranches(prPayload, branches)
	klog.Infof("Deleting branches: %v", branchesToDelete)

	if err = gh.UpdateDependingPRs(prNum, prPayload.PullRequest.Base.Ref, branchesToDelete); err != nil {
		return err
	}

	if err = gitRepo.DeleteRemoteBranches(branchesToDelete); err != nil {
		klog.Errorf("Something happened removing branches: %s", err)
	} else {
		gh.CommentOnPR(prNum, "Closed branches: %s", branchesToDelete)
	}
	return err
}

func filterVersionBranches(pr *github.PullRequestPayload, branches git.Branches) []string {
	branchesToDelete := []string{}
	verBase := versionedBranch(pr) + "/"
	verBranch := versionedBranch(pr)
	for branch := range branches {
		if strings.HasPrefix(branch, verBase) || branch == verBranch {
			branchesToDelete = append(branchesToDelete, branch)
		}
	}
	return branchesToDelete
}

func versionedBranchFmt(pr *github.PullRequestPayload) string {
	return versionedBranch(pr) + "/%d"
}

func versionedBranch(pr *github.PullRequestPayload) string {
	return fmt.Sprintf("z_pr%d/%s/%s", pr.Number, pr.PullRequest.Head.User.Login, pr.PullRequest.Head.Ref)
}
