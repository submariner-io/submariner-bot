package git

// github.com/src-d/go-git

import (
	"path"

	gogit "gopkg.in/src-d/go-git.v4"
	gogitConfig "gopkg.in/src-d/go-git.v4/config"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/config"
)

type Git struct {
	Repo *gogit.Repository
	Name string
	URL  string
}

func New(name, url string) (*Git, error) {

	signer, err := config.GetSSHKey()
	if err != nil {
		return nil, err
	}
	dirName := path.Join("/tmp", name)

	auth := &ssh2.PublicKeys{User: "git", Signer: signer}
	repo, err := gogit.PlainClone(dirName, false,
		&gogit.CloneOptions{Auth: auth, URL: url})
	if err == nil {
		klog.Info("Repo %s cloned to %s from %s", name, dirName, url)
	} else if err != nil && err.Error() == "repository already exists" {
		repo, err = gogit.PlainOpen(dirName)
		if err != nil {
			return nil, err
		}
		klog.Info("Repo %s from disk: %s", name, dirName)
	} else {
		return nil, err
	}

	git := &Git{Repo: repo, URL: url, Name: name}

	err = git.EnsureRemote("origin", url)

	return git, err
}

func (git *Git) EnsureRemote(name, url string) error {
	git.Repo.DeleteRemote(name)
	git.Repo.CreateRemote(&gogitConfig.RemoteConfig{Name: name, URLs: []string{url}})
	return git.Repo.Fetch(&gogit.FetchOptions{RemoteName: name})
}
