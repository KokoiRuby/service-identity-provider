## Service-Identity-Provider

### Overiview

**Service-Identity-Provider (SIP)** provides X.509 keypair to secure Kubernetes in-cluster mTLS. By leveraging [PKI](https://developer.hashicorp.com/vault/docs/secrets/pki) secrets engine of HashiCorp [Vault](https://www.vaultproject.io/), the keypair(s) can be automatically provisioned without going through usual manual process of generation.

An application can request SIP for keypair needed to estabilish mTLS between **Service Provider** & **Service Consumer** (client of **Service Provider**) by declaring Kubernetes [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRs) according to Kubernetes [Custom Resource Definition](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/) (CRD).

The signature algorithms used for generating keypair is ECDSA-256 and SHA-256 for hash algorithm.

#### Use Cases

- Request internal certificate for server & client authentication.
- Request internal CA certificate for client authentication.

#### Features

- HashiCorp Vault [PKI](https://developer.hashicorp.com/vault/docs/secrets/pki) secrets engine
- Kubernetes [Opeartor pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Kubebuilder](https://book.kubebuilder.io/)

#### Architecture

**Service Identity Provider** watches CRs created by **Service Provider** & **Service Consumer** & reconcile the secrets for mTLS.

**Serivce Provider** & **Service Consumer** (client of **Service Provider**) communicate with each other in mTLS given keypair in secrets.

![arch](README.assets/arch.png)

#### Certificate Relationships

Root CA certficate for **Service Provide**r is shared by all **Service Consumers**.

Intermediate CA Certificate for **Service Provider** is shared by all **Service Providers**.

![cert_rel](README.assets/cert_rel.png)

### Deployment

#### Prerequisite

An out-of-box Kubernetes cluster environment, try [kind](https://kind.sigs.k8s.io/). 👈 Note: 3 [workers](https://kind.sigs.k8s.io/docs/user/configuration/#nodes) are required for Vault HA.

Availability of [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) CLI & [helm](https://helm.sh/), the package manager to interact with Kubernetes cluster.

#### Helm

```bash
$ export CERTM_NAMESPACE="cert-manager"
$ export SIP_NAMESPACE="sip"
$ export SIP_RELEASE="sip"
```

Note: because we've enabled the [webhooks](https://book.kubebuilder.io/reference/webhook-overview) in Operators, [cert-manager](https://cert-manager.io/) is required for provisioning certificates for the webhook server. **Please follow the cert-manager [documentation](https://cert-manager.io/docs/installation/helm/) to install it first.**

```bash
$ helm repo add jetstack https://charts.jetstack.io --force-update
```

```bash
$ helm install \
	cert-manager jetstack/cert-manager \
	--namespace $CERTM_NAMESPACE \
	--create-namespace \
	--version v1.15.3 \
	--set crds.enabled=true
```

```bash
$ helm install \
	$SIP_RELEASE Service-Identity-Provider-0.1.6.tgz \
	--namespace $SIP_NAMESPACE \
	--create-namespace
```

##### Demo Application

```bash
$ kubectl -n $SIP_NAMESPACE apply -f helm/manifests/demo/service-provider
$ kubectl -n $SIP_NAMESPACE apply -f helm/manifests/demo/service-consumer
# verify
$ kubectl -n $SIP_NAMESPACE get po | grep consumer
$ kubectl -n $SIP_NAMESPACE logs -f <podName>
```

##### Clean up

```bash
$ helm uninstall sip -n $SIP_NAMESPACE
$ kubectl delete ns $SIP_NAMESPACE
```

#### Manual

##### Vault

> HashiCorp [Vault](https://www.vaultproject.io/) is an identity-based secrets and encryption management system.
>
> A secret is anything that you want to tightly control access to. In our case, a X.509 certificate or key.
>
> Note: [HA](https://developer.hashicorp.com/vault/docs/concepts/ha) is enabled against outage.

The Vault cluster is deployed as follows:

1. Create Namespace.

```bash
$ export SIP_NAMESPACE="sip"
$ export VAULT_RELEASE="vault"
$ kubectl create ns $NAMESPACE
```

2. Bootstrap a keypair secret to enable Vault TLS mode. 

Note: the certificate is signed by Kubernetes cluster Root CA, this makes sure that any pods hold `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` can validate Vault cluster.

```bash
$ kubectl apply -f helm/manifests/vault/bootstrap/ -n $SIP_NAMESPACE
```

```bash
# verify
$ kubectl get secret -n $SIP_NAMESPACE | grep vault-ha-tls
```

3. Deploy Vault.

Note: the `helm-vault-raft-values-tls.yaml` is extracted from vault-**0.28.1**.tgz.

```bash
$ helm repo add hashicorp https://helm.releases.hashicorp.com
```

```bash
$ helm repo update
```

```bash
$ helm install $VAULT_RELEASE hashicorp/vault \
	--version 0.28.1 \
	--values helm/manifests/vault/deploy/helm-vault-raft-values-tls.yaml \
	--namespace $SIP_NAMESPACE
```

Note: the `vault-*` are not ready yet, it's okay.

```bash
# verify
$ kubectl get po -n $SIP_NAMESPACE | grep vault
```

4. Inititial and Unseal Vault cluster.

Note: If you happen to see any vault-* restarted before, you have to re-run it to [unseal](https://developer.hashicorp.com/vault/docs/concepts/seal) Vault cluster once again to make service ready. Of course you have to delete previous `Succeeded` job to re-run.

```bash
$ kubectl apply -f helm/manifests/vault/init/ -n $SIP_NAMESPACE
```

You should be able to see all `vault-*` are ready now.

```bash
# verify
$ kubectl get po -n $SIP_NAMESPACE | grep vault
```

5. Enable PKI for Server AuthN.

Note: the root-token will be kept somewhere safe in the future release.

```bash
$ kubectl apply -f helm/manifests/vault/pki/ -n $SIP_NAMESPACE
```

You should be able to see [pki secrets engine](https://developer.hashicorp.com/vault/docs/secrets/pki) `sip-root-ca/` & `sip-interm-ca/` are enabled.

```bash
# verify
$ kubectl -n $SIP_NAMESPACE get secret vault-root-token -o json | jq -r .data.token | base64 -d -
$ kubectl -n $SIP_NAMESPACE exec -it vault-0 -- sh
$ vault login
$ vault secrets list
```

##### SIP

Note: because we've enabled the [webhooks](https://book.kubebuilder.io/reference/webhook-overview) in Operators, [cert-manager](https://cert-manager.io/) is required for provisioning certificates for the webhook server. **Please follow [the cert-manager documentation](https://cert-manager.io/docs/installation/) to install it first.**

```bash
$ kubectl apply -f helm/manifests/vault/sip/ -n $SIP_NAMESPACE
```

```bash
# verify
$ kubectl -n $SIP_NAMESPACE get po | grep controller-manager
```

##### Demo Application

```bash
$ kubectl -n $SIP_NAMESPACE apply -f helm/manifests/demo/service-provider
$ kubectl -n $SIP_NAMESPACE apply -f helm/manifests/demo/service-consumer
```

```bash
# verify
$ kubectl -n $SIP_NAMESPACE get po | grep consumer
$ kubectl -n $SIP_NAMESPACE logs -f <podName>
```

##### Clean up

```bash
$ kubectl delete ns $SIP_NAMESPACE
```

### API

The API is used to enable X.509 certificate-based authentication & mTLS between **Service Provider** & **Service Consumer**.

The API that scaffoled by [Kubebuilder](https://book.kubebuilder.io/) sticks to [OpenAPI V3](https://swagger.io/specification/) & [K8s API Conventions](https://kubernetes.io/docs/reference/using-api/api-concepts/).

#### sip.sec.com/InternalCertificate

The **sip.sec.com/InternalCertificate** is used for generating PEM encoded keypair stored in Kubernetes Secrets for **Service Provider** & **Service Consumer**.

##### v1alpha

The following parameters/paths are supported (under CRD YAML `spec` section)

| Parameter/Path                          |  Type  | Is Mandatory | Description                                                  |
| :-------------------------------------- | :----: | :----------: | ------------------------------------------------------------ |
| certificate.extendedKeyUsage.ClientAuth |  bool  |     True     | It indicates the [extended key usage](https://golang.org/pkg/crypto/x509/#ExtKeyUsage) constraint on the issued certificate. Note: ClientAuth & ServerAuth cannot be both true or false. (Achieved by Validating Webhook) |
| certificate.extendedKeyUsage.ServerAuth |  bool  |     True     | It indicates the [extended key usage](https://golang.org/pkg/crypto/x509/#ExtKeyUsage) constraint on the issued certificate. Note: ClientAuth & ServerAuth cannot be both true or false. (Achieved by Validating Webhook) |
| certificate.issuer                      | string |    False     | It indicates the issuer of certificate. Note: mandatory if ClientAuth is true. (Achieved by Validating Webhook) |
| certificate.subject.cn                  | string |     True     | It indicates Common Name of certificate.                     |
| secret.certName                         | string |    False     | It indicates the name of certificate. Note: `tls.cert` is used if not given. (Achieved by Defaulting Webhook) |
| secret.keyName                          | string |    False     | It indicates the name of key. Note: `tls.key` is used if not given. (Achieved by Defaulting Webhook) |
| secret.name                             | string |     True     | It indicates the name of secret to create in Kubernetes cluster. |

#### sip.sec.com/InternalClientCA

The **sip.sec.com/InternalClientCA** is used for enabling Vault PKI secrets engine CA to issue client keypair for **Service Consumer**.

##### v1alpha

The following parameters/paths are supported (under CRD YAML `spec` section)

| Parameter/Path         |  Type  | Is Mandatory | Description                                                  |
| :--------------------- | :----: | :----------: | ------------------------------------------------------------ |
| certificate.subject.cn | string |     True     | It indicates Common Name of certificate.                     |
| secret.certName        | string |    False     | It indicates the name of certificate. Note: `client-ca.pem` is used if not given. (Achieved by Defaulting Webhook) |
| secret.name            | string |     True     | It indicates the name of secret to create in Kubernetes cluster. |

### Implementation

The core logics are implemented in the **Reconcile** function inside Operator (controller of Custom Resource) scaffolded by [Kubebuilder](https://book.kubebuilder.io/).

See more in [UML](https://github.com/KokoiRuby/service-identity-provider/tree/main/uml). 👈

### Limitation

- The validity of InternalCertificate does not support renewal for the moment. Default is 1 week. You have to delete & re-create custom resources to renew.

### Operation and Maintenance

TODO...

### Troubleshooting

TODO...

### Q&A

> [Q] Why not using a Job rather than initContainers to initialize & unseal Vault?

The Vault APIs are not availables until pod status becomes `Running`. The initContainers run ahead of vault main container and they will never reach the APIs. The pod ends up in `Initializing` forever. **If you have any better ideas, please drop a pull request.**

> [Q] Could we just use one controller-manager to reconcile multiple custom resources?

As mentioned in [Good Practices](https://book.kubebuilder.io/reference/good-practices#why-should-one-avoid-a-system-design-where-a-single-controller-is-responsible-for-managing-multiple-crds-custom-resource-definitionsfor-example-an-install_all_controllergo) of Kubebuilder, it is against the design purpose of [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime).

### Reference

[Vault PKI secrets engine (API)](https://developer.hashicorp.com/vault/api-docs/secret/pki)

[Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources)

[Kubebuilder](https://book.kubebuilder.io/)

### TODO

- Helm package
- K8s AuthN
- Thinking about moving root token to somewhere safe...or simply just revoke it...
- Certificate Renewal
- Decouple config from src
- ++ more customizable fields for CR