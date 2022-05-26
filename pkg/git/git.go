package git

// github.com/src-d/go-git

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"golang.org/x/crypto/ssh"
	gogit "gopkg.in/src-d/go-git.v4"
	gogitConfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/config"
)

const Origin = "origin"

var projectsLock sync.Mutex
var projects map[string]*Git = make(map[string]*Git)

type Git struct {
	repo *gogit.Repository
	name string
	url  string
	auth transport.AuthMethod
	lock sync.Mutex
}

func New(name, url string) (*Git, error) {
	projectsLock.Lock()
	defer projectsLock.Unlock()

	if val, ok := projects[name]; ok {
		val.Lock()
		defer val.Unlock()
		err := val.EnsureAndFetchOrigin()
		return val, err
	}

	signer, err := config.GetSSHKey()

	if err != nil {
		return nil, err
	}
	dirName := dirName(name)

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

	err = git.EnsureAndFetchOrigin()
	projects[name] = git

	return git, err
}

func (git *Git) EnsureAndFetchOrigin() error {
	return git.EnsureAndFetch(Origin, git.url)
}

func dirName(name string) string {
	return path.Join("/tmp", "git", name)
}

func (git *Git) Lock() {
	git.lock.Lock()
}

func (git *Git) Unlock() {
	git.lock.Unlock()
}

func (git *Git) EnsureAndFetch(name, url string) error {
	if err := git.repo.DeleteRemote(name); err != nil {
		if err != gogit.ErrRemoteNotFound {
			return err
		}
	}
	_, err := git.repo.CreateRemote(&gogitConfig.RemoteConfig{Name: name, URLs: []string{url}})
	if err != nil {
		return err
	}
	klog.Infof("Remote %s ensured from %s", name, url)

	err = git.repo.Fetch(&gogit.FetchOptions{RemoteName: name, Auth: git.auth})
	if err == nil || err.Error() == "already up-to-date" {
		klog.Infof("Remote %s fetched", name)
		return nil
	}

	klog.Errorf("Issue fetching remote %s : %s", name, err)
	return err
}

type Branches map[string]*plumbing.Hash

func (git *Git) GetBranches() (Branches, error) {
	branches := make(map[string]*plumbing.Hash)

	remote, err := git.repo.Remote(Origin)
	if err != nil {
		return nil, err
	}

	rfs, err := remote.List(&gogit.ListOptions{Auth: git.auth})
	if err != nil {
		return nil, err
	}

	for _, rf := range rfs {
		name := rf.Name()
		if name.IsBranch() {
			hash := rf.Hash()
			branchName := name.Short()
			branches[branchName] = &hash
			klog.Infof(" branch: %s hash: %s", branchName, hash.String())
		}
	}
	return branches, nil
}

func (gitRepo *Git) CreateBranch(branch, sha string) error {
	ref := plumbing.NewBranchReferenceName(branch)
	refHash, err := getHash(sha)
	if err != nil {
		return err
	}
	hr := plumbing.NewHashReference(ref, refHash)
	err = gitRepo.repo.Storer.SetReference(hr)
	if err != nil {
		return fmt.Errorf("Error creating reference for %s", ref)
	}
	return err
}

func getHash(sha string) (plumbing.Hash, error) {
	hash, _ := hex.DecodeString(sha)
	var refHash plumbing.Hash
	if len(hash) != len(refHash) {
		return plumbing.Hash{}, fmt.Errorf("Lengths don't match for sha hash %d != %d", len(hash), len(refHash))

	} else {
		copy(refHash[:], hash)
	}
	return refHash, nil
}

func (gitRepo *Git) Push(branch string) error {
	ref := plumbing.NewBranchReferenceName(branch)
	pushOptions := gogit.PushOptions{
		RemoteName: Origin,
		Auth:       gitRepo.auth,
		RefSpecs: []gogitConfig.RefSpec{
			gogitConfig.RefSpec(fmt.Sprintf("+%s:%s", ref, ref))},
	}
	return gitRepo.repo.Push(&pushOptions)
}

func (gitRepo *Git) DeleteRemoteBranches(branches []string) error {
	refSpecs := []gogitConfig.RefSpec{}

	for _, branch := range branches {
		refSpecs = append(refSpecs, gogitConfig.RefSpec(fmt.Sprintf(":%s", plumbing.NewBranchReferenceName(branch))))
	}

	pushOptions := gogit.PushOptions{
		RemoteName: Origin,
		Auth:       gitRepo.auth,
		RefSpecs:   refSpecs,
		Progress:   os.Stderr,
	}

	origin, err := gitRepo.repo.Remote(Origin)
	if err != nil {
		return err
	}
	return origin.Push(&pushOptions)
}

func (g *Git) CheckoutHash(hash string) error {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Hash:  plumbing.NewHash(hash),
		Force: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *Git) ReadFile(file string) ([]byte, error) {
	filename := path.Join(dirName(g.name), file)
	return ioutil.ReadFile(filename)
}
