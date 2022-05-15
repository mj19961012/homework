### K8S 从安装到放弃

#### 1.安装 gcc

```shell
yum -y install gcc
yum -y install gcc-c++
```

#### 2.卸载旧版本 docker

```shell
yum remove docker \
                  docker-client \
                  docker-client-latest \
                  docker-common \
                  docker-latest \
                  docker-latest-logrotate \
                  docker-logrotate \
                  docker-engine
```

#### 3.安装 yum-utils软件包

```shell
yum install -y yum-utils
```

#### 4.设置镜像仓库源

```shell
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    
yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
```

#### 5.更新yum 索引

```shell
yum clean all
yum makecache fast
yum update --allowerasing 
```

#### 6.安装 docker-ce

```shell
yum install -y docker-ce docker-ce-cli containerd.io
```

#### 7.安装 k8s

```shell
cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
exclude=kubelet kubeadm kubectl
EOF

# 将 SELinux 设置为 permissive 模式（相当于将其禁用）
setenforce 0
sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
swapoff -a

#允许 iptables 检查桥接流量
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system
#install kubelet kubeadm kubectl
yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes

kubeadm config images list

yum install -y kubelet-1.19.13 kubeadm-1.19.13 kubectl-1.19.13 --disableexcludes=kubernetes

systemctl enable --now kubelet
#修改 docker cgroup 驱动为systemd
mkdir /etc/docker
cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ],
  "insecure-registries":["172.16.49.20:5000"],
  "data-root": "/data/docker"
}
EOF
systemctl daemon-reload
systemctl restart docker

#修改hostsname
echo '172.16.39.202 k8s-master' >> /etc/hosts
echo '172.16.49.20 k8s-node1' >> /etc/hosts
echo '172.16.49.21 k8s-node2' >> /etc/hosts
systemctl stop firewalld && systemctl disable firewalld


kubeadm init \
--apiserver-advertise-address=172.16.39.202 \
--image-repository registry.aliyuncs.com/google_containers \
--kubernetes-version v1.23.4 \
--service-cidr=192.96.0.0/12,2001:db8:42:1::/112 \
--pod-network-cidr=192.168.0.0/16,2001:db8:42:0::/56

kubeadm init --pod-network-cidr=10.244.0.0/16,2001:db8:42:0::/56 --service-cidr=10.96.0.0/16,2001:db8:42:1::/112
#podnetwork
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
https://kubernetes.io/docs/concepts/cluster-administration/addons/

#kubeadm init报错
[ERROR FileContent--proc-sys-net-bridge-bridge-nf-call-iptables]: /proc/sys/net/bridge/bridge-nf-call-iptables does not exist
#解决办法
vi /etc/sysctl.conf

在/etc/sysctl.conf中添加：

net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1

#保存后执行 sysctl -p
sysctl: cannot stat /proc/sys/net/bridge/bridge-nf-call-ip6tables: No such file or directory
sysctl: cannot stat /proc/sys/net/bridge/bridge-nf-call-iptables: No such file or directory

#报错信息如上执行 modprobe br_netfilter
#在执行 sysctl -p
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1

#kubeadm init报错
[ERROR FileContent--proc-sys-net-ipv4-ip_forward]: /proc/sys/net/ipv4/ip_forward contents are not set to 1
#解决办法
sysctl -w net.ipv4.ip_forward=1

#要使非 root 用户可以运行 kubectl，请运行以下命令， 它们也是 kubeadm init 输出的一部分：
rm -rf $HOME/.kube
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
#如果你是 root 用户，则可以运行：
export KUBECONFIG=/etc/kubernetes/admin.conf

kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml

kubeadm join 172.16.39.202:6443 --token djq1fd.iuun2whpgz166n4t \
	--discovery-token-ca-cert-hash sha256:27325aa2364549db96da506015a396394d4fe100ff5559ead246d1ec8758efe2

#kubeadm join 报错
kubeadm join命令，将node加入master时，出现error execution phase preflight: couldn't validate the identity of the API Server: abort connecting to API servers after timeout
[preflight] Running pre-flight checks
error execution phase preflight: couldn't validate the identity of the API Server: Get "https://172.16.39.202:6443/api/v1/namespaces/kube-public/configmaps/cluster-info?timeout=10s": x509: certificate has expired or is not yet valid:
#解决：重新生成新token
# kubeadm token create
djq1fd.iuun2whpgz166n4t
# openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | sed 's/^.* //'
27325aa2364549db96da506015a396394d4fe100ff5559ead246d1ec8758efe2
kubeadm join 172.16.39.202:6443 --token djq1fd.iuun2whpgz166n4t \
	--discovery-token-ca-cert-hash sha256:27325aa2364549db96da506015a396394d4fe100ff5559ead246d1ec8758efe2


kubectl label nodes k8s-node1 node-role.kubernetes.io/worker=
kubectl label nodes k8s-node2 node-role.kubernetes.io/worker=

kubectl apply -f kube-flannel.yml

#启用 shell 自动补全功能
yum install bash-completion
#上述命令将创建文件 /usr/share/bash-completion/bash_completion，它是 bash-completion 的主脚本。 依据包管理工具的实际情况，你需要在 ~/.bashrc 文件中手工导入此文件。要查看结果，请重新加载你的 shell，并运行命令 
type _init_completion
#如果命令执行成功，则设置完成，否则将下面内容添加到文件 ~/.bashrc 中：
source /usr/share/bash-completion/bash_completion
#重新加载 shell，再输入命令 type _init_completion 来验证 bash-completion 的安装状态。

#启动 kubectl 自动补全功能
#在文件 ~/.bashrc 中导入（source）补全脚本
echo 'source <(kubectl completion bash)' >>~/.bashrc
#或者将补全脚本添加到目录 /etc/bash_completion.d 中
kubectl completion bash >/etc/bash_completion.d/kubectl

#如果 kubectl 有关联的别名，你可以扩展 shell 补全来适配此别名
echo 'alias k=kubectl' >>~/.bashrc
echo 'complete -F __start_kubectl k' >>~/.bashrc
```

