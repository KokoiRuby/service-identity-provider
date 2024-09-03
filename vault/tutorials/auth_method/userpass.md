AuthN against **username (lowercase) & password**.

[API](https://developer.hashicorp.com/vault/api-docs/auth/userpass).

```bash
$ vault login -method=userpass \
    username=name \
    password=pwd
    
$ curl \
    --request POST \
    --data '{"password": "pwd"}' \
    http://127.0.0.1:8200/v1/auth/userpass/login/{name}
```

```bash
# enable
# Default /auth/userpass.
$ vault auth enable userpass
# or another path
$ vault auth enable -path=<path> userpass
# ++ user & permission
$ vault write auth/<userpass:path>/users/{name} \
    password=pwd \
    policies=<policy>
```

Lockout if bad cred in quick succession.

Enabled by default.

- "lockout threshold" is 5 attempts
- "lockout duration" is 15 minutes
- "lockout counter reset" is 15 minutes.

Disable globally by ENV `VAULT_DISABLE_USER_LOCKOUT`.

Disable specifically by `disable_lockout` in [conf](https://developer.hashicorp.com/vault/docs/configuration/user-lockout#user_lockout-stanza) or [auth tune](https://developer.hashicorp.com/vault/docs/commands/auth/tune) ([api](https://developer.hashicorp.com/vault/api-docs/system/auth#tune-auth-method)).