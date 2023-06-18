#!/bin/bash

wget https://github.com/kubernetes/kubernetes/releases/download/v1.25.0/kubernetes.tar.gz

tar -xzvf kubernetes.tar.gz

cd kubernetes

./cluster/get-kube-binaries.sh

tar -xvf server/kubernetes-server-linux-amd64.tar.gz

# 复制到目标机器
scp -r kubernetes/server/bin/{kubectl,kube-proxy,kubelet} root@10.101.1.68:/usr/local/bin/