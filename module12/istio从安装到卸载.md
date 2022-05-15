## istio安装

```shell
curl -L https://istio.io/downloadIstio | sh -
Istio 1.13.3 Download Complete!

istioctl install --set profile=demo -y
✔ Istio core installed
- Processing resources for Istiod.
✔ Istiod installed
✔ Ingress gateways installed
✔ Egress gateways installed
✔ Installation complete       
```

> label ns

```shell
kubectl label ns default istio-injection=enabled
namespace/default labeled

kubectl get ns -L istio-injection
default           Active        3d      enabled

```

> test

```shell
kubectl -n default create deployment my-nginx --image=nginx
```

