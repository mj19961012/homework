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

> 签发证书

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=secsmart /CN=*.secsmart.network' -keyout secsmart.network.key -out secsmart.network.crt

kubectl create -n istio-system secret tls wildcard-credential --key=secsmart.network.key --cert=secsmart.network.crt
```

