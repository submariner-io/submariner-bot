module github.com/submariner-io/pr-brancher-webhook

go 1.12

require (
	github.com/go-playground/webhooks v5.17.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/sethvargo/go-password v0.2.0
	golang.org/x/crypto v0.0.0-20191205180655-e7c4368fe9dd
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.1
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/klog v1.0.0

)

replace gopkg.in/src-d/go-git.v4 v4.13.1 => github.com/src-d/go-git v0.0.0-20190801152248-0d1a009cbb60
