#!/bin/sh
set -e
set -x

docker build -t quay.io/submariner/submariner-bot:dev .
docker push quay.io/submariner/submariner-bot:dev

kubectl -n pr-brancher-webhook delete pods -l app=submariner-bot
sleep 1


while [[ $(kubectl get pods -n pr-brancher-webhook -l app=submariner-bot -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
	echo "waiting for pod" && sleep 1
done
kubectl get pods -n pr-brancher-webhook

POD=$(kubectl get pod -n pr-brancher-webhook -l app=submariner-bot -o jsonpath="{.items[0].metadata.name}")
kubectl logs -n pr-brancher-webhook -f $POD


