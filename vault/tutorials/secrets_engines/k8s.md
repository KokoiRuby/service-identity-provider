## [K8s secrets engine](https://developer.hashicorp.com/vault/docs/secrets/kubernetes)

It generates K8s **service account tokens**, and optionally **service accounts, role bindings, and roles**.

Token with configurable **TTL**, automatically deleted when lease expires.

[API](https://developer.hashicorp.com/vault/api-docs/secret/kubernetes)

### Setup

By default, Vault will connect to K8s using its **own service account**.

It's necessary to ensure that the service account Vault uses will have **permissions** to manage what it needs.

**ClusterRole**

```yaml
# min. to create sa token
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-minimal-secrets-abilities
rules:
- apiGroups: [""]
  resources: ["serviceaccounts/token"]
  verbs: ["create"]

# full to manage toke, sa, bindings & roles
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-full-secrets-abilities
rules:
# ++ label selection to configure ns on which a role can act
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get"]
# sa & sa token
- apiGroups: [""]
  resources: ["serviceaccounts", "serviceaccounts/token"]
  verbs: ["create", "update", "delete"]
# bindings
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["rolebindings", "clusterrolebindings"]
  verbs: ["create", "update", "delete"]
# roles
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["roles", "clusterroles"]
  verbs: ["bind", "escalate", "create", "update", "delete"]
 
```

**ClusterRoleBinding**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: vault-token-creator-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-minimal-secrets-abilities # or k8s-full-secrets-abilities
# to vault sa
subjects:
- kind: ServiceAccount
  name: vault
  namespace: vault
```

**注**：若 Vault 不会自动管理 roles & service accounts，需要给 Vault 设置一个颁布 token 的 service account。

**不建议**使用 vault 自带的 service account 用于颁布 token。

### Example

```bash
$ kubectl create namespace test
```

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-service-account-with-generated-token
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: test-role-list-pods
  namespace: test
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: test-role-abilities
  namespace: test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: test-role-list-pods
subjects:
- kind: ServiceAccount
  name: test-service-account-with-generated-token
  namespace: test
```

```bash
# enable
$ vault secrets enable kubernetes
# conf, see more in API
$ vault write -f kubernetes/config
# create a role to genearte token for sa
$ vault write kubernetes/roles/my-role \
	allowed_kubernetes_namespaces="*" \
    service_account_name="test-service-account-with-generated-token" \
    token_default_ttl="10m"
```

### Generating credentials

一旦 sa 通过了认证，一个可写的 `creds` endpoint 会暴露在对应 role 下，可返回一个新的 service account token。

可以使用该 token 发起任何权限允许内的 K8s API 请求。TTL 到期，token 会被撤销而失效。

```bash
$ vault write kubernetes/creds/my-role \
    kubernetes_namespace=test
    
$ curl -sk $(kubectl config view --minify -o 'jsonpath={.clusters[].cluster.server}')/api/v1/namespaces/test/pods \
    --header "Authorization: Bearer eyJHbGci0iJSUzI1Ni..."
```

### TTL

创建 Vault role 时可指定 TTL `token_default_ttl` & `token_max_ttl`。

```bash
$ vault write kubernetes/roles/my-role \
    allowed_kubernetes_namespaces="*" \
    service_account_name="new-service-account-with-generated-token" \
    token_default_ttl="10m" \
    token_max_ttl="2h"
```

也可以在访问 `creds` endpoint 生成 token 时指定。

```bash
$ vault write kubernetes/creds/my-role \
    kubernetes_namespace=test \
    ttl=20m
```

Verify TTL

```bash
$ echo 'eyJhbGc...' | cut -d'.' -f2 | base64 -d  | jq -r '.iat,.exp|todate'
```

### Audiences

Service Account Audience 是指一个用于指定服务账户所代表的实体的标识符。

这个标识符通常是一个 URL，用于指示服务账户生成的令牌 token 应该用于哪个服务或资源。

在创建 role 时 `token_default_audiences` 指定

```bash
$ vault write kubernetes/roles/my-role \
    allowed_kubernetes_namespaces="*" \
    service_account_name="new-service-account-with-generated-token" \
    token_default_audiences="custom-audience"
```

也可以在访问 `creds` endpoint 生成 token 时指定。

```bash
$ vault write kubernetes/creds/my-role \
    kubernetes_namespace=test \
    audiences="another-custom-audience"
```

Verify audience

```bash
$ echo 'eyJhbGc...' | cut -d'.' -f2 | base64 -d
```

### Automatically managing roles & service accounts

使用 K8s 已有的 role。

```bash
$ vault write kubernetes/roles/auto-managed-sa-role \
    allowed_kubernetes_namespaces="test" \
    kubernetes_role_name="test-role-list-pods"
    
# equivalent to
kubectl -n test create rolebinding \
	--role test-role-list-pods \
	--serviceaccount=vault:vault \
	vault-test-role-abilitie
	
# get cred
$ vault write kubernetes/creds/auto-managed-sa-role \
    kubernetes_namespace=test
```

自动创建 role 以及 sa 并进行 rolebinding。

```bash
$ vault write kubernetes/roles/auto-managed-sa-and-role \
    allowed_kubernetes_namespaces="test" \
    generated_role_rules='{"rules":[{"apiGroups":[""],"resources":["pods"],"verbs":["list"]}]}'

# get cred
$ vault write kubernetes/creds/auto-managed-sa-role \
    kubernetes_namespace=test
```



