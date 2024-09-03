### Recap

A quickstart of deploying a client & server (written in [Go](https://go.dev/)) in mTLS using [Docker Compose](https://docs.docker.com/compose/).

| Instance         | IP            | Port                                           | HostPort                                       |
| ---------------- | ------------- | ---------------------------------------------- | ---------------------------------------------- |
| service-provider | 192.168.60.10 | 8080 (no TLS)<br />8088 (TLS)<br />8089 (mTLS) | 8080 (no TLS)<br />8088 (TLS)<br />8089 (mTLS) |
| service-consumer | 192.168.60.20 | N/A                                            | N/A                                            |

1. Create network

```bash
$ docker network create --subnet=192.168.0.0/16 local
```

2. Docker Compose up

```bash
$ cd recap
$ docker compose up
```

3. Verify

```bash
$ docker logs service-provider
$ docker logs service-consumer
```

#### Helm

A quickstart of deploying a client & server (written in [Go](https://go.dev/)) in mTLS using [Helm](https://helm.sh/) chart.

Note: if u don't have a K8s cluster, try [kind](https://kind.sigs.k8s.io/).

```bash
$ cd recap
$ helm install my-mtls-demo mtls-demo-1.0.0.tgz
```

#### Rev

**v1.2.1 Load keypair by names defined in tls type secret; service-consumer requests by service name instead of fixed IP.**

v1.2.0 Separate ca & service keypair to different directories to adapt volumnMount mountPath uniqueness in pod spec.

v1.1.0 Remove keypair from image during build phase, use docker volume to mount.

v1.0.0 Keypair included in image during build phase.

#### Cheatsheet

```bash
$ docker build -t service-provider -f Dockerfile .

$ docker run -d \
	-p 8080:8080 -p 8088:8088 -p 8089:8089 \
	--net local --ip 192.168.60.10 \
	--hostname service-provider \
	--name service-provider \
	service-provider 

$ docker tag service-provider yukanyan/service-provider:v1.1

$ docker push yukanyan/service-provider:v1.1
```

```bash
$ docker build -t service-consumer -f Dockerfile .

$ docker run -d \
	--hostname service-consumer \
	--net local --ip 192.168.60.20 \
	--name service-consumer \
	service-consumer

$ docker tag service-consumer yukanyan/service-consumer:v1.1

$ docker push yukanyan/service-consumer:v1.1
```

```bash
$ kubectl create secret tls service-provider-keypair \
	--cert=./certs/server-chain.pem \
	--key=./certs/server-key.pem \
	--dry-run=client \
	-o yaml > ./manifests/secret-service-provider.yaml

$ kubectl create secret tls service-consumer-keypair \
	--cert=./certs/client-chain.pem \
	--key=./certs/client-key.pem \
	--dry-run=client \
	-o yaml > ./manifests/secret-service-consumer.yaml
	
$ kubectl create secret generic ca-cert \
	--from-file=ca.pem=./certs/ca.pem \
	--dry-run=client \
	-o yaml > ./manifests/secret-ca.yaml
```













