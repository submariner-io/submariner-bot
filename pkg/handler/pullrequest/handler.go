package pullrequest

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/webhooks/github"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
)

func Handle(pr github.PullRequestPayload) error {

	logPullRequestInfo(&pr)

	gitRepo, err := git.New(pr.PullRequest.Base.Repo.FullName, pr.PullRequest.Base.Repo.SSHURL)
	if err != nil {
		klog.Errorf("creating git object: %s", err)
		return err
	}

	switch pr.Action {
	case "opened":
		return openOrSync(gitRepo, &pr)
	case "synchronize":
		return openOrSync(gitRepo, &pr)
	case "closed":
		//TODO: if closed and pr.PullRequest.Merged == true, look for existing PR's pointing to the
		// merged version and change the base to "master" or pr.PullRequest.Base.Ref
		return closeBranches(gitRepo, &pr)
	case "reopened":
		//TODO: when re-opened it would be ideal to recover the previous branches, how?
		return openOrSync(gitRepo, &pr)

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

func openOrSync(gitRepo *git.Git, pr *github.PullRequestPayload) error {
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

	versionBranch := getNextVersionBranch(pr, branches)

	err = gitRepo.CreateBranch(versionBranch, pr.PullRequest.Head.Sha)
	if err != nil {
		return err
	}

	// This one we don't delete it later, but it's stored internally in git, and
	// we can restore it if the PR is re-opened
	prRef := "refs/pr/" + versionBranch
	err = gitRepo.CreateRef(prRef, pr.PullRequest.Head.Sha)

	klog.Infof("Created branch: %s", versionBranch)

	if err = gitRepo.Push(git.Origin, versionBranch); err != nil {
		klog.Errorf("Error pushing origin with the new branch")
		return err
	}

	if err = gitRepo.PushRef(git.Origin, prRef); err != nil {
		klog.Errorf("Error pushing origin with the new pr ref")
		return err
	}

	klog.Infof("Pushed branch: %s , and ref: %s", versionBranch, prRef)
	return err
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

func closeBranches(gitRepo *git.Git, pr *github.PullRequestPayload) error {

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

	gitRepo.DeleteRemoteBranches(git.Origin, branchesToDelete)

	return err
}

func filterVersionBranches(pr *github.PullRequestPayload, branches git.Branches) []string {
	branchesToDelete := []string{}
	verBase := versionedBranchBase(pr)
	for branch, _ := range branches {
		if strings.HasPrefix(branch, verBase) {
			branchesToDelete = append(branchesToDelete, branch)
		}
	}
	return branchesToDelete
}

func versionedBranchFmt(pr *github.PullRequestPayload) string {
	return versionedBranchBase(pr) + "%d"
}

func versionedBranchBase(pr *github.PullRequestPayload) string {
	return "z_pr/" +
		pr.PullRequest.Head.User.Login + "/" +
		pr.PullRequest.Head.Ref + "/"
}
