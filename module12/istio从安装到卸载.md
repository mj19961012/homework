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

> 部署应用

```shell
kubectl apply -f httpserver_dep0.yaml
kubectl apply -f httpserver_dep1.yaml
kubectl apply -f httpserver_dep2.yaml

kubectl apply -f httpserver_svc0.yaml
kubectl apply -f httpserver_svc1.yaml
kubectl apply -f httpserver_svc2.yaml
```

> 部署gateway

```shell
kubectl apply -f httpserver-gw.yaml 
kubectl apply -f httpserver-vs.yaml
```



> 测试应用

```shell
 kubectl get svc -n istio-system 
NAME                   TYPE           CLUSTER-IP        EXTERNAL-IP   PORT(S)                                                   
istio-egressgateway    ClusterIP      192.105.236.38    <none>        80/TCP,443/TCP                                             
istio-ingressgateway   LoadBalancer   192.108.32.84     <pending>     15021:32710/TCP,80:31212/TCP,443:31920/TCP,31400:32131/TCP,15443:31061/TCP
istiod                 ClusterIP      192.102.162.167   <none>        15010/TCP,15012/TCP,443/TCP,15014/TCP            

curl --resolve test.secsmart.network:443:192.108.32.84 https://test.secsmart.network/healthz -H "Custom-header: hello" -v -k

* Added test.secsmart.network:443:192.108.32.84 to DNS cache
* About to connect() to test.secsmart.network port 443 (#0)
*   Trying 192.108.32.84...
* Connected to test.secsmart.network (192.108.32.84) port 443 (#0)
* Initializing NSS with certpath: sql:/etc/pki/nssdb
* skipping SSL peer certificate verification
* SSL connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate:
* 	subject: CN=*.secsmart.network,O="secsmart "
* 	start date: May 15 13:16:47 2022 GMT
* 	expire date: May 15 13:16:47 2023 GMT
* 	common name: *.secsmart.network
* 	issuer: CN=*.secsmart.network,O="secsmart "
> GET /healthz HTTP/1.1
> User-Agent: curl/7.29.0
> Host: test.secsmart.network
> Accept: */*
> Custom-header: hello
> 
< HTTP/1.1 200 OK
< date: Sun, 15 May 2022 13:52:09 GMT
< content-length: 3
< content-type: text/plain; charset=utf-8
< x-envoy-upstream-service-time: 11
< server: istio-envoy
< 
ok
* Connection #0 to host test.secsmart.network left intact

```

> jaeger test

```shell
istioctl dashboard jaeger --address 0.0.0.0

#浏览器访问
```

![image-20220515221549175](C:\Users\Jey\AppData\Roaming\Typora\typora-user-images\image-20220515221549175.png)