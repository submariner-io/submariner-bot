module github.com/submariner-io/pr-brancher-webhook

go 1.12

require (
	github.com/go-playground/webhooks v5.13.0+incompatible
	github.com/src-d/go-git v4.0.0+incompatible // indirect
	golang.org/x/crypto v0.0.0-20191205180655-e7c4368fe9dd
	gopkg.in/src-d/go-git.v4 v4.13.1
)

replace gopkg.in/src-d/go-git.v4 v4.13.1 => github.com/src-d/go-git v0.0.0-20190801152248-0d1a009cbb60
