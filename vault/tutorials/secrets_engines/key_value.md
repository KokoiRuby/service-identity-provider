## [KV secrets engine](https://developer.hashicorp.com/vault/docs/secrets/kv)

A generic Key-Value store used to **store arbitrary secrets** within the configured **physical storage** for Vault.

### [KV version 1](https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v1)

**Non-versioned**, stores the **most recently** written value.

Any update will **overwrite** the original and **not recover-able**.

[API](https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v1)

:smile: Better runtime perf, storage size ↓ since no additional meta or history.

```bash
# enable
$ vault secrets enable -path=kv1 -version=1 kv
# diable
$ vault secrets disable kv1
```

Usage

```bash
# write
$ vault kv put kv/my-secret my-key=my-value
# read
$ vault kv get kv/my-secret
# delete
$ vault kv delete kv/my-secret
```

Use [Vault's password policy](https://developer.hashicorp.com/vault/docs/concepts/password-policies) to generate arbitrary values for key.

```bash
$ vault write sys/policies/password/example policy=-<<EOF

  length=20

  rule "charset" {
    charset = "abcdefghij0123456789"
    min-chars = 1
  }

  rule "charset" {
    charset = "!@#$%^&*STUVWXYZ"
    min-chars = 1
  }

EOF

# write
$ vault kv put kv/my-generated-secret \
    password=$(vault read -field password sys/policies/password/example/generate)
# read    
$ vault kv get kv/my-generated-secret

# delete
$ vault kv delete kv/my-generated-secret
$ vault delete sys/policies/password/example
```

TTL for lease duration as a **hint** for how often consumers should check back for a new value.

```bash
# just advisory kv
$ vault kv put kv/my-secret ttl=30m my-key=my-value
$ vault kv get kv/my-secret
$ vault kv delete kv/my-secret
```

### [KV version 2](https://developer.hashicorp.com/vault/docs/secrets/kv)

**Versioned**, default up to 10 versions retainable.

**Check-and-set** op to prevent unintentional overwrite. 本质是一个乐观锁，只有版本匹配才能修改，否则冲突拒绝执行。

Marked if a version is deleted，the underlying data is not removed, to remove a version's data by `vault kv destory`.

Delete all ver. & meta by `vault kv metadata delete`.

### vs.

API endpoint are different. Assuming the KV secrets engine is enabled at `secret/`.

[API](https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v2)

| Command           | KV v1 endpoint    | KV v2 endpoint                 |
| :---------------- | :---------------- | :----------------------------- |
| `vault kv get`    | secret/<key_path> | secret/**data**/<key_path>     |
| `vault kv put`    | secret/<key_path> | secret/**data**/<key_path>     |
| `vault kv list`   | secret/<key_path> | secret/**metadata**/<key_path> |
| `vault kv delete` | secret/<key_path> | secret/**data**/<key_path>     |

KV v2 only

| Command             | KV v2 endpoint                 |
| :------------------ | :----------------------------- |
| `vault kv patch`    | secret/**data**/<key_path>     |
| `vault kv rollback` | secret/**data**/<key_path>     |
| `vault kv undelete` | secret/**undelete**/<key_path> |
| `vault kv destroy`  | secret/**destroy**/<key_path>  |
| `vault kv metadata` | secret/**metadata**/<key_path> |

```bash
# enable
$ vault secrets enable -path=kv2 -version=2 kv
# diable
$ vault secrets disable kv2
```

Upgrade from version 1

:warning: [Change ACL rules](https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2#acl-rules) before upgrading.

```bash
# cli
$ vault kv enable-versioning kv1/

# api
$ cat > payload.json << EOF
{
  "options": {
      "version": "2"
  }
}
EOF

$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data @payload.json \
    http://127.0.0.1:8200/v1/sys/mounts/secret/tune
```

Usage, usage `-mount=secret` to avoid mistakes.

`-cas=1` is set to expect version to perform a check-and-set operation, otherwise rejected.

`-method=patch|rw` to specify HTTP `PATCH` or read-then write.

```bash
# write
$ vault kv put -mount=kv2 my-secret my-key=my-value
# read
$ vault kv get -mount=kv2 my-secret
# write another ver
$ vault kv put -mount=kv2 -cas=1 my-secret my-key=my-value my-key1=my-value1
# patch
$ vault kv patch -mount=kv2 -cas=2 my-secret my-key2=my-value2
# patch method
$ vault kv patch -mount=kv2 -method=patch -cas=3 my-secret my-key3=my-value3
$ vault kv patch -mount=kv2 -method=rw -cas=4 my-secret my-key4=my-value4
# previous version
$ vault kv get -mount=kv2 -version=1 my-secret
```

Use [Vault's password policy](https://developer.hashicorp.com/vault/docs/concepts/password-policies) to generate arbitrary values for key.

```bash
$ vault write sys/policies/password/example policy=-<<EOF

  length=20

  rule "charset" {
    charset = "abcdefghij0123456789"
    min-chars = 1
  }

  rule "charset" {
    charset = "!@#$%^&*STUVWXYZ"
    min-chars = 1
  }

EOF

# write
$ vault kv put -mount=kv2 my-generated-secret \
    password=$(vault read -field password sys/policies/password/example/generate)
# read    
$ vault kv get kv2/my-generated-secret

# delete
$ vault kv delete kv2/my-generated-secret
$ vault delete sys/policies/password/example
```

(Soft) Delete & (Hard) Destroy

```bash
# delete latest
$ vault kv delete -mount=kv2 my-secret
# delete prior to
$ vault kv delete -mount=kv2 -versions=2 my-secret
# un-delete
$ vault kv undelete -mount=kv2 -versions=2 my-secret
# destory version
$ vault kv destroy -mount=kv2 -version=5 my-secret
```

Meta

```bash
# tracking
$ vault kv metadata get -mount=kv2 my-secret
# delete version after & keey max versions to 2, max version will be applied to the next write
$ vault kv metadata put -mount=kv2 -max-versions 2 -delete-version-after="1h" my-secret
# add custom meta
$ vault kv metadata put -mount=kv2 \
	-custom-metadata=metaKey1=metaValue1 \
	-custom-metadata=metaKey2=metaValue2 \
	my-secret
# patch custom meta
$ vault kv metadata patch -mount=kv2 -custom-metadata=metaKey3=metaValue3 my-secret
# delete all meta version for a key
$ $ vault kv metadata delete -mount=kv2 my-secret
```

## [Versioned Key/value secrets engine](https://developer.hashicorp.com/vault/tutorials/secrets-management/versioned-kv)

:cry: v1 does not provide rollback, risk of unintentional data loss or overwrite.

:smile: v2 can retain a configurable number of secret versions (default is 10). Check-and-Set as optimistic lock.

Policy requirements: instead of root, use token with below set of [policies](https://developer.hashicorp.com/vault/tutorials/policies/policies).

```json
# Write and manage secrets in key-value secrets engine
path "secret*" {
  capabilities = [ "create", "read", "update", "delete", "list", "patch" ]
}

# To enable secrets engines
path "sys/mounts/*" {
  capabilities = [ "create", "read", "update", "delete" ]
}
```

Set up Lab

```bash
$ vault server -dev -dev-root-token-id root
```

Check KV secrets engine version

```bash
$ vault secrets list
$ vault secrets list -detailed
```

Write secrets

```bash
$ vault kv put -mount=secret /customer/acme \
	customer_name="ACME Inc." \
	contact_email="john.smith@acme.com"
$ vault kv put -mount=secret /customer/acme \
	customer_name="ACME Inc." \
    contact_email="john.smith@acme.com"
    
# chk
$ vault kv get -mount=secret customer/acme

# patch
$ vault kv patch -mount=secret /customer/acme contact_email="admin@acme.com"

# add custom meta
$ vault kv metadata put -custom-metadata=Membership="Platinum" -mount=secret /customer/acme
```

Retrieve a specific version of secret

```bash
# get version 1
$ vault kv get -version=1 -mount=secret /customer/acme
# get meta
$ vault kv metadata get secret/customer/acme
```

Specify the number of versions to keep

```bash
# limit max 4 version at path secret/
$ vault write secret/config max_versions=4
# chk
$ vault read secret/config
# limit max 4 version at path secret/customer/acme
$ vault kv metadata put -max-versions=4 -mount=secret /customer/acme
# create 4 more
$ vault kv put secret/customer/acme \
	name="ACME Inc." \
	contact_email="admin@acme.com"
# chk meta
$ vault kv metadata get -mount=secret /customer/acme
# failed to get previous
$ vault kv get -version=1 -mount=secret /customer/acme
```

Delete versions of secret

```bash
# delete 4 & 5
$ vault kv delete -versions="4,5" -mount=secret /customer/acme
# chk, deletion_time are set
$ vault kv metadata get -mount=secret /customer/acme
# undelete 5
$ vault kv undelete -versions=5 -mount= secret/customer/acme
```

Permanently delete data

```bash
# destroy 4
$ vault kv destroy -versions=4 -mount=secret /customer/acme
# deelte all
$ vault kv metadata delete -mount=secret /customer/acme
```

Configure automatic data deletion

```bash
# delete versions after 40s
$ vault kv metadata put -delete-version-after=40s -mount=secret /test                                               # ++
$ vault kv put -mount=secret /test message="data1"
$ vault kv put -mount=secret /test message="data2"
$ vault kv put -mount=secret /test message="data3"
# chk
$ vault kv metadata get -mount=secret /test
```

Check-and-Set (optimisic lock)

```bash
# chk if enabled
$ vault read secret/config
# enable
$ vault write secret/config cas_required=true
# create
$ vault kv put -cas=0 -mount=secret /partner \
	name="Example Co." \
	partner_id="123456789"
# overwrite
$ vault kv put -cas=1 -mount=secret /partner \
	name="Example Co." \
	partner_id="ABCDEFGHIJKLMN"
```

## Q&A

> How do I enter my secrets without exposing the secret in my shell's history?

1. Using dash `-`

```bash
$ vault kv put kv-v1/eng/apikey/Google key=-
...
<Ctrl+d>
```

2. Read from file

```bash
$ tee apikey.json <<EOF
{
  "key": "AAaaBBccDDeeOTXzSMT1234BB_Z8JzG7JkSVxI"
}
EOF
$ vault kv put kv-v1/eng/apikey/Google @apikey.json
```

3. Disable vault cmd history

```bash
$ export HISTIGNORE="&:vault*"
```

> How do I save multiple values at once?

1. Using back slash `\`

```bash
$ vault kv put kv-v1/dev/config/mongodb \
	url=foo.example.com:35533 \
	db_name=users \
	username=admin \
	password=passw0rd
```

2. From file

```bash
tee mongodb.json <<EOF
{
    "url": "foo.example.com:35533",
    "db_name": "users",
    "username": "admin",
    "password": "pa$$w0rd"
}
EOF

$ vault kv put kv-v1/dev/config/mongodb @mongodb.json
```

> How do I store secrets generated by Vault in KV secrets engine?

```bash
# example: get secret id of jenkins role, encoded with base64 as value stored in kv-v1/secret-id
$ vault kv put kv-v1/secret-id \
   jenkins-secret-id="$(vault write -f -field=secret_id auth/approle/role/jenkins/secret-id | base64)"
```

