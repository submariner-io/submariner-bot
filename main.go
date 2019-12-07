package main

import (
	"net/http"

	"github.com/go-playground/webhooks/github"
	"k8s.io/klog"

	"github.com/submariner-io/pr-brancher-webhook/pkg/config"
	"github.com/submariner-io/pr-brancher-webhook/pkg/handler"
)

const (
	path = "/webhooks"
)

func main() {

	webhookSecret, err := config.GetWebhookSecret()
	if err != nil {
		klog.Errorf("Error while trying to retrieve webhook secret: %s", err)
		klog.Fatalf("The webhook secret can be provided as env var via %s", config.WebhookSecretEnvVar)
	}

	hook, _ := github.New(github.Options.Secret(webhookSecret))

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
