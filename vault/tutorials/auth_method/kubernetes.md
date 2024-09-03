AuthN against K8s Service Account Token.

Default at `/auth/kubernetes`.

[API](https://developer.hashicorp.com/vault/api-docs/auth/kubernetes).

```bash
$ vault write auth/kubernetes/login role=demo jwt=...
# api
￥ curl \
    --request POST \
    --data '{"jwt": "<your service account jwt>", "role": "demo"}' \
    http://127.0.0.1:8200/v1/auth/kubernetes/login
```

Use `kubectl cluster-info` to validate the Kubernetes host address and TCP port.

```bash
# enable
$ vault auth enable kubernetes
# conf
$ vault write auth/kubernetes/config \
    token_reviewer_jwt="<your reviewer service account JWT>" \
    kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
    kubernetes_ca_cert=@ca.crt
# create role
$ vault write auth/kubernetes/role/demo \
    bound_service_account_names=myapp \
    bound_service_account_namespaces=default \
    policies=default \
    ttl=1h
```

`token_reviewer_jwt` 是 HashiCorp Vault 中用于配置 Kubernetes 身份验证方法的一部分。

它在 Vault 与 Kubernetes 集群之间建立身份验证时扮演重要角色。

当 Vault 接收到一个 Kubernetes Pod 提交的请求（比如登录请求）时，它需要验证这个请求是否来自于一个合法的 Pod。验证是通过 Kubernetes 的 `TokenReview` API 完成的。

需要创建或使用一个具有 `TokenReview` 权限的服务账户，然后从中提取 SA 的 JWT。

`system:auth-delegator` 提供 TokenReview API 权限。

```bash
$ kubectl create serviceaccount vault-reviewer
$ kubectl create clusterrolebinding vault-reviewer-binding \
    --clusterrole=system:auth-delegator \
    --serviceaccount=default:vault-reviewer

$ TOKEN_REVIEWER_JWT=$(kubectl get secret $(kubectl get serviceaccount vault-reviewer -o jsonpath='{.secrets[0].name}') -o jsonpath='{.data.token}' | base64 --decode)

# inside container
$ cat /var/run/secrets/kubernetes.io/serviceaccount/token
```

### How to work with short-lived K8s tokens

| Option                               | All tokens are short-lived | Can revoke tokens early | Other considerations                                         |
| :----------------------------------- | :------------------------- | :---------------------- | :----------------------------------------------------------- |
| Use local token as reviewer JWT      | Yes                        | Yes                     | Requires Vault (1.9.3+) to be deployed on the Kubernetes cluster |
| Use client JWT as reviewer JWT       | Yes                        | Yes                     | Operational overhead                                         |
| Use long-lived token as reviewer JWT | No                         | Yes                     |                                                              |
| Use JWT auth instead                 | Yes                        | No                      |                                                              |

1. **Use local token as reviewer JWT (Recommended)**

- Running Vault in K8s pod
- Omit `token_reviewer_jwt` & `kubernetes_ca_cert`
- Vault will attempt to load them from `token` and `ca.crt` respectively

```bash
$ vault write auth/kubernetes/config \
    kubernetes_host=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT
```

2. **Use client JWT as reviewer JWT**

- Omit the `token_reviewer_jwt`
- `disable_local_ca_jwt=true` if Vault is running in K8s
- Each client of Vault would need the `system:auth-delegator` ClusterRole

```bash
$ kubectl create clusterrolebinding vault-client-auth-delegator \
    --clusterrole=system:auth-delegator \
    --group=group1 \
    --serviceaccount=default:svcaccount1 \
    ...
```

3. **Use long-lived token as reviewer JWT**

- Create service account with `system:auth-delegator` ClusterRole.

```yaml
$ kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: vault-k8s-auth-secret
  annotations:
    kubernetes.io/service-account.name: vault
type: kubernetes.io/service-account-token
EOF
```

4. **Use JWT AuthN**

- K8s as OIDC provider
- Client tokens cannot be revoked before their TTL expires, remember to keep TTL short.

### Discovering the service account `issuer`

v1.21+, issuer of service account may require setting to kube-apiserver `--service-account-issuer`

```bash
#  chk iss field
$ echo '{"apiVersion": "authentication.k8s.io/v1", "kind": "TokenRequest"}' \
  | kubectl create -f- --raw /api/v1/namespaces/default/serviceaccounts/default/token \
  | jq -r '.status.token' \
  | cut -d . -f2 \
  | base64 -d
# or
$ kubectl get --raw /.well-known/openid-configuration | jq -r .issuer
# conf
$ vault write auth/kubernetes/config \
    kubernetes_host="https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT" \
    issuer="\"test-aks-cluster-dns-d6cbb78e.hcp.uksouth.azmk8s.io\""
```

### [Code example](https://developer.hashicorp.com/vault/docs/auth/kubernetes#code-example)
