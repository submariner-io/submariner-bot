package git

// github.com/src-d/go-git

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/ssh"
	gogit "gopkg.in/src-d/go-git.v4"
	gogitConfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/config"
)

const origin = "origin"

type Git struct {
	repo *gogit.Repository
	name string
	url  string
	auth transport.AuthMethod
}

func New(name, url string) (*Git, error) {

	signer, err := config.GetSSHKey()

	if err != nil {
		return nil, err
	}
	dirName := path.Join("/tmp", "git", name)

	auth := &ssh2.PublicKeys{User: "git", Signer: signer}
	auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	repo, err := gogit.PlainClone(dirName, false, &gogit.CloneOptions{Auth: auth, URL: url})
	if err == nil {
		klog.Infof("Repo %s cloned to %s from %s", name, dirName, url)
	} else if err != nil && err.Error() == "repository already exists" {
		repo, err = gogit.PlainOpen(dirName)
		if err != nil {
			return nil, err
		}
		klog.Infof("Repo %s from disk: %s", name, dirName)
	} else {
		return nil, err
	}

	git := &Git{repo: repo, url: url, name: name, auth: auth}

	err = git.EnsureRemote(origin, url)

	return git, err
}

func (git *Git) EnsureRemote(name, url string) error {
	git.repo.DeleteRemote(name)
	git.repo.CreateRemote(&gogitConfig.RemoteConfig{Name: name, URLs: []string{url}})
	klog.Infof("Remote %s ensured from %s", name, url)
	err := git.FetchRemote(name)

	return err
}

func (git *Git) FetchRemote(name string) error {
	err := git.repo.Fetch(&gogit.FetchOptions{RemoteName: name, Auth: git.auth})
	if err == nil || err.Error() == "already up-to-date" {
		klog.Infof("Remote %s fetched", name)
		return nil
	} else {
		klog.Errorf("Issue fetching remote %s : %s", name, err)
	}
	return err
}

type Branches map[string]*plumbing.Hash

func (git *Git) GetBranches() (Branches, error) {
	branches := make(map[string]*plumbing.Hash)

	remote, err := git.repo.Remote(origin)
	if err != nil {
		return nil, err
	}

	rfs, err := remote.List(&gogit.ListOptions{Auth: git.auth})
	if err != nil {
		return nil, err
	}

	const branchPrefix = "refs/heads/"

	for _, rf := range rfs {
		name := rf.Name().String()
		if strings.HasPrefix(name, branchPrefix) {
			hash := rf.Hash()
			branchName := name[len(branchPrefix):]
			branches[branchName] = &hash
			klog.Infof(" branch: %s hash: %s", branchName, hash.String())
		}
	}
	return branches, nil
}

func (gitRepo *Git) CreateBranch(versionBranch, sha string) error {
	return gitRepo.CreateRef("refs/heads/"+versionBranch, sha)
}

func (gitRepo *Git) CreateRef(refStr, sha string) error {
	ref := plumbing.ReferenceName(refStr)
	refHash, err := getHash(sha)
	if err != nil {
		return err
	}
	hr := plumbing.NewHashReference(ref, refHash)
	err = gitRepo.repo.Storer.SetReference(hr)
	if err != nil {
		return fmt.Errorf("Error creating reference for %s", refStr)
	}
	return err
}

func getHash(sha string) (plumbing.Hash, error) {
	hash, _ := hex.DecodeString(sha)
	var refHash plumbing.Hash
	if len(hash) != len(refHash) {
		return plumbing.Hash{}, fmt.Errorf("Lengths don't match for sha hash %d != %d", len(hash), len(refHash))

	} else {
		for i, bt := range hash {
			refHash[i] = bt
		}
	}
	return refHash, nil
}

func (gitRepo *Git) Push(versionBranch string) error {
	return gitRepo.PushRef(origin, "refs/heads/"+versionBranch)
}

func (gitRepo *Git) PushRef(remote, ref string) error {
	pushOptions := gogit.PushOptions{
		RemoteName: remote,
		Auth:       gitRepo.auth,
		RefSpecs: []gogitConfig.RefSpec{
			gogitConfig.RefSpec(fmt.Sprintf("+%s:%s", ref, ref))},
	}
	return gitRepo.repo.Push(&pushOptions)
}

func (gitRepo *Git) DeleteRemoteBranches(branches []string) error {
	refSpecs := []gogitConfig.RefSpec{}

	for _, branch := range branches {
		refSpecs = append(refSpecs, gogitConfig.RefSpec(fmt.Sprintf(":refs/heads/%s", branch)))
	}

	pushOptions := gogit.PushOptions{
		RemoteName: origin,
		Auth:       gitRepo.auth,
		RefSpecs:   refSpecs,
		Progress:   os.Stderr,
	}

	origin, err := gitRepo.repo.Remote(origin)
	origin.Push(&pushOptions)
	return err
}
