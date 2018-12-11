#!/bin/bash
cd /etc/kubernetes/
sudo rm *.conf
sudo minikube delete && sudo rm -rf ~/.minikube && sudo rm -rf ~/.kube