#### 8.卸载 k8s

```shell
# 适当的凭证与控制平面节点通信，运行：
kubectl drain <node name> --delete-local-data --force --ignore-daemonsets
# 卸载服务
kubeadm reset
# 重置过程不会重置或清除 iptables 规则或 IPVS 表。如果你希望重置 iptables，则必须手动进行：
iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X
# 如果要重置 IPVS 表，则必须运行以下命令：
ipvsadm -C
# 现在删除节点：
kubectl delete node <node name>
# 删除rpm包
rpm -qa|grep kube*|xargs rpm --nodeps -e
```

#### 9.安装harbor

1.下载安装包

```shell
wget https://github.com/goharbor/harbor/releases/download/v2.5.0-rc3/harbor-offline-installer-v2.5.0-rc3.tgz
```

2.安装 docker && docker-compose

```shell
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
docker-compose --version
```

3.配置证书

```shell
#生成证书颁发机构证书
#1.生成CA证书私钥。
openssl genrsa -out ca.key 4096
#2.生成CA证书。
openssl req -x509 -new -nodes -sha512 -days 3650 \
 -subj "/C=CN/ST=Beijing/L=Beijing/O=example/OU=Personal/CN=yourdomain.com.cn" \
 -key ca.key \
 -out ca.crt
 
#生成服务器证书
#1.生成私钥。
openssl genrsa -out secsmart.com.cn.key 4096
#2.生成证书签名请求 (CSR)。
openssl req -sha512 -new -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=secsmart/OU=networkdlp/CN=reg.secsmart.com.cn" -key secsmart.com.cn.key -out secsmart.com.cn.csr
#3.生成一个x509 v3扩展名文件。
cat > v3.ext <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1=secsmart.com.cn
DNS.2=secsmart.com
DNS.3=secsmart
DNS.4=hostname
EOF
#4.使用v3.ext文件为您的Harbor主机生成证书。
openssl x509 -req -sha512 -days 3650 \
    -extfile v3.ext \
    -CA ca.crt -CAkey ca.key -CAcreateserial \
    -in secsmart.com.cn.csr \
    -out secsmart.com.cn.crt
```

4.向Harbor和Docker提供证书

```shell
#1.将服务器证书和密钥复制到Harbor主机上的certficates文件夹中。
cp secsmart.com.cn.crt /data/cert/
cp secsmart.com.cn.key /data/cert/
#2.将yourdomain.com.crt转换为yourdomain.com.cert，供Docker使用
openssl x509 -inform PEM -in secsmart.com.cn.crt -out secsmart.com.cn.cert
#3.将服务器证书、密钥和CA文件复制到Harbor主机上的Docker证书文件夹中。您必须先创建相应的文件夹。
cp secsmart.com.cn.cert /etc/docker/certs.d/secsmart.com.cn/
cp secsmart.com.cn.key /etc/docker/certs.d/secsmart.com.cn/
cp ca.crt /etc/docker/certs.d/secsmart.com.cn/
#如果您将默认nginx端口443映射到其他端口，请创建文件夹/etc/docker/certs.d/yourdomain.com:port或/etc/docker/certs.d/harbor_IP:port。
#4.重新启动Docker引擎。
systemctl restart docker
```


