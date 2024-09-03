### [CRD Generation](https://book.kubebuilder.io/reference/markers/crd)

```bash
$ mkdir internal-client-ca internal-cert

# init & create api
$ cd internal-client-ca
$ kubebuilder init \
	--domain sec.com \
	--repo github.com/KokoiRuby/service-identity-provider/operator/internal-client-ca
$ kubebuilder create api \
	--group sip \
	--version v1alpha1 \
	--kind InternalClientCA

# init & create api
$ cd internal-cert
$ kubebuilder init \
	--domain sec.com \
	--repo github.com/KokoiRuby/service-identity-provider/operator/internal-cert
$ kubebuilder create api \
	--group sip \
	--version v1alpha1 \
	--kind InternalCertificate
```

```bash
# install crd & sample cr
$ make manifests
$ kubectl apply -f ./config/samples
```

### Controller

