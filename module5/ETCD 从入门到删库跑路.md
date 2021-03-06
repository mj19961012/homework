### ETCD 从入门到删库跑路

#### 1.安装 NFS

```shell
yum -y install nfs-utils
```

#### 2.配置 NFS

```shell
mkdir -p /nfs/data/
chmod -R 777 /nfs/data
vim /etc/exports #添加以下内容 /nfs/data *(rw,sync,all_squash,insecure)

cat /etc/exports
/nfs/data *(rw,sync,all_squash,insecure)

exportfs -r
systemctl restart rpcbind && systemctl enable rpcbind
systemctl restart nfs && systemctl enable nfs
rpcinfo -p localhost

showmount -e

curl -skSL https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/deploy/install-driver.sh | bash -s master --
Installing NFS CSI driver, version: master ...
serviceaccount/csi-nfs-controller-sa created
serviceaccount/csi-nfs-node-sa created
clusterrole.rbac.authorization.k8s.io/nfs-external-provisioner-role created
clusterrolebinding.rbac.authorization.k8s.io/nfs-csi-provisioner-binding created
csidriver.storage.k8s.io/nfs.csi.k8s.io unchanged
deployment.apps/csi-nfs-controller configured
daemonset.apps/csi-nfs-node configured
NFS CSI driver installed successfully.

cat nfs-sc.yaml
```

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: nfs-csi
provisioner: nfs.csi.k8s.io
parameters:
  server: 172.16.49.20
  share: /home/nfs/data
reclaimPolicy: Delete
volumeBindingMode: Immediate
mountOptions:
  - hard
  - nfsvers=4.1
```

```shell
kubectl apply -f nfs-sc.yaml 
```



#### 3.配置 etcd

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: etcd
  name: etcd
spec:
  replicas: 3
  selector:
    matchLabels:
      app: etcd
  serviceName: etcd
  template:
    metadata:
      labels:
        app: etcd
    spec:
      containers:
        - name: etcd
          image: quay.io/coreos/etcd:v3.3.8
          imagePullPolicy: IfNotPresent
          ports:
          - containerPort: 2380
            name: peer
            protocol: TCP
          - containerPort: 2379
            name: client
            protocol: TCP
          env:
          - name: INITIAL_CLUSTER_SIZE
            value: "3"
          - name: MY_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: SET_NAME
            value: "etcd"
          command:
            - /bin/sh
            - -ec
            - |
              HOSTNAME=$(hostname)

              export ETCDCTL_API=3

              eps() {
                  EPS=""
                  for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                      EPS="${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379"
                  done
                  echo ${EPS}
              }

              member_hash() {
                  etcdctl member list | grep -w "$HOSTNAME" | awk '{ print $1}' | awk -F "," '{ print $1}'
              }

              initial_peers() {
                  PEERS=""
                  for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                    PEERS="${PEERS}${PEERS:+,}${SET_NAME}-${i}=http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380"
                  done
                  echo ${PEERS}
              }

              # etcd-SET_ID
              SET_ID=${HOSTNAME##*-}

              # adding a new member to existing cluster (assuming all initial pods are available)
              if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
                  # export ETCDCTL_ENDPOINTS=$(eps)
                  # member already added?

                  MEMBER_HASH=$(member_hash)
                  if [ -n "${MEMBER_HASH}" ]; then
                      # the member hash exists but for some reason etcd failed
                      # as the datadir has not be created, we can remove the member
                      # and retrieve new hash
                      echo "Remove member ${MEMBER_HASH}"
                      etcdctl --endpoints=$(eps) member remove ${MEMBER_HASH}
                  fi

                  echo "Adding new member"

                  echo "etcdctl --endpoints=$(eps) member add ${HOSTNAME} --peer-urls=http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380"
                  etcdctl member --endpoints=$(eps) add ${HOSTNAME} --peer-urls=http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 | grep "^ETCD_" > /var/run/etcd/new_member_envs

                  if [ $? -ne 0 ]; then
                      echo "member add ${HOSTNAME} error."
                      rm -f /var/run/etcd/new_member_envs
                      exit 1
                  fi

                  echo "==> Loading env vars of existing cluster..."
                  sed -ie "s/^/export /" /var/run/etcd/new_member_envs
                  cat /var/run/etcd/new_member_envs
                  . /var/run/etcd/new_member_envs

                  echo "etcd --name ${HOSTNAME} --initial-advertise-peer-urls ${ETCD_INITIAL_ADVERTISE_PEER_URLS} --listen-peer-urls http://${POD_IP}:2380 --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 --data-dir /var/run/etcd/default.etcd --initial-cluster ${ETCD_INITIAL_CLUSTER} --initial-cluster-state ${ETCD_INITIAL_CLUSTER_STATE}"

                  exec etcd --listen-peer-urls http://${POD_IP}:2380 \
                      --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \
                      --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \
                      --data-dir /var/run/etcd/default.etcd
              fi

              for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                  while true; do
                      echo "Waiting for ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local to come up"
                      ping -W 1 -c 1 ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local > /dev/null && break
                      sleep 1s
                  done
              done

              echo "join member ${HOSTNAME}"
              # join member
              exec etcd --name ${HOSTNAME} \
                  --initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 \
                  --listen-peer-urls http://${POD_IP}:2380 \
                  --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \
                  --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \
                  --initial-cluster-token etcd-cluster-1 \
                  --data-dir /var/run/etcd/default.etcd \
                  --initial-cluster $(initial_peers) \
                  --initial-cluster-state new
          lifecycle:
            preStop:
              exec:
                command:
                  - /bin/sh
                  - -ec
                  - |
                    HOSTNAME=$(hostname)

                    member_hash() {
                        etcdctl member list | grep -w "$HOSTNAME" | awk '{ print $1}' | awk -F "," '{ print $1}'
                    }

                    eps() {
                        EPS=""
                        for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                            EPS="${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379"
                        done
                        echo ${EPS}
                    }

                    export ETCDCTL_ENDPOINTS=$(eps)
                    SET_ID=${HOSTNAME##*-}

                    # Removing member from cluster
                    if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
                        echo "Removing ${HOSTNAME} from etcd cluster"
                        etcdctl member remove $(member_hash)
                        if [ $? -eq 0 ]; then
                            # Remove everything otherwise the cluster will no longer scale-up
                            rm -rf /var/run/etcd/*
                        fi
                    fi
          volumeMounts:
          - mountPath: /var/run/etcd
            name: datadir
  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      accessModes:
      - "ReadWriteOnce"
      resources:
        requests:
          # upstream recommended max is 700M
          storage: 1Gi

```

#### 4.创建 etcd

```shell
kubectl apply -f sts.yaml

kubectl get pod
NAME     READY   STATUS    RESTARTS   AGE
etcd-0   0/1     Pending   0          53s

kubectl describe pod etcd-0
Events:
  Type     Reason            Age   From               Message
  ----     ------            ----  ----               -------
  Warning  FailedScheduling  110s  default-scheduler  0/3 nodes are available: 3 pod has unbound immediate PersistentVolumeClaims.
  Warning  FailedScheduling  49s   default-scheduler  0/3 nodes are available: 3 pod has unbound immediate PersistentVolumeClaims.
  

```

