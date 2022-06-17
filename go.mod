module github.com/submariner-io/pr-brancher-webhook

go 1.12

require (
	github.com/go-playground/webhooks v5.17.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/sethvargo/go-password v0.2.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
	k8s.io/client-go v0.24.2
	k8s.io/klog v1.0.0

)

replace gopkg.in/src-d/go-git.v4 v4.13.1 => github.com/src-d/go-git v0.0.0-20190801152248-0d1a009cbb60
