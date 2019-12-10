package pullrequest

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/go-playground/webhooks/github"
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
)

func Handle(pr github.PullRequestPayload) error {

	baseFullName := logPullRequestInfo(&pr)

	gitRepo, err := git.New(baseFullName, pr.PullRequest.Base.Repo.SSHURL)
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
		return closeBranches(gitRepo, &pr)
	}

	return nil
}

func logPullRequestInfo(pr *github.PullRequestPayload) string {
	prNum := pr.Number
	title := pr.PullRequest.Title
	user := pr.PullRequest.User.Login
	headSSHURL := pr.PullRequest.Head.Repo.SSHURL
	headFullName := pr.PullRequest.Head.Repo.FullName
	headBranch := pr.PullRequest.Head.Ref
	baseSSHURL := pr.PullRequest.Base.Repo.SSHURL
	baseFullName := pr.PullRequest.Base.Repo.FullName
	baseBranch := pr.PullRequest.Base.Ref
	klog.Infof("PR %d %s: %s", prNum, pr.Action, title)
	klog.Infof("  user: %s", user)
	klog.Infof("   head      ssh: %s", headSSHURL)
	klog.Infof("          branch: %s", headBranch)
	klog.Infof("            name: %s", headFullName)
	klog.Infof("   base      ssh: %s", baseSSHURL)
	klog.Infof("          branch: %s", baseBranch)
	klog.Infof("            name: %s", baseFullName)
	return baseFullName
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

	klog.Infof("Branches: %v", branches)
	verFmt := versionedBranchFmt(pr)
	versionBranch := ""
	for v := 1; ; v++ {
		versionBranch = fmt.Sprintf(verFmt, v)
		if branches[versionBranch] == nil {
			break // we found an unused version of the branch, let's use it
		}
	}

	ref := plumbing.ReferenceName("refs/heads/" + versionBranch)
	hash, _ := hex.DecodeString(pr.PullRequest.Head.Sha)

	// TODO: I'm sure there's a better way to do this:
	var refHash plumbing.Hash
	if len(hash) != len(refHash) {
		klog.Errorf("Lengths don't match %d != %d", len(hash), len(refHash))
		return nil
	} else {
		for i, bt := range hash {
			refHash[i] = bt
		}
	}

	hr := plumbing.NewHashReference(ref, refHash)
	err = gitRepo.Repo.Storer.SetReference(hr)

	if err != nil {
		klog.Errorf("Error creating branch reference for %s", versionBranch)
		return err
	}

	klog.Infof("Created branch: %s", ref)

	// https://git-scm.com/book/es/v2/Git-Internals-The-Refspec
	pushOptions := git2.PushOptions{
		RemoteName: "origin",
		Auth:       gitRepo.Auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/heads/%s", versionBranch, versionBranch))},
	}
	gitRepo.Repo.Push(&pushOptions)
	if err != nil {
		klog.Errorf("Error pushing origin with the new branch")
		return err
	}

	klog.Infof("Pushed branch: %s, with options: %v", ref, pushOptions)
	return err
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

	refSpecs := []config.RefSpec{}

	verFmt := versionedBranchFmt(pr)
	versionBranch := ""
	for v := 1; ; v++ {
		versionBranch = fmt.Sprintf(verFmt, v)
		if branches[versionBranch] != nil {
			refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf(":refs/heads/%s", versionBranch)))
			klog.Infof("Deleting branch: %s", versionBranch)
		} else {
			break
		}
	}

	pushOptions := git2.PushOptions{
		RemoteName: "origin",
		Auth:       gitRepo.Auth,
		RefSpecs:   refSpecs,
		Progress:   os.Stderr,
	}

	origin, err := gitRepo.Repo.Remote("origin")
	origin.Push(&pushOptions)

	return err
}

func versionedBranchFmt(pr *github.PullRequestPayload) string {
	return "z_pr/" +
		pr.PullRequest.Head.User.Login + "/" +
		pr.PullRequest.Head.Ref + "/%d"
}
