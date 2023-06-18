#!/bin/bash
# 下载cni二进制文件
wget -O cni-plugins-linux-amd64-v1.1.1.tgz https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz

# 创建临时目录，用于解压
mkdir  ./tmp

# 解压cni二进制文件
tar -Czxvf ./tmp cni-plugins-linux-amd64-v1.1.1.tgz

# 创建目标机器cni目录
ssh root@10.101.1.68 "mkdir -p /opt/cni/bin"

# 复制cni到目标机器cni目录
scp tmp/* root@10.101.1.68:/opt/cni/bin