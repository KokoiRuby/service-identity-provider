AuthN against [token](https://developer.hashicorp.com/vault/docs/concepts/tokens), as well to create/revoke.

Built-in automatically avail at `/auth/token`.

It can be used to bypass any other auth method.

[API](https://developer.hashicorp.com/vault/api-docs/auth/token)

注：使用 curl 不需要进行登录，直接带上 header 调用其他 API 即可。

```bash
$ vault login token=<token>

# api header
curl --header "X-Vault-Token: <your-token>" \
     --request GET \
     https://127.0.0.1:8200/v1/secret/data/mysecret
curl --header "Authorization: Bearer <token>" \
     --request GET \
     https://127.0.0.1:8200/v1/secret/data/mysecret
```