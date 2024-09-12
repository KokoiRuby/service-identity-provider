## Helm

### [with Integrated Storage](https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-minikube-raft)

Install

```yaml
# helm-vault-raft-values.yml
server:
   affinity: ""
   ha:
      enabled: true
      raft: 
         enabled: true
         setNodeId: true
         config: |
            cluster_name = "vault-integrated-storage"
            storage "raft" {
               path    = "/vault/data/"
            }

            listener "tcp" {
               address = "[::]:8200"
               cluster_address = "[::]:8201"
               tls_disable = "true"
            }
            service_registration "kubernetes" {}

```

```bash
$ helm install vault hashicorp/vault --values helm-vault-raft-values.yml

# init vault-0
$ kubectl exec vault-0 -- vault operator init \
	-key-shares=1 \
    -key-threshold=1 \
    -format=json > cluster-keys.json
    
# unseal vault-0
$ jq -r ".unseal_keys_b64[]" cluster-keys.json
$ VAULT_UNSEAL_KEY=$(jq -r ".unseal_keys_b64[]" cluster-keys.json)
$ kubectl exec vault-0 -- vault operator unseal $VAULT_UNSEAL_KEY
```

```bash
# join & unseal vault-1 & vault-2
$ kubectl exec -ti vault-1 -- vault operator raft join http://vault-0.vault-internal:8200
$ kubectl exec -ti vault-2 -- vault operator raft join http://vault-0.vault-internal:8200

$ kubectl exec -ti vault-1 -- vault operator unseal $VAULT_UNSEAL_KEY
$ kubectl exec -ti vault-2 -- vault operator unseal $VAULT_UNSEAL_KEY
```

Secrets engine

```bash
$ jq -r ".root_token" cluster-keys.json
$ kubectl exec -ti vault-0 -- /bin/sh

$ vault secrets enable -path=secret kv-v2
$ vault kv put secret/webapp/config username="static-user" password="static-password"
$ vault kv get secret/webapp/config
```

Kubernetes authentication

```bash
$ jq -r ".root_token" cluster-keys.json
$ kubectl exec -ti vault-0 -- /bin/sh

$ vault login
$ vault auth enable kubernetes
$ vault write auth/kubernetes/config \
	kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"

# policy
$ vault policy write webapp - <<EOF
path "secret/data/webapp/config" {
  capabilities = ["read"]
}
EOF

# create role & connect to k8s service account & policy
$ vault write auth/kubernetes/role/webapp \
	bound_service_account_names=vault \
	bound_service_account_namespaces=default \
	policies=webapp \
	ttl=24h
```

Web Application

```yaml
$ cat > deployment01-webapp.yml << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  labels:
    app: webapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      serviceAccountName: vault
      containers:
        - name: app
          image: hashieducation/simple-vault-client:latest
          imagePullPolicy: Always
          env:
            - name: VAULT_ADDR
              value: 'http://vault:8200'
            - name: JWT_PATH
              value: '/var/run/secrets/kubernetes.io/serviceaccount/token'
            - name: SERV
EOF
```

```bash
$ kubectl port-forward \
	$(kubectl get pod -l app=webapp -o jsonpath="{.items[0].metadata.name}") \
    8080:8080

$ curl http://localhost:8080
```

Clean up

```bash
$ kubectl delete -f deployment01-webapp.yml
$ helm uninstall vault
```

### [With Consul](https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-minikube-consul)

[Consul](https://developer.hashicorp.com/consul) is a [service mesh](https://www.hashicorp.com/resources/what-is-a-service-mesh) solution that launches with a key-value store.

Vault requires a storage backend like Consul to manage its configuration and secrets when it is run in high-availability.

```yaml
$ cat > helm-consul-values.yml << EOF
global:
  datacenter: vault-kubernetes-tutorial

client:
  enabled: true

server:
  replicas: 1
  bootstrapExpect: 1
  disruptionBudget:
    maxUnavailable: 0
EOF
```

```bash
$ helm repo add hashicorp https://helm.releases.hashicorp.com
$ helm repo update
$ helm install consul hashicorp/consul --values helm-consul-values.yml
```

```yaml
$ cat > helm-vault-concul-values.yml << EOF
server:
  affinity: ""
  ha:
    enabled: true
EOF
```

```bash
$ helm install vault hashicorp/vault --values helm-vault-concul-values.yml
```

Init & Unseal

```bash
$ kubectl exec vault-0 -- vault operator init \
	-key-shares=1 \
	-key-threshold=1 \
	-format=json > cluster-keys.json
	
$ cat cluster-keys.json | jq -r ".unseal_keys_b64[]"
$ VAULT_UNSEAL_KEY=$(cat cluster-keys.json | jq -r ".unseal_keys_b64[]")
```

Secrets engine

```bash
$ cat cluster-keys.json | jq -r ".root_token"
$ kubectl exec -ti vault-0 -- /bin/sh

$ vault login
$ vault secrets enable -path=secret kv-v2
$ vault kv put secret/webapp/config username="static-user" password="static-password"
$ vault kv get secret/webapp/config
```

Kubernetes authentication

```bash
$ cat cluster-keys.json | jq -r ".root_token"
$ kubectl exec -ti vault-0 -- /bin/sh

$ vault login
$ vault auth enable kubernetes
$ vault write auth/kubernetes/config \
	kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"

# policy
$ vault policy write webapp - <<EOF
path "secret/data/webapp/config" {
  capabilities = ["read"]
}
EOF

# create role & connect to k8s service account & policy
$ vault write auth/kubernetes/role/webapp \
	bound_service_account_names=vault \
	bound_service_account_namespaces=default \
	policies=webapp \
	ttl=24h
```

Web Application

```yaml
$ cat > deployment01-webapp.yml << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  labels:
    app: webapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      serviceAccountName: vault
      containers:
        - name: app
          image: hashieducation/simple-vault-client:latest
          imagePullPolicy: Always
          env:
            - name: VAULT_ADDR
              value: 'http://vault:8200'
            - name: JWT_PATH
              value: '/var/run/secrets/kubernetes.io/serviceaccount/token'
            - name: SERV
EOF
```

```bash
$ kubectl port-forward \
	$(kubectl get pod -l app=webapp -o jsonpath="{.items[0].metadata.name}") \
    8080:8080

$ curl http://localhost:8080
```

Clean up

```bash
$ kubectl delete -f deployment01-webapp.yml
$ helm uninstall vault
$ helm uninstall consul
```

### [TLS](https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-minikube-tls)

Create certificate

```bash
$ mkdir /tmp/vault
$ export VAULT_K8S_NAMESPACE="sip"
$ export VAULT_HELM_RELEASE_NAME="vault"
$ export VAULT_SERVICE_NAME="vault-internal"
$ export K8S_CLUSTER_NAME="cluster.local"
$ export WORKDIR=/tmp/vault
```

```bash
# private key
$ openssl genrsa -out ${WORKDIR}/vault.key 2048

# csr
$ cat > ${WORKDIR}/vault-csr.conf <<EOF
[req]
default_bits = 2048
prompt = no
encrypt_key = yes
default_md = sha256
distinguished_name = kubelet_serving
req_extensions = v3_req
[ kubelet_serving ]
O = system:nodes
CN = system:node:*.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = *.${VAULT_SERVICE_NAME}
DNS.2 = *.${VAULT_SERVICE_NAME}.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}
DNS.3 = *.${VAULT_K8S_NAMESPACE}
IP.1 = 127.0.0.1
EOF

$ openssl req -new \
    -key ${WORKDIR}/vault.key \
    -out ${WORKDIR}/vault.csr \
    -config ${WORKDIR}/vault-csr.conf

# issue
$ cat > ${WORKDIR}/csr.yaml << EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
   name: vault.svc
spec:
   signerName: kubernetes.io/kubelet-serving
   expirationSeconds: 8640000
   request: $(cat ${WORKDIR}/vault.csr|base64|tr -d '\n')
   usages:
   - digital signature
   - key encipherment
   - server auth
EOF

# approve
$ kubectl certificate approve vault.svc
$ kubectl get csr
```

Store keypair in K8s

```bash
# cert
$ kubectl get csr vault.svc -o jsonpath='{.status.certificate}' | openssl base64 -d -A -out ${WORKDIR}/vault.crt

# ca cert
$ kubectl config view \
	--raw \
	--minify \
	--flatten \
	-o jsonpath='{.clusters[].cluster.certificate-authority-data}' | base64 -d > ${WORKDIR}/vault.ca

# ns
$ kubectl create namespace $VAULT_K8S_NAMESPACE

# secret
$ kubectl create secret generic vault-ha-tls \
	-n $VAULT_K8S_NAMESPACE \
	--from-file=vault.key=${WORKDIR}/vault.key \
	--from-file=vault.crt=${WORKDIR}/vault.crt \
	--from-file=vault.ca=${WORKDIR}/vault.ca
```

Deploy

```yaml
$ cat > ${WORKDIR}/helm-vault-raft-values-tls.yaml <<EOF
global:
   enabled: true
   tlsDisable: false
injector:
   enabled: true
server:
   extraEnvironmentVars:
      VAULT_CACERT: /vault/userconfig/vault-ha-tls/vault.ca
      VAULT_TLSCERT: /vault/userconfig/vault-ha-tls/vault.crt
      VAULT_TLSKEY: /vault/userconfig/vault-ha-tls/vault.key
   volumes:
      - name: userconfig-vault-ha-tls
        secret:
         defaultMode: 420
         secretName: vault-ha-tls
   volumeMounts:
      - mountPath: /vault/userconfig/vault-ha-tls
        name: userconfig-vault-ha-tls
        readOnly: true
   standalone:
      enabled: false
   affinity: ""
   ha:
      enabled: true
      replicas: 3
      raft:
         enabled: true
         setNodeId: true
         config: |
            cluster_name = "vault-integrated-storage"
            ui = true
            listener "tcp" {
               tls_disable = 0
               address = "[::]:8200"
               cluster_address = "[::]:8201"
               tls_cert_file = "/vault/userconfig/vault-ha-tls/vault.crt"
               tls_key_file  = "/vault/userconfig/vault-ha-tls/vault.key"
               tls_client_ca_file = "/vault/userconfig/vault-ha-tls/vault.ca"
            }
            storage "raft" {
               path = "/vault/data"
            }
            disable_mlock = true
            service_registration "kubernetes" {}
EOF
```

```bash
$ helm install $VAULT_HELM_RELEASE_NAME hashicorp/vault \
	 -n $VAULT_K8S_NAMESPACE \
	-f ${WORKDIR}/helm-vault-raft-values-tls.yaml
```

```bash
# init
$ kubectl exec -n $VAULT_K8S_NAMESPACE vault-0 -- vault operator init \
    -key-shares=1 \
    -key-threshold=1 \
    -format=json > ${WORKDIR}/cluster-keys.json

# unseal    
$ jq -r ".unseal_keys_b64[]" ${WORKDIR}/cluster-keys.json
$ VAULT_UNSEAL_KEY=$(jq -r ".unseal_keys_b64[]" ${WORKDIR}/cluster-keys.json)
$ kubectl exec -n $VAULT_K8S_NAMESPACE vault-0 -- vault operator unseal $VAULT_UNSEAL_KEY
```

```bash
# join & unseal for vault-1
$ kubectl exec -n $VAULT_K8S_NAMESPACE -it vault-1 -- /bin/sh

$ vault operator raft join \
	-address=https://vault-1.vault-internal:8200 \
	-leader-ca-cert="$(cat /vault/userconfig/vault-ha-tls/vault.ca)" \
	-leader-client-cert="$(cat /vault/userconfig/vault-ha-tls/vault.crt)" \
	-leader-client-key="$(cat /vault/userconfig/vault-ha-tls/vault.key)" \
	https://vault-0.vault-internal:8200

$ kubectl exec -n $VAULT_K8S_NAMESPACE -ti vault-1 -- vault operator unseal $VAULT_UNSEAL_KEY

# join & unseal for vault-2
$ kubectl exec -n $VAULT_K8S_NAMESPACE -it vault-2 -- /bin/sh

$ vault operator raft join \
	-address=https://vault-2.vault-internal:8200 \
	-leader-ca-cert="$(cat /vault/userconfig/vault-ha-tls/vault.ca)" \
	-leader-client-cert="$(cat /vault/userconfig/vault-ha-tls/vault.crt)" \
	-leader-client-key="$(cat /vault/userconfig/vault-ha-tls/vault.key)" \
	https://vault-0.vault-internal:8200

$ kubectl exec -n $VAULT_K8S_NAMESPACE -ti vault-2 -- vault operator unseal $VAULT_UNSEAL_KEY
```

```bash
# cluster root token
$ export CLUSTER_ROOT_TOKEN=$(cat ${WORKDIR}/cluster-keys.json | jq -r ".root_token")

# login
$ kubectl exec -n $VAULT_K8S_NAMESPACE vault-0 -- vault login $CLUSTER_ROOT_TOKEN

# list peer & status
$ kubectl exec -n $VAULT_K8S_NAMESPACE vault-0 -- vault operator raft list-peers
$ kubectl exec -n $VAULT_K8S_NAMESPACE vault-0 -- vault status
```

Create a secret

```bash
$ kubectl exec -n $VAULT_K8S_NAMESPACE -it vault-0 -- /bin/sh

$ vault login
$ vault secrets enable -path=secret kv-v2
$ vault kv put secret/tls/apitest username="apiuser" password="supersecret"
$ vault kv get secret/tls/apitest
```

Get via API

```bash
$ kubectl -n vault port-forward service/vault 8200:8200
$ curl --cacert $WORKDIR/vault.ca \
	--header "X-Vault-Token: $CLUSTER_ROOT_TOKEN" \
	https://127.0.0.1:8200/v1/secret/data/tls/apitest | jq .data.data
```

Clean up

```bash
$ helm uninstall vault
$ rm -r $WORKDIR
```

