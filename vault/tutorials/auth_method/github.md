AuthN against [GitHub personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens).

Any valid GitHub access token with the `read:org` scope for any user belonging to the Vault-configured organization can be used for authentication.

任何属于 Vault 配置的组织的用户，只要拥有 `read:org` 范围的有效 GitHub 访问令牌，都可以用于身份验证。

[API](https://developer.hashicorp.com/vault/api-docs/auth/github).

```bash
$ vault login -method=github token="MY_TOKEN"
# api
$ curl \
    --request POST \
    --data '{"token": "MY_TOKEN"}' \
    http://127.0.0.1:8200/v1/auth/github/login
```

Conf

```bash
# enable
$ vault auth enable github
# conf
$ vault write auth/github/config organization=hashicorp
# map users/teams of org to policies
$ vault write auth/github/map/users/sethvargo value=sethvargo-policy
$ vault write auth/github/map/teams/dev value=dev-policy
```

