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
const listenAddr = ":3000"

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(200)
			w.Write([]byte(":-)"))
		} else {
			w.WriteHeader(404)
			w.Write([]byte("Nothing here..."))
		}
	})

	klog.Infof("Listening for webhook requests on %s", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}
