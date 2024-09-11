## Bootstrap

```bash
$ kubectl get csr vault.svc -o jsonpath='{.status.certificate}' | \
	openssl base64 -d -A -out ./vault.crt
$ openssl x509 -in vault.crt -text -noout	
```



## Init

## PKI