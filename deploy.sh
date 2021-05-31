#!/bin/sh
set -e
set -x

NS=pr-brancher-webhook

docker build -t quay.io/submariner/submariner-bot:dev .
docker tag quay.io/submariner/submariner-bot:dev quay.io/submariner/submariner-bot:$(git describe --tags)
docker push quay.io/submariner/submariner-bot:dev
docker push quay.io/submariner/submariner-bot:$(git describe --tags)

kubectl -n $NS delete pods -l app=submariner-bot

sleep 1


while [[ $(kubectl get pods -n $NS -l app=submariner-bot -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
	echo "waiting for pod" && sleep 1
done
kubectl get pods -n $NS

POD=$(kubectl get pod -n $NS -l app=submariner-bot -o jsonpath="{.items[0].metadata.name}")
kubectl logs -n $NS -f $POD


