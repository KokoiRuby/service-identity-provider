## [vault-client-go](https://github.com/hashicorp/vault-client-go)



### Internal Certificate

> TODO: initContainer

### Internal Client CA

> TODO: Managed by Operator

## CLI

### Internal Certificate

> TODO: initContainer

1. Enable PKI secrets engine.

```bash
$ vault secrets enable \
	-path=sip-root-ca \
	-description="CA certificate backend created by sip for server authn" \
	pki

$ vault secrets enable \
	-path=sip-interm-ca \
	-description="CA certificate backend created by sip for server authn" \
	pki

$ vault secrets list
```

2. Set TTL.

```bash
$ vault secrets tune \
	-max-lease-ttl=876000h \
	sip-root-ca

$ vault secrets tune \
	-max-lease-ttl=168h \
	sip-interm-ca
	
$ vault read -format=json sys/mounts/sip-root-ca | jq .data.config.max_lease_ttl
$ vault read -format=json sys/mounts/sip-interm-ca | jq .data.config.max_lease_ttl
```

3. Configure CA keypair.

```bash
$ vault write sip-root-ca/root/generate/internal \
    common_name="sip Internal Root CA" \
    key_type=ec \
    key_bits=256 \
    ttl=876000h
    
$ vault list -format=json sip-root-ca/keys 
```

4. Update CRL location & issuing certificates, can be updated in the future.

```bash
$ vault write sip-root-ca/config/urls \
    issuing_certificates="http://127.0.0.1:8200/v1/sip-root-ca/ca" \
    crl_distribution_points="http://127.0.0.1:8200/v1/sip-root-ca/crl"
    
$ vault write sip-interm-ca/config/urls \
     issuing_certificates="http://127.0.0.1:8200/v1/sip-interm-ca/ca" \
     crl_distribution_points="http://127.0.0.1:8200/v1/sip-interm-ca/crl"    
```

5. [Configure](https://developer.hashicorp.com/vault/api-docs/secret/pki) a role that maps a name in Vault to a procedure for generating a certificate.

```bash
$ vault write sip-root-ca/roles/root-ca \
	key_type=ec \
	key_bits=256 \
    key_usage="CertSign,CRLSign" \
    server_flag=false \
    client_flag=false \
    allowed_domains="service-provider,cluster.local" \
    allow_subdomains=true \
    enforce_hostnames=false \
    max_ttl=876000h
    
$ vault write sip-interm-ca/roles/interm-ca \
	key_type=ec \
	key_bits=256 \
    key_usage="DigitalSignature" \
    server_flag=false \
    client_flag=false \
    ext_key_usage="ServerAuth" \
    allowed_domains="service-provider,cluster.local" \
    allow_subdomains=true \
    enforce_hostnames=false \
    max_ttl=168h

$ vault read -format=json sip-root-ca/roles/root-ca
$ vault read -format=json sip-interm-ca/roles/interm-ca
```

6. Generate intermediate CA [CSR](https://developer.hashicorp.com/vault/api-docs/secret/pki#generate-intermediate-csr).

```bash
$ vault write sip-interm-ca/intermediate/generate/internal \
	common_name="sip Internal Intermediate CA" \
	key_type=ec \
	key_bits=256 \
	add_basic_constraints=true \
	ttl=168h \
	-format=json | jq -r .data.csr > sip_interm.csr
	
$ openssl req -text -noout -verify -in sip_interm.csr
```

7. Sign by Root CA.

```bash
$ vault write sip-root-ca/root/sign-intermediate \
	csr=@sip_interm.csr \
	format=pem \
	use_csr_values=true \
	ttl=876000h \
	-format=json | jq -r .data.certificate > sip_interm.pem

$ openssl x509 -in sip_interm.pem -text -noout
```

8. Set the intermediate CA signing certificate to the root-signed certificate.

```bash
$ vault write sip-interm-ca/intermediate/set-signed \
	certificate=@sip_interm.pem
```

9. Issue certificates

```bash
$ vault write sip-root-ca/issue/root-ca \
    common_name=service-provider.cluster.local \
    -format=json > /tmp/issue_output.json
    
$ jq -r '.data.certificate' /tmp/issue_output.json > server.pem
$ jq -r '.data.private_key' /tmp/issue_output.json > server-key.pem

$ vault write sip-interm-ca/issue/interm-ca \
    common_name=service-provider.cluster.local \
    -format=json > /tmp/issue_output.json
    
$ jq -r '.data.certificate' /tmp/issue_output.json > server.pem
$ jq -r '.data.private_key' /tmp/issue_output.json > server-key.pem

$ openssl x509 -in server.pem -text -noout
```

10. Clean up

```bash
$ vault secrets disable sip-root-ca
$ vault secrets disable sip-interm-ca
$ vault secrets list
```

### Internal Client CA

> TODO: Managed by Operator

1. Enable PKI secrets engine.

```bash
$ vault secrets enable \
	-path=sip-client-ca/service-provider-ca \
	-description="CA certificate backend created by sip for client authn" \
	pki
	
$ vault secrets list
```

2. Set TTL.

```bash
$ vault secrets tune \
	-max-lease-ttl=87600h \
	sip-client-ca/service-provider-ca
	
$ vault read sys/mounts/sip-client-ca/service-provider-ca | grep ttl	
```

3. Configure CA keypair.

```bash
$ vault write sip-client-ca/service-provider-ca/root/generate/internal \
    common_name="service-provider" \
    key_type=ec \
    key_bits=256 \
    ttl=87600h
    
$ vault list sip-client-ca/service-provider-ca/keys
```

4. Update CRL location & issuing certificates, can be updated in the future.

```bash
$ vault write sip-client-ca/service-provider-ca/config/urls \
    issuing_certificates="http://127.0.0.1:8200/v1/sip-client-ca/service-provider-ca/ca" \
    crl_distribution_points="http://127.0.0.1:8200/v1/sip-client-ca/service-provider-ca/crl"
```

5. [Configure](https://developer.hashicorp.com/vault/api-docs/secret/pki) a role that maps a name in Vault to a procedure for generating a certificate.

```bash
$ vault write sip-client-ca/service-provider-ca/roles/client-ca \
	key_type=ec \
	key_bits=256 \
    key_usage="DigitalSignature" \
    server_flag=false \
    client_flag=false \
    ext_key_usage="ClientAuth" \
    allowed_domains="service-provider,cluster.local" \
    allow_subdomains=true \
    enforce_hostnames=false \
    max_ttl=168h
    
$ vault read sip-client-ca/service-provider-ca/roles/client-ca
$ vault read sip-client-ca/service-provider-ca/roles/client-ca | grep _flag
```

6. Issue certificates

```bash
$ vault write sip-client-ca/service-provider-ca/issue/client-ca \
    common_name=service-provider.cluster.local \
    -format=json > /tmp/issue_output.json
    
$ jq -r '.data.certificate' /tmp/issue_output.json > client.pem
$ jq -r '.data.private_key' /tmp/issue_output.json > client-key.pem    
    
$ openssl x509 -in client.pem -text -noout
```

7. Clean up

```bash
$ vault secrets disable sip-client-ca/service-provider-ca
$ vault secrets list
```

### 