#!/bin/bash
wget https://github.com/kubernetes/kubernetes/releases/download/v1.25.0/kubernetes.tar.gz

tar -xzvf kubernetes.tar.gz

cd kubernetes

./cluster/get-kube-binaries.sh

tar -xvf server/kubernetes-server-linux-amd64.tar.gz


scp -r kubernetes/server/bin/{kube-apiserver,kube-controller-manager,kube-scheduler,kubectl,kube-proxy,kubelet} root@10.101.1.73:/usr/local/bin/

scp -r kubernetes/server/bin/{kube-apiserver,kube-controller-manager,kube-scheduler,kubectl,kube-proxy,kubelet} root@10.101.1.64:/usr/local/bin/

scp -r kubernetes/server/bin/{kube-apiserver,kube-controller-manager,kube-scheduler,kubectl,kube-proxy,kubelet} root@10.101.1.67:/usr/local/bin/