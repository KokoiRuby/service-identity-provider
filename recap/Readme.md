## Service-Identity-Provider

TODO...

### Quick Start

TODO...

### Recap

A quickstart of deploying a client & server (written in [Go](https://go.dev/)) in mTLS using [Docker Compose](https://docs.docker.com/compose/).

#### IP Plan

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
$ cd recap && docker compose up
```

3. Verify

```bash
$ docker logs service-provider
$ docker logs service-consumer
```

#### Rev

v1.1 Remove keypair from image during build phase, use docker volume to mount.

v1.0 Keypair included in image during build phase.

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













