#!/bin/bash
# 下载containerd二进制文件
wget -O containerd-1.6.6-linux-amd64.tar.gz https://github.com/containerd/containerd/releases/download/v1.6.6/containerd-1.6.6-linux-amd64.tar.gz

# 解压containerd二进制文件
tar zxvf containerd-1.6.6-linux-amd64.tar.gz

# 复制到需要安装node节点的机器，也是安装containerd的节点
scp bin/* root@10.101.1.68:/usr/local/bin