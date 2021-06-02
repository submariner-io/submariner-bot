# submariner-bot

## How it works

submariner-bot listens for webhook events on port 3000 over http. Those events
are in the github webhook event format
described [here](https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads)

We use a [library](https://github.com/go-playground/webhooks/tree/master/github) that provides a good interface to handle those events
which are handled [here](https://github.com/submariner-io/submariner-bot/blob/devel/pkg/handler/handler.go):

## Developing and testing locally

You need Go to run and test submariner-bot locally:

```bash
export GO111MODULE=on
export GITHUB_TOKEN=<a token, you can create one in https://github.com/settings/tokens>
export WEBHOOK_SECRET=your-random-phrase # this is a password for anybody accessing submariner-bot
export SSH_PK=/your/ssh/private/key
go run pkg/main/main.go  # or just do it from your favorite IDE
```

Then you can push events to localhost:3000.
There are probably tools to simulate events but if you want to simulate it on your host you can create a public endpoint with
[ngrok](https://ngrok.com/).
You should sign up for the free version so your tunnels will not be time-limited.

On another terminal (keep it open):

```bash
$ ngrok authtoken <this is provided when you sign-up, optional>
$ ngrok http  3000
ngrok by @inconshreveable                                                    (Ctrl+C to quit)

Session Status                online
Account                       your-email (Plan: Free)
Version                       2.3.40
Region                        United States (us)
Web Interface                 http://127.0.0.1:4040
Forwarding                    http://f05c3eb7fe60.ngrok.io -> http://localhost:3000
Forwarding                    https://f05c3eb7fe60.ngrok.io -> http://localhost:3000

Connections                   ttl     opn     rt1     rt5     p50     p90
                              0       0       0.00    0.00    0.00    0.00
```

At this point you'll need a submariner admin to [setup your webhook](https://github.com/organizations/submariner-io/settings/hooks/new)
while in development with the forwarding address `https://f05c3eb7fe60.ngrok.io` and your WEBHOOK_SECRET.

Try not to close ngrok to avoid the webhook URL from changing (starting a new session)

## setup

```bash
export NS=pr-brancher-webhook
kubectl create namespace $NS
# create a bot account with permission to your repos, and create a token in your bot account: https://github.com/settings/tokens
kubectl create -n $NS secret generic pr-brancher-secrets --from-file=ssh_pk=./id_rsa --from-literal=githubToken=$GITHUB_TOKEN
kubectl apply -n $NS -f deployment/role.yml
kubectl apply -n $NS -f deployment/deployment.yaml
kubectl apply -n $NS -f deployment/service.yml

```

### setup with https/letsencrypt

```bash
kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.12.0/cert-manager.yaml
kubectl apply -f deployment/letsencrypt-prod-issuer.yaml # you may need to edit the class in the yaml based on your ingress
```

## update image

```bash
kubectl rollout restart deployment/pr-brancher
```
