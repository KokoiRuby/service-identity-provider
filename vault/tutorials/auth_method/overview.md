### [AuthN](https://developer.hashicorp.com/vault/docs/concepts/auth)

**认证 = “Who are you”**

Before a client can interact with Vault, it must *authenticate* against an [auth method](https://developer.hashicorp.com/vault/docs/auth).

Upon AuthN, a **token** is generated. This token is conceptually similar to a **session ID** on a website.

The token **may** have attached **policy**, which is mapped at authentication time.

```bash
# help
$ vault path-help auth/my-auth
# to enable
$ vault auth enable <type> -path=my-auth-path
```

AuthN via CLI/API

```bash
$ vault login -method=<method> token=<token>
# API help
$ vault path-help auth/<type>/login
```

Re-AuthN is required after the given [lease](https://developer.hashicorp.com/vault/docs/concepts/lease) period to continue accessing Vault.

To renew token.

```bash
$ vault token renew <token>
```

[Code example](https://developer.hashicorp.com/vault/docs/concepts/auth#code-example).

