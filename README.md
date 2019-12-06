# pr-brancher-webhook


# setup
```bash
export NS=pr-brancher-webhook
kubectl create namespace $NS
kubectl create -n $NS secret generic pr-brancher-secrets --from-file=ssh_pk=./id_rsa

```
