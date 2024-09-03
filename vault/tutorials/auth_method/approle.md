AuthN against defined **roles for machines/apps**.

AppRole is is oriented to **automated** workflows, recommend using batch tokens.

An "AppRole" represents **a set of Vault policies and login constraints** that must be met to receive a token with those policies.

[API](https://developer.hashicorp.com/vault/api-docs/auth/approle).

```bash
$ vault write auth/approle/login \
    role_id=<role_id> \
    secret_id=<secret_id>    
# api
$ curl \
    --request POST \
    --data '{"role_id":"<role_id>","secret_id":"<secret_id> "}' \
    http://127.0.0.1:8200/v1/auth/approle/login
```

Conf

**RoleID** 是用于选择 AppRole 的标识符，其他凭据将基于该 AppRole 进行验证。

**SecretID** 是默认情况下进行登录时所需的凭据 `secret_id`，并且应始终保密。

```bash
# enable
$ vault auth enable approle
# create role
$ vault write auth/approle/role/my-role \
    token_type=batch \
    secret_id_ttl=10m \
    token_ttl=20m \
    token_max_ttl=30m \
    secret_id_num_uses=40
```

```bash
# get role id & secret id
$ vault read auth/approle/role/my-role/role-id
$ vault write -f auth/approle/role/my-role/secret-id
```

```bash
# api
# enable
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data '{"type": "approle"}' \
    http://127.0.0.1:8200/v1/sys/auth/approle
# create role
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data '{"policies": "dev-policy,test-policy", "token_type": "batch"}' \
    http://127.0.0.1:8200/v1/auth/approle/role/my-role
# get role id & secret id
$ curl \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/v1/auth/approle/role/my-role/role-id
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
     http://127.0.0.1:8200/v1/auth/approle/role/my-role/secret-id
```

[Code Example](https://developer.hashicorp.com/vault/docs/auth/approle#code-example)