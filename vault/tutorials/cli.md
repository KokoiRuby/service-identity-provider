### [Starting the server](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-dev-server)

dev server

```bash
$ vault server -dev
$ export VAULT_ADDR='http://127.0.0.1:8200'
$ export VAULT_TOKEN="hvs.6j4cuewowBGit65rheNoceI7"
$ vault status
```

### [Your first secret](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-first-secret)

[Key/Value v2 secrets engine](https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2) is a generic kv store and is enabled at `secret/` path.

Secrets written to Vault are **encrypted** and then written to backend storage.

:warning: Sending data as a part of the CLI ends up in shell history unencrypted, use [Versioned K/V secrets engine](https://developer.hashicorp.com/vault/tutorials/secrets-management/versioned-kv) instead.

```bash
$ vault kv -help

# write
$ vault kv put -help
# -mount=secret is required for v2, otherwise v1
# secret path: secret/data/hello
$ vault kv put -mount=secret hello foo=world
$ vault kv put -mount=secret hello foo=world excited=yes

# read
# secret path: secret/data/hello
$ vault kv get -mount=secret hello
# only value of a given field
$ vault kv get -mount=secret -field=excited hello
# json
$ vault kv get -mount=secret -format=json hello | jq -r .data.data.excited

# delete
$ vault kv delete -mount=secret hello
$ vault kv get -mount=secret hello
# undelete if destroyed is false
$ vault kv undelete -mount=secret -versions=2 hello
$ vault kv get -mount=secret hello
```

### [Secrets engines](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-secrets-engines)

A number of [secrets engines](https://developer.hashicorp.com/vault/docs/secrets) avail. [Secrets Management](https://developer.hashicorp.com/vault/tutorials/secrets-management) tutorial.

Using **kv v1** to quickly demonstrate some concepts → Vault is similar to **virtual filesystem**, which is path-based.

```bash
# The path prefix tells Vault which secrets engine to which it should route traffic.
# since no secret engine is mounted at foo, it ends up an error.
$ vault kv put foo/bar a=b
```

```bash
# enable a secrets engine
$ vault secrets enable -path=kv kv
# or 
# since the path where the secrets engine is enabled defaults to the name of the secrets engine.
$ vault secrets enable kv

# chk
$ vault secrets list

# create secret
$ vault kv put kv/hello target=world
$ vault kv get kv/hello

# create another but diff path
$ vault kv put kv/my-secret value="s3c(eT"
$ vault kv get kv/my-secret

# delete
$ vault kv delete kv/my-secret
$ vault kv list kv/

# diable
$ vault secrets disable kv/
```

### [Dynamic Secrets](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-dynamic-secrets)

Dynamic secrets are generated when they are accessed = on demand. [AWS account](https://aws.amazon.com/) is needed in this demo.

enable aws secrets engine

```bash
# enable aws secrets engine
$ vault secrets enable -path=aws aws

# conf
$ export AWS_ACCESS_KEY_ID=<aws_access_key_id>
$ export AWS_SECRET_ACCESS_KEY=<aws_secret_key>

# write aws cred into secrets engine
$ vault write aws/config/root \
    access_key=$AWS_ACCESS_KEY_ID \
    secret_key=$AWS_SECRET_ACCESS_KEY \
    region=us-east-1
```

```bash
# create a policy → AWS IAM policy that enables all actions on EC2
$ vault write aws/roles/my-role \
        credential_type=iam_user \
        policy_document=-<<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1426528957000",
      "Effect": "Allow",
      "Action": [
        "ec2:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
```

```bash
# gen key pair for the role
$ vault read aws/creds/my-role

# revoke
$ vault lease revoke <lease_id>
```

### [Build-in help](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-help)

```bash
$ vault path-help ...
```

### [Authentication](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-authentication)

Token authentication is **automatically** enabled.

**root** token with `root` policy is used for dev server.

```bash
# create child of root token
$ vault token create
# login with newly-created token
$ vault login
# revoke if no longer needed
$ vault token revoke <token>
```

GitHub AuthN 可通过提供 GitHub cred 获取 Vault token.

Must have a GitHub profile, belong to a team in a GitHub organization, with GitHub Access token `read:org` scope.

```bash
$ vault auth enable github
$ vault auth list
$ vault auth help github

# set org
$ vault write auth/github/config organization=hashicorp
# engineering team in hashicorp is granted to default & application policies
$ vault write auth/github/map/teams/engineering value=default,applications

# login by GitHub Personal Access Token
$ vault login -method=github

# revoke
$ vault login root
$ vault token revoke -mode path auth/github

# disable
$ vault auth disable github
```

### [Policies](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-policies)

Policies in Vault control what a user can access → **AuthZ**.

JSON-compatible **HCL** (Hashicorp Configuration Language).

Based on **prefix matching system on the API path** to determine access control.

Default built-in: `root` & `default`.

```json
# example
# a user could write any secret to secret/data/, except to secret/data/foo
path "secret/data/*" {
  capabilities = ["create", "update"]
}

path "secret/data/foo" {
  capabilities = ["read"]
}
```

```bash
$ vault policy read default
```

write a policy

```bash
$ vault policy write -h
```

```bash
$ vault policy write my-policy - << EOF
# Dev servers have version 2 of KV secrets engine mounted by default
path "secret/data/*" {
  capabilities = ["create", "update"]
}

path "secret/data/foo" {
  capabilities = ["read"]
}
EOF

$ vault policy list
$ vault policy read my-policy
```

test the policy

```bash
# create a token given policy
$ export VAULT_TOKEN="$(vault token create -field token -policy=my-policy)"
# validate if token binds to policy
$ vault token lookup | grep policies

# write a secret to path secret/data/creds
$ vault kv put -mount=secret creds password="my-long-password"
# permission denied
$ vault kv put -mount=secret foo robot=beepboop
```

associate polices to auth method via role

```bash
# chk if enabled
$ vault auth list | grep 'approle/'
# enable
$ vault auth enable approle
# create role, attach to policy
$ vault write auth/approle/role/my-role \
    secret_id_ttl=10m \
    token_num_uses=10 \
    token_ttl=20m \
    token_max_ttl=30m \
    secret_id_num_uses=40 \
    token_policies=my-policy

# verify
$ export ROLE_ID="$(vault read -field=role_id auth/approle/role/my-role/role-id)"
$ export SECRET_ID="$(vault write -f -field=secret_id auth/approle/role/my-role/secret-id)"
$ vault write auth/approle/login role_id="$ROLE_ID" secret_id="$SECRET_ID"
```

### Deploy Vault

via HCL `config.hcl`

```bash
# physical backend that Vault uses for storage
storage "raft" {
  path    = "./vault/data"
  node_id = "node1"
}

# determine how Vault listens for API requests
listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = "true"
}

# Incase Error initializing core: Failed to lock memory: cannot allocate memory
disable_mlock = true

# addr to advertise to route client requests
api_addr = "http://127.0.0.1:8200"
# addr used for comm btw the Vault nodes in a cluster.
cluster_addr = "https://127.0.0.1:8201"
ui = true
```

```bash
$ make -p ./vault/data
# start with conf
$ vault server -config=config.hcl
```

init vault

```bash
$ export VAULT_ADDR='http://127.0.0.1:8200'
# unseal keys & initial root token
$ vault operator init
```

seal/unseal

Every initialized Vault server starts in the *sealed* state.

Unsealing has to happen every time Vault starts via command line or API

```bash
# input unseal keys
$ vault operator unseal
# initial root token
$ vault login
# seal again
$ vault operator seal
```

### [Using HTTP APIs with AuthN](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-apis)

```bash
# conf
$ tee config.hcl <<EOF
storage "file" {
  path = "vault-data"
}

listener "tcp" {
  tls_disable = "true"
}
EOF

# start server with conf
$ vault server -config=config.hcl
```

```bash
# init to get unseal key & token
$ curl \
    --request POST \
    --data '{"secret_shares": 1, "secret_threshold": 1}' \
    http://127.0.0.1:8200/v1/sys/init | j
    
$ export VAULT_TOKEN="s.Ga5jyNq6kNfRMVQk2LY1j9iu"

# unseal
$ curl \
    --request POST \
    --data '{"key": "/ye2PeRrd/qruh9Ppu9EyUjk1vLqIflg1qqw6w9OE5E="}' \
    http://127.0.0.1:8200/v1/sys/unseal | jq
```

```bash
# any auth method can be enabled & configured after unseal
$ vault auth enable -output-curl-string approle
# equivalent to
$ curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
    --request POST \
    --data '{"type": "approle"}' \
    http://127.0.0.1:8200/v1/sys/auth/approle
    
# create policy
$ curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
    --request PUT \
    --data '{"policy":"# Dev servers have version 2 of KV secrets engine mounted by default, so will\n# need these paths to grant permissions:\npath \"secret/data/*\" {\n  capabilities = [\"create\", \"update\"]\n}\n\npath \"secret/data/foo\" {\n  capabilities = [\"read\"]\n}\n"}' \
    http://127.0.0.1:8200/v1/sys/policies/acl/my-policy

# enable secrets engine
$ curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
    --request POST \
    --data '{ "type":"kv-v2" }' \
    http://127.0.0.1:8200/v1/sys/mounts/secret
    
# create role associate with policy
$ curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
    --request POST \
    --data '{"policies": ["my-policy"]}' \
    http://127.0.0.1:8200/v1/auth/approle/role/my-role

# get role id & secret id
$ curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
     http://127.0.0.1:8200/v1/auth/approle/role/my-role/role-id | jq -r ".data"
$ curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
    --request POST \
    http://127.0.0.1:8200/v1/auth/approle/role/my-role/secret-id | jq -r ".data"

# login
$ curl --request POST \
       --data '{"role_id": "3c301960-8a02-d776-f025-c3443d513a18", "secret_id": "22d1e0d6-a70b-f91f-f918-a0ee8902666b"}' \
       http://127.0.0.1:8200/v1/auth/approle/login | jq -r ".auth"
       
# set token
$ export VAULT_TOKEN="s.p5NB4dTlsPiUU94RA5IfbzXv"
# create a secret
$ curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
    --request POST \
    --data '{ "data": {"password": "my-long-password"} }' \
    http://127.0.0.1:8200/v1/secret/data/creds | jq -r ".data"
```

clean up

```bash
$ unset VAULT_TOKEN
$ rm -r vault-data
```

### [UI](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-ui)

