Via PKI sercrets engine, services can get certificates **without** going through the usual **manual** process.

证书吊销列表（Certificate Revocation List, CRL）由 CA 签发，包含已撤销的证书序列号列表，由 CA 维护。

客户端和服务器都会定期更新并缓存 CRL 文件，通过检查证书序列号是否在 CRL 中判断对方证书是否吊销。

[API](https://developer.hashicorp.com/vault/api-docs/secret/pki)

### [Setup & Usage](https://developer.hashicorp.com/vault/docs/secrets/pki/setup)

 ```bash
# help
$ vault path-help pki
 ```

```bash
# enable
$ vault secrets enable pki
# set global ttl
$ vault secrets tune -max-lease-ttl=8760h pki
```

Vault can accept an **existing** key pair, or it can generate its own **self-signed** root.

**Recommend** maintaining your root CA outside of Vault and providing Vault a signed intermediate CA.

```bash
# conf CA keypair
$ vault write pki/root/generate/internal \
    common_name=my-website.com \
    ttl=8760h
```

Update the **CRL** location and **issuing** certificates.

```bash
$ vault write pki/config/urls \
    issuing_certificates="http://127.0.0.1:8200/v1/pki/ca" \
    crl_distribution_points="http://127.0.0.1:8200/v1/pki/crl"
```

Configure a **role (user/machine)** to gen a cert.

```bash
$ vault write pki/roles/example-dot-com \
    allowed_domains=my-website.com \
    allow_subdomains=true \
    max_ttl=72h
```

`/issue` endpoint to gen new cred.

```bash
$ vault write pki/issue/example-dot-com \
    common_name=www.my-website.com
```

### [Root CA setup](https://developer.hashicorp.com/vault/docs/secrets/pki/quick-start-root-ca)

```bash
# enable
$ vault secrets enable pki
# conf 10 yrs
$ vault secrets tune -max-lease-ttl=87600h pki
# gen
$ vault write pki/root/generate/internal \
	common_name=myvault.com \
	ttl=87600h
# url conf
$ vault write pki/config/urls \
	issuing_certificates="http://vault.example.com:8200/v1/pki/ca"
	crl_distribution_points="http://vault.example.com:8200/v1/pki/crl"
# role to gen cred
$ vault write pki/roles/example-dot-com \
    allowed_domains=example.com \
    allow_subdomains=true \
    max_ttl=72h
# gen cert via /issue endpoint with role name
$ vault write pki/issue/example-dot-com \
    common_name=blah.example.com
```

### [Intermediate setup](https://developer.hashicorp.com/vault/docs/secrets/pki/quick-start-intermediate-ca)

Create intermediate CA using the root CA to sign the intermediate's certificate.

```bash
# enable
$ vault secrets enable -path=pki_int pki
# conf 5 yrs
$ vault secrets tune -max-lease-ttl=43800h pki_int
# gen csr
$ vault write pki_int/intermediate/generate/internal \
	common_name="myvault.com Intermediate Authority" \
	ttl=43800h
# sign using CA
$ vault write pki/root/sign-intermediate \
	csr=@pki_int.csr \
	format=pem_bundle \
	ttl=43800h
# set the intermediate ca to root-signed cert
$ vault write pki_int/intermediate/set-signed \
	ertificate=@signed_certificate.pem
# set URL conf
$ vault write pki_int/config/urls \
	ssuing_certificates="http://127.0.0.1:8200/v1/pki_int/ca"
	crl_distribution_points="http://127.0.0.1:8200/v1/pki_int/crl"
# role to gen cert
$ vault write pki_int/roles/example-dot-com \
    allowed_domains=example.com \
    allow_subdomains=true max_ttl=72h
# issue cert
$ vault write pki_int/issue/example-dot-com \
    common_name=blah.example.com
```

