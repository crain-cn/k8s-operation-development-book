#!/bin/bash
# 下载生成证书主程序并设置执行权限
wget -O /usr/local/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
chmod +x /usr/local/bin/cfssl
 
# 下载利用Json生成证书工具并设置执行权限
wget -O /usr/local/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
chmod +x /usr/local/bin/cfssljson

# 下载查看证书信息的工具并设置执行权限
wget -O /usr/local/bin/cfssl-certinfo https://pkg.cfssl.org/R1.2/cfssl-certinfo_linux-amd64
chmod +x /usr/local/bin/cfssl-certinfo