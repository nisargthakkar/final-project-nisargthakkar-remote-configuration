#!/bin/bash
set -E
if ! sudo minikube status > /dev/null; then 
    sudo minikube start --vm-driver=none --kubernetes-version v1.13.0 # --bootstrapper=localkube
else
    echo 'Minikube is already running. Use `sudo minikube stop` to stop it if necessary'
		echo 'If you wish to reset minikube, run `sudo minikube delete && sudo rm -rf ~/.minikube && sudo rm -rf ~/.kube`'
fi
