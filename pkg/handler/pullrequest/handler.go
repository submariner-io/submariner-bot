package pullrequest

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/webhooks/github"
	goGithub "github.com/google/go-github/v28/github"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
)

//NOTE: this has been disabled in code for just in case we think it'd valuable to enable later
const enableVersionBranches = false

func Handle(pr github.PullRequestPayload) error {

	logPullRequestInfo(&pr)
	ghClient, err := getGithubClient()
	if err != nil {
		klog.Errorf("creating github client: %s", err)
		return err
	}

	gitRepo, err := git.New(pr.PullRequest.Base.Repo.FullName, pr.PullRequest.Base.Repo.SSHURL)
	if err != nil {
		klog.Errorf("creating git object: %s", err)
		return err
	}

	switch pr.Action {
	case "opened":
		return openOrSync(gitRepo, &pr, ghClient)
	case "synchronize":
		return openOrSync(gitRepo, &pr, ghClient)
	case "closed":
		//TODO: if closed and pr.PullRequest.Merged == true, look for existing PR's pointing to the
		// merged version and change the base to "master" or pr.PullRequest.Base.Ref
		return closeBranches(gitRepo, &pr, ghClient)
	case "reopened":
		//TODO: when re-opened it would be ideal to recover the previous branches, how?
		return openOrSync(gitRepo, &pr, ghClient)

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

func openOrSync(gitRepo *git.Git, pr *github.PullRequestPayload, ghClient *goGithub.Client) error {

	// If the pull request is coming from a local branch
	if pr.PullRequest.Base.Repo.FullName == pr.PullRequest.Head.Repo.FullName {
		if pr.Action == "opened" {
			commentOnPR(pr, ghClient,
				"I see This pr is using the local branch workflow, ignoring it on my side, have fun!")
		}
		return nil
	}

	err := gitRepo.EnsureRemote(pr.PullRequest.User.Login, pr.PullRequest.Head.Repo.SSHURL)
	if err != nil {
		klog.Errorf("git remote setup failed: %s", err)
		return err
	}

	branches, err := gitRepo.GetBranches(git.Origin)
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
		infoMsg = fmt.Sprintf("Created branch: %s", versionBranch)
	} else {
		infoMsg = fmt.Sprintf("Updated branch: %s", versionBranch)
	}

	klog.Infof(infoMsg)

	if err = gitRepo.Push(git.Origin, versionBranch); err != nil {
		klog.Errorf("Error pushing origin with the new branch: %s", err)
		commentOnPR(pr, ghClient, fmt.Sprintf("I had an issue pushing the updated branch: %s", err))
		return err
	}

	commentOnPR(pr, ghClient, infoMsg)

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

func closeBranches(gitRepo *git.Git, pr *github.PullRequestPayload, ghClient *goGithub.Client) error {

	err := gitRepo.EnsureRemote(pr.PullRequest.User.Login, pr.PullRequest.Head.Repo.SSHURL)
	if err != nil {
		klog.Errorf("git remote setup failed: %s", err)
		return err
	}

	branches, err := gitRepo.GetBranches(git.Origin)
	if err != nil {
		klog.Errorf("Error getting branches for origin repo")
		return nil
	}

	branchesToDelete := filterVersionBranches(pr, branches)
	klog.Infof("Deleting branches: %v", branchesToDelete)

	err = gitRepo.DeleteRemoteBranches(git.Origin, branchesToDelete)

	if err != nil {
		klog.Error("Something happened removing branches: %s", err)
	}

	commentOnPR(pr, ghClient, fmt.Sprintf("Closed branches: %s", branchesToDelete))

	return err
}

func filterVersionBranches(pr *github.PullRequestPayload, branches git.Branches) []string {
	branchesToDelete := []string{}
	verBase := versionedBranch(pr) + "/"
	verBranch := versionedBranch(pr)
	for branch, _ := range branches {
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
	return "z_pr/" +
		pr.PullRequest.Head.User.Login + "/" +
		pr.PullRequest.Head.Ref
}
