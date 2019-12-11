package pullrequest

import (
	"fmt"

	"github.com/go-playground/webhooks/github"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
)

func Handle(pr github.PullRequestPayload) error {

	logPullRequestInfo(&pr)

	gitRepo, err := git.New(pr.PullRequest.Base.Ref, pr.PullRequest.Base.Repo.SSHURL)
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

	branches, err := gitRepo.GetBranches("origin")
	if err != nil {
		klog.Errorf("Error getting branches for origin repo")
		return nil
	}

	versionBranch := getNextVersionBranch(pr, branches)

	err = gitRepo.CreateBranch(versionBranch, pr.PullRequest.Head.Sha)
	if err != nil {
		return err
	}

	klog.Infof("Created branch: %s", versionBranch)

	if err = gitRepo.Push("origin", versionBranch); err != nil {
		klog.Errorf("Error pushing origin with the new branch")
		return err
	}

	klog.Infof("Pushed branch: %s", versionBranch)
	return err
}

func getNextVersionBranch(pr *github.PullRequestPayload, branches git.Branches) string {
	verFmt := versionedBranchFmt(pr)
	versionBranch := ""
	for v := 1; ; v++ {
		versionBranch = fmt.Sprintf(verFmt, v)
		if branches[versionBranch] == nil {
			break // we found an unused version of the branch, let's use it
		}
	}
	return versionBranch
}

func closeBranches(gitRepo *git.Git, pr *github.PullRequestPayload) error {

	err := gitRepo.EnsureRemote(pr.PullRequest.User.Login, pr.PullRequest.Head.Repo.SSHURL)
	if err != nil {
		klog.Errorf("git remote setup failed: %s", err)
		return err
	}

	branches, err := gitRepo.GetBranches("origin")
	if err != nil {
		klog.Errorf("Error getting branches for origin repo")
		return nil
	}

	branchesToDelete := filterBranchesToDelete(pr, branches)

	gitRepo.DeleteRemoteBranches("origin", branchesToDelete)

	return err
}

func filterBranchesToDelete(pr *github.PullRequestPayload, branches git.Branches) []string {
	branchesToDelete := []string{}
	verFmt := versionedBranchFmt(pr)
	versionBranch := ""
	for v := 1; ; v++ {
		versionBranch = fmt.Sprintf(verFmt, v)
		if branches[versionBranch] == nil {
			break
		}
		branchesToDelete = append(branchesToDelete, versionBranch)
		klog.Infof("Deleting branch: %s", versionBranch)
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
