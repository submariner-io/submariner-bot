package main

import (
	"net/http"

	"github.com/go-playground/webhooks/v6/github"
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
		payload, err := hook.Parse(r, handler.EventsToHandle()...)
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
			_, err := w.Write([]byte("An error happened: " + err.Error()))
			if err != nil {
				klog.Errorf("Failed to write response: %s", err)
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(200)
			_, err := w.Write([]byte(":-)"))
			if err != nil {
				klog.Errorf("Failed to write response: %s", err)
			}
		} else {
			w.WriteHeader(404)
			_, err := w.Write([]byte("Nothing here..."))
			if err != nil {
				klog.Errorf("Failed to write response: %s", err)
			}
		}
	})

	klog.Infof("Listening for webhook requests on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		klog.Fatalf("Can't start listening for requests: %s", err)
	}
}
