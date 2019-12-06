package git

// github.com/src-d/go-git

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
	gogit "gopkg.in/src-d/go-git.v4"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func Run() {
	pem, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		panic(err)
	}
	auth := &ssh2.PublicKeys{User: "git", Signer: signer}
	repo, err := gogit.PlainClone("/tmp/git/", false,
		&gogit.CloneOptions{Auth: auth, URL: "git@github.com:submariner-io/actions-and-branch-playground.git"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", repo)
}
