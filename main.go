package main

import (
	"net/http"
	"os"

	"github.com/go-playground/webhooks/github"

	"github.com/submariner-io/pr-brancher-webhook/pkg/git"
	"github.com/submariner-io/pr-brancher-webhook/pkg/handler"
)

const (
	path = "/webhooks"
)

func main() {
	git.Run()
	os.Exit(0)
	hook, _ := github.New(github.Options.Secret("MyGitHubSuperSecretSecrect...?"))

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.ReleaseEvent, github.PullRequestEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
			} else {
				w.WriteHeader(500)
			}
			return
		}
		err = handler.Handle(payload)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("An error happened: " + err.Error()))
		}
	})
	http.ListenAndServe(":3000", nil)
}
