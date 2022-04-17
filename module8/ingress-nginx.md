### 1.install-helm

```shell
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```

### 2.install-ingress-nginx

```shell
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx --create-namespace --namespace ingress

docker pull anjia0532/google-containers.ingress-nginx.controller:v1.1.3
docker tag anjia0532/google-containers.ingress-nginx.controller:v1.1.3 k8s.gcr.io/ingress-nginx/controller:v1.1.3
docker images | grep $(echo k8s.gcr.io/ingress-nginx/controller:v1.1.3|awk -F':' '{print $1}')

docker pull anjia0532/google-containers.ingress-nginx.kube-webhook-certgen:v1.1.1
docker tag anjia0532/google-containers.ingress-nginx.kube-webhook-certgen:v1.1.1 k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1
docker images | grep $(echo k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1|awk -F':' '{print $1}')


[root@k8s-master ~]# kubectl get pod -n ingress 
NAME                                   READY   STATUS              RESTARTS   AGE
ingress-nginx-admission-create-kxzfx   0/1     Terminating         0          7m37s
ingress-nginx-admission-create-qrls6   0/1     ContainerCreating   0          3s
[root@k8s-master ~]# kubectl edit pod -n ingress ingress-nginx-admission-create-kxzfx
Error from server (NotFound): pods "ingress-nginx-admission-create-kxzfx" not found
[root@k8s-master ~]# kubectl get pod -n ingress 
NAME                                   READY   STATUS         RESTARTS   AGE
ingress-nginx-admission-create-qrls6   0/1     ErrImagePull   0          26s
[root@k8s-master ~]# kubectl edit pod -n ingress ingress-nginx-admission-create-qrls6
pod/ingress-nginx-admission-create-qrls6 edited
[root@k8s-master ~]# kubectl get pod -n ingress -w
NAME                                        READY   STATUS              RESTARTS   AGE
ingress-nginx-admission-patch-nk6cm         0/1     ContainerCreating   0          3s
ingress-nginx-controller-69fbbf9f9c-wcjgg   0/1     ContainerCreating   0          3s
^C[root@k8s-master ~]# kubectl edit pod ingress-nginx-admission-patch-nk6cm -n ingress 
pod/ingress-nginx-admission-patch-nk6cm edited
^C[root@k8s-master ~]# kubectl get pod -n ingress
NAME                                        READY   STATUS             RESTARTS   AGE
ingress-nginx-controller-69fbbf9f9c-wcjgg   0/1     ImagePullBackOff   0          59s
[root@k8s-master ~]# kubectl edit pod -n ingress ingress-nginx-controller-69fbbf9f9c-wcjgg
pod/ingress-nginx-controller-69fbbf9f9c-wcjgg edited
[root@k8s-master ~]# kubectl get pod -n ingress
NAME                                        READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-69fbbf9f9c-wcjgg   0/1     Running   0          100s
[root@k8s-master ~]# kubectl get pod -n ingress
NAME                                        READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-69fbbf9f9c-wcjgg   0/1     Running   0          102s


[root@k8s-master ~]# kubectl delete -A ValidatingWebhookConfiguration ingress-nginx-admission
validatingwebhookconfiguration.admissionregistration.k8s.io "ingress-nginx-admission" deleted
[root@k8s-master ~]# kubectl apply -f httpserver_ingress.yaml 
ingress.networking.k8s.io/httpserver-80 created
```

