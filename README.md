# pr-brancher-webhook


# setup
```bash
export NS=pr-brancher-webhook
kubectl create namespace $NS
# create a bot account with permission to your repos, and create a token in your bot account: https://github.com/settings/tokens
kubectl create -n $NS secret generic pr-brancher-secrets --from-file=ssh_pk=./id_rsa --from-literal=githubToken=$GITHUB_TOKEN
kubectl apply -n $NS -f deployment/role.yml
kubectl apply -n $NS -f deployment/deployment.yaml
kubectl apply -n $NS -f deployment/service.yml

```

# update image
```bash
kubectl rollout restart deployment/pr-brancher
```
