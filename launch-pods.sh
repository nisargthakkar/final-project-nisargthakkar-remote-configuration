#!/bin/bash
kubectl create -f ./k8s-configs/config-management-server.yml
kubectl create -f ./k8s-configs/getconfigbash.yml
kubectl create -f ./k8s-configs/configpollpythonweb.yml
kubectl create -f ./k8s-configs/configuration-update-client.yml
kubectl create -f ./k8s-configs/configuration-management-service-ingress.yml