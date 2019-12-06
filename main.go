package main

import (
    "fmt"
    "net/http"
    "github.com/go-playground/webhooks/github"
)

const (
    path = "/webhooks"
)

func main() {
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
        switch payload.(type) {

            case github.PullRequestPayload:
                pullRequest := payload.(github.PullRequestPayload)
                // Do whatever you want from here...
                fmt.Printf("%+v\n", pullRequest)
                // pullRequest.Number
                // pullRequest.Repository.SSHURL
                // pullRequest.PullRequest.Head.Repo.SSHURL
                // pullRequest.PullRequest.Base.Label
                // pullRequest.Sender.Login

                branchName := fmt.Sprintf("z_pr%d/%s/%s",
                    pullRequest.Number,
                    pullRequest.Sender.Login,
                    pullRequest.PullRequest.Head.Ref)

                fmt.Printf("branch-name: %s\n", branchName)
        }
    })
    http.ListenAndServe(":3000", nil)
}
