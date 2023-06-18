
wget https://github.com/coreos/etcd/releases/download/v3.4.14/etcd-v3.4.14-linux-amd64.tar.gz

tar -xvf etcd-v3.4.14-linux-amd64.tar.gz

scp etcd-v3.4.14-linux-amd64/etcd* root@10.101.1.73:/usr/local/bin

scp etcd-v3.4.14-linux-amd64/etcd* root@10.101.1.64:/usr/local/bin

scp etcd-v3.4.14-linux-amd64/etcd* root@10.101.1.67:/usr/local/bin