package pullrequest

import (
	"github.com/go-playground/webhooks/github"
	git2 "gopkg.in/src-d/go-git.v4"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
)

func Handle(pr github.PullRequestPayload) error {

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

	gitRepo, err := git.New(baseFullName, baseSSHURL)
	if err != nil {
		klog.Errorf("git checkout error: %s", err)
		return err
	}
	err = gitRepo.EnsureRemote(user, headSSHURL)
	if err != nil {
		klog.Errorf("git remote setup failed: %s", err)
		return err
	}

	remote, err := gitRepo.Repo.Remote("origin")

	rfs, err := remote.List(&git2.ListOptions{Auth: gitRepo.Auth})

	for _, rf := range rfs {
		klog.Infof(" ref: %s", rf.Name())
	}

	gitRepo.Repo.CreateBranch()
	switch pr.Action {
	case "opened":
	case "synchronize":
	case "closed":
	}

	return nil
}
