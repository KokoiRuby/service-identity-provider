### [What is Vault?](https://developer.hashicorp.com/vault/tutorials/get-started/what-is-vault)

**Vault** can be used to secure, store, and control access to tokens, passwords, certificates, and encryption keys for **protecting secrets and other sensitive data**.

- **Data encryption 数据加密**：保证用户数据在传输和持久全程加密。
- **Access control 访问控制**：**ACL polices** 绑定 Identity；**Control groups** 执行第三方的授权；**Sentinel policies** 提供丰富的逻辑条件。
- **Time boxed access 时间窗口访问**：TTL 关联 credentiasls & leases，超时自动撤销。
- **Disaster recovery support 灾难恢复支持**：主从 primary/secondary 集群数据复制，自动数据快照。
- **Performance scaling 性能扩容**：active for write & standy for read。
- **Identity-based security 基于身份安全性**：通过身份标识客户端并绑定附着 polices 的 token。
- **Anti Secrets sprawl 反密钥蔓延**：集中存储/管理/访问。
- **Cloud and vendor agnostic 跨云平台 & 供应商**：not specific to。
- **Human and machine authentication 人机验证**：username/pwd or k8s workloads。
- **Secrets engines 密钥引擎**：static/dynamic (with expiry)。

### [Avail editions](https://developer.hashicorp.com/vault/tutorials/get-started/available-editions#available-vault-editions)

Community vs. Enterprise

### [Plugins](https://developer.hashicorp.com/vault/tutorials/get-started/discover-plugins)

Vault 通过基于插件框架的**内置插件**来提供所有功能。

- Auth plugins 提供与不同服务的集成，以允许个人或工作负载使用 Vault 进行身份验证。
- Secret engine plugins 存储 static 密钥以及即时生成 dynamic 密钥用于不同平台。
- Database plugins 扩展 secret engine 与允许创建 dynamic 密钥。

**内置插件生命周期**：Register → Enable → Configure → Use → Disable。

**内置插件工作流**：

1. 在**指定路径**下启用插件实例
2. 配置插件
3. 为 Database 插件创建 role
4. 升级插件二进制文件
5. 禁用插件

**外部插件**：Vault 可以加载指定目录 `plugin_directory` 下的外部插件。

**外部插件工作流**：

1. 配置 `plugin_directory`
2. 将外部插件二进制置于 `plugin_directory` 下
3. 使用 register plugin API 或 Vault CLI 注册插件
4. 在**指定路径**下启用外部插件实例
5. 配置插件
6. 为 Database 插件创建 role
7. 升级插件二进制文件
8. 禁用插件

**容器化插件**：Vault 支持运行容器化的插件，但有一些约束：

- Vault server 必须运行在 Linux 上，且可以访问 Docker Engine API
- Vault server 所在环境必须安装 Container Runtime
- 必须要手动拉取镜像到 Vault server 所在环境

### [Install](https://developer.hashicorp.com/vault/tutorials/get-started/install-binary)

[Linux](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-install)

```bash
# export VAULT_ADDR='http://127.0.0.1:8200'
# export VAULT_TOKEN="hvs.2Q3xkL2Ub8UDzfRXnOIul8ub"
# Unseal Key: jfWjrX2AsFDOOwCZvX9/YkwLythaAl3Bu+J50+gYAcQ=
# Root Token: hvs.2Q3xkL2Ub8UDzfRXnOIul8ub
$ vault server -dev
```

[Helm](https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-raft-deployment-guide)

```bash
# Unseal Key 1: MrslTDlWKPxbSFNvineS3mFHvzIl45kBnAb68icL0rVh
# Unseal Key 2: /haXjcjWfFT+Cw0tP1MCB7bvCAxRpg0EQ4b85k+y3fVy
# Unseal Key 3: UUaEJ+7O/iCKYtHJEFbmlD6tr6DUwdorGGV2nLCsXFPN
# Unseal Key 4: xc1YleYMi0eMXr8kiGAa8Q9NWt/+O1CrGO9yCXjq44SL
# Unseal Key 5: NrBBEUfDPwV/ITz086FU7g7HgwEBpuY9Su5ayLHiJUqa

# Initial Root Token: hvs.Fzn5qNB5io7RLzzMzc3UcUT6
$ vault operator init

# At least 3 of these keys to unseal it
$ vault operator unseal 
```

### [Set up Vault](https://developer.hashicorp.com/vault/tutorials/get-started/setup)

| **Dev mode Vault**          | **Self-managed Vault** :ballot_box_with_check: | **HCP Vault Dedicated**                   |
| :-------------------------- | :--------------------------------------------- | :---------------------------------------- |
| In-memory 存储后端          | 可配置存储后端                                 | 已集成存储后端                            |
| 自动初始化和解封            | 需手动初始化和解封                             | 自动初始化和解封                          |
| 单密钥密封                  | 可配置密封                                     | 云自动密封                                |
| 运行时自动初始化 root token | 解封时输出 root token                          | 无 root token，HCP UI 中生成 admin tokens |

```bash
$ mkdir conf tls vault-data
$ openssl req -x509 -newkey rsa:4096 -sha256 -days 365 \
	-nodes -keyout ./tls/vault-key.pem -out ./tls/vault-cert.pem \
    -subj "/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
```

```bash
# Vault server conf (min.)
$ cat > ./conf/vault-server.hcl << EOF
# Global
# server API addr for clients to interact with Vault
api_addr                = "https://127.0.0.1:8200"
# server cluster add for intra-cluster comm
cluster_addr            = "https://127.0.0.1:8201"
# server cluster name for all servers participating in
cluster_name            = "learn-vault-cluster"
# for servers which use integrated storage only, disable operating system memory locking for the Vault process
disable_mlock           = true
# web ui enabler
ui                      = true

# listener defines how & where the Vault server listens
# TLS is enabled by default if tcp
listener "tcp" {
address       = "127.0.0.1:8200"
tls_cert_file = "./tls/vault-cert.pem"
tls_key_file  = "./tls/vault-key.pem"
}

# storage
backend "raft" {
path    = "./vault-data"
node_id = "learn-vault-server"
}
EOF
```

```bash
$ vault server -config=./conf/vault-server.hcl
```

**New terminal**

```bash
$ export VAULT_ADDR=https://127.0.0.1:8200
$ export VAULT_SKIP_VERIFY=true

# Unseal Key 1: A/s04AjMQTWZeOgjaVCLRIwbBf2c6Q8HrZxNFLm0gO8=
# Initial Root Token: hvs.zlC0lWZ3ReP0UxZr5U63yXCB
$ vault operator init -key-shares=1 -key-threshold=1

# against unseal key
$ vault operator unseal

# chk
$ vault status
```

Authenticate

```bash
# against root token
$ vault login
```

### [Tokens](https://developer.hashicorp.com/vault/tutorials/get-started/introduction-tokens)

Vault 支持两种验证 `userpass` for human & `kubernetes` for workload。

#### **Flow**

1. 客户端认证 Vault
2. Vault 通过 trusted provider 验证其身份
3. 验证成功后会颁发一个 token 给客户端。

初次启动 Vault 会返回 Root token 具备所有权限。`token` auth method 生成 token 并将 `root` policy 与之绑定。

#### **Token meta**

```bash
Key                  Value
---                  -----
token                hvs.5k2fSn7gmsLyotTp8roccTkK
# uid that can be used to lookup, renew, or revoke a token
token_accessor       uyDvDGJiLVLTAOuez3SbhPYp
# ttl, default is 32 days
token_duration       ∞
token_renewable      false
token_policies       ["root"]
# policies attached to token for authorization
identity_policies    []
policies             ["root"]

# more details
$ vault token lookup -accessor uyDvDGJiLVLTAOuez3SbhPYp
```

#### Types of tokens

By prefix

- Service tokens  `hvs.string` [vs.](https://developer.hashicorp.com/vault/tutorials/get-started/introduction-tokens#service-tokens-vs-batch-tokens) Batch tokens `hvb.string`
- Recovery tokens `hvr.string`

Others

- **Periodic tokens** created by root, with TTL but no max ttl.
- **Orphan tokens** created by root, do not expire when their parent does.

#### Renew

`-period` 指定 Token 的有效期限

`-explicit-max-ttl` 指定 Token 的明确最大生命周期；即使 `-period` 设置了更长的时间，也会受到 `-explicit-max-ttl` 的限制。

```bash
$ vault token create -policy="default" -period=1m -explicit-max-ttl=2m
$ vault token renew -accessor <accessor>
```

#### Revoke

```bash
$ vault token revoke <accessor>
```

### [Policies](https://developer.hashicorp.com/vault/tutorials/get-started/introduction-policies)

A declarative way to **grant or forbid access** to operations in Vault.

Policies 通过绑定一个 token 来生效。

Policies 使用 HashiCorp 配置语言 (HCL)，通过 API **路径前缀匹配**来查找有效的访问控制。

#### Create a policy

```bash
path "secret/data/creds" {
  capabilities = ["create", "update"]
}

path "secret/data/creds/confidential" {
  capabilities = ["read"]
}

path "dev-secrets/data/creds" {
  capabilities = ["create", "list", "read", "update"]
}

path "sre-secrets/data/creds" {
  capabilities = ["create", "list", "read", "update"]
}

# wildcard
path "dev-secrets/data/*" {
  capabilities = ["create", "list", "read", "update"]
}

path "dev-secrets/data/cred*" {
  capabilities = ["create", "list", "read", "update"]
}

# + wildcard
path "dev-secrets/+/creds" {
  capabilities = ["create", "list", "read", "update"]
}
```

#### Use templates

```bash
# identity.entity.name is the name of user authn to Vault
path "dev-secrets/+/creds/{{identity.entity.name}}" {
  capabilities = ["create", "list", "read", "update"]
}
```

#### Constraints

指定请求必须要包含的参数列表，否则拒绝。

```bash
path "dev-secrets/+/creds" {
  capabilities = ["create", "list", "read", "update"]
  required_parameters = ["username"]
}
```

#### Explicit deny

`deny` will always take precedence over other capabilities.

```bash
path "dev-secrets/+/*" {
  capabilities = ["create", "list", "read", "update"]
}

path "dev-secrets/+/root" {
  capabilities = ["deny"]
}
```

#### Sentinel

An **advanced form of** policies available with HCP Vault Dedicated and Vault Enterprise.

### [Roles](https://developer.hashicorp.com/vault/tutorials/get-started/introduction-roles)

A collection of parm that you group together to simplify plugin configuration 用于简化插件配置的参数集合。

UC: user auth plugins & secret mgmt plugins 只需传递角色名，而不用通过传递多个 arg。

#### Create a role

```bash
# enable auth method at auth/kubernetes
$ vault auth enable kubernetes

# integrate so that Vault can connect to K8s
$ vault write auth/kubernetes/config \
    token_reviewer_jwt="$K8S_SERVICE_ACCOUNT_TOKEN" \
    kubernetes_host=https://192.168.99.100:443 \
    kubernetes_ca_cert=@ca.crt

# create a role, and bind to service account given policies
# so that any pods with this service account can authn with Vault & have access against policies
$ vault write auth/kubernetes/role/hashicupsApp \
     bound_service_account_names=k8sHashicupsAppSA \
     bound_service_account_namespaces=k8sDevNamespace \
     policies=default,dev-secrets \
     ttl=1h \
     
# pod login to Vault
$ POD_TOKEN=$(kubectl exec exampleapp-auth -- curl --silent \
   --request POST \
   --header "X-Vault-Namespace: $VAULT_NAMESPACE" \
   --data '{"jwt": "'$k8sHashicupsAppSA_TOKEN'", "role": "hashicupsApp"}' \
   $VAULT_ADDR/v1/auth/kubernetes/login | jq -r '.auth | .client_token')
```

### Learn to use the Vault [CLI](https://developer.hashicorp.com/vault/tutorials/get-started/learn-cli)

**setup the lab**

```bash
$ vault server -dev -dev-root-token-id root -dev-tls
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_CACERT='/tmp/vault-tls1395032274/vault-ca.pem'
$ vault login root
```

**server status**

```bash
$ vault status --help | head -n 12
$ vault status
$ vault status -format=json
# vault <command> [options] [path] [args]
$ vault help
```

**auth method**

```bash
$ vault auth enable userpass
$ vault auth list
# path help
$ vault path-help /auth/userpass
```

**policy**

```bash
# get
$ vault kv get -output-policy dev-secrets/creds

# create policy
$ vault policy write developer-vault-policy - << EOF
path "dev-secrets/+/creds" {
   capabilities = ["create", "list", "read", "update"]
}
EOF

# craete a username & password
vault write /auth/userpass/users/danielle-vault-user \
    password='Flyaway Cavalier Primary Depose' \
    policies=developer-vault-policy
```

secrets engine

```bash
$ vault secrets list
# enable v2 kv under /dev-secrets
$ vault secrets enable -path=dev-secrets -version=2 kv
```

authenticate & create secret

```bash
$ vault login -method=userpass username=danielle-vault-user
# or explicitly
$ vault login \
    -no-print \
    -method=userpass \
    username=danielle-vault-user \
    password='Flyaway Cavalier Primary Depose'
    
# put a new secret in secret engine at /dev-secrets
$ vault kv put /dev-secrets/creds api-key=E6BED968-0FE3-411E-9B9B-C45812E4737A
$ vault kv get /dev-secrets/creds
```

clean up

```bash
$ pkill vault
$ unset VAULT_ADDR VAULT_CACERT
```

### Learn to use the Vault [UI](https://developer.hashicorp.com/vault/tutorials/get-started/learn-ui)

**setup the lab**

```bash
$ vault server -dev -dev-root-token-id root -dev-tls
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_CACERT='/tmp/vault-tls1395032274/vault-ca.pem'
$ vault login root
```

https://127.0.0.1:8200/ui

https://127.0.0.1:8200/v1/sys/seal-status

### Learn to use the Vault HTTP [API](https://developer.hashicorp.com/vault/tutorials/get-started/learn-http-api)

**setup the lab**

```bash
$ vault server -dev -dev-root-token-id root -dev-tls
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ VAULT_CACERT='/tmp/vault-tls2208524476/vault-ca.pem'
$ export VAULT_TOKEN=root
```

**check server status**

```bash
$ export CURL_CA_BUNDLE=$VAULT_CACERT
$ curl -s $VAULT_ADDR/v1/sys/seal-status | jq
# same to 
$ vault read /sys/seal-status
```

**headers and paths**

Vault 要求每个请求必须要将 token 置于 `X-Vault-token` header，设置了 ENV `VAULT_TOKEN`，会自动填充。

```bash
$ vault status -output-curl-string
```

auth method

```bash
# enable userpass auth method
# chk console log to see if it was successful
$ curl \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -X POST \
  -d '{"type": "userpass"}' \
  $VAULT_ADDR/v1/sys/auth/userpass

# or
$ curl \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  $VAULT_ADDR/v1/sys/auth  | jq ".data"
  
# create user in userpass
$ curl \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -X POST \
  -d '{"password":"Imprint Bacteria Marathon Aflutter","token_policies":"developer-vault-policy"}' \
  $VAULT_ADDR/v1/auth/userpass/users/danielle-vault-user

# create policy
$ curl \
    -H "X-Vault-Token: $VAULT_TOKEN" \
    -X PUT \
    -d '{"policy":"path \"dev-secrets/data/creds\" {\n  capabilities = [\"create\", \"update\"]\n}\n\npath \"dev-secrets/data/creds\" {\n  capabilities = [\"read\"]\n}\n"}' \
    $VAULT_ADDR/v1/sys/policies/acl/developer-vault-policy
    
# chk
$ curl -s -H "X-Vault-Token: $VAULT_TOKEN" $VAULT_ADDR/v1/sys/policy | jq ".data.policies"
```

secrets engine

```bash
# chk mounted 
$ curl -s \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  $VAULT_ADDR/v1/sys/mounts | jq ".data"
  
# create
curl \
    -H "X-Vault-Token: $VAULT_TOKEN" \
    -X POST \
    -d '{ "type":"kv-v2" }' \
    $VAULT_ADDR/v1/sys/mounts/dev-secrets
```

authn & create secrets

```bash
# login
$ curl -s \
  -X POST \
  -d '{ "password": "Imprint Bacteria Marathon Aflutter" }' \
  $VAULT_ADDR/v1/auth/userpass/login/danielle-vault-user | jq ".auth.client_token"

# ENV
$ export DANIELLE_DEV_TOKEN="hvs.CAESIIwTHqwzadSQG0B3cX3doR0SLvj_b5WoshPeClc_N7jvGh4KHGh2cy5ob2ZvWVM0Nnp0dXZvWTBiTGNTWXZld2U"

# create secret
$ curl -s \
  -H "X-Vault-Token: $DANIELLE_DEV_TOKEN" \
  -X PUT \
  -d '{ "data": {"password": "Driven Siberian Pantyhose Equinox"} }' \
  $VAULT_ADDR/v1/dev-secrets/data/creds | jq ".data"

# show secret
$ curl -s \
  -H "X-Vault-Token: $DANIELLE_DEV_TOKEN" \
  $VAULT_ADDR/v1/dev-secrets/data/creds | jq ".data"
```

clean up

```bash
$ unset VAULT_TOKEN && unset VAULT_ADDR && unset DANIELLE_DEV_TOKEN
$ pkill vault
```

### Learn to use the Vault [Terraform provider](https://developer.hashicorp.com/vault/tutorials/get-started/learn-terraform)

[Install terraform](https://developer.hashicorp.com/terraform/install?product_intent=terraform)

Vault provider 使用一系列配置文件通过 Vault HTTP API 与 Vault 进行交互。

setup the lab

```bash
$ vault server -dev -dev-root-token-id root -dev-tls
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_CACERT='/tmp/vault-tls1137972627/vault-ca.pem'
$ export VAULT_TOKEN=root
```

vault terraform provider

```bash
$ git clone https://github.com/hashicorp-education/learn-vault-foundations.git
$ cd learn-vault-foundations/terraform/oliver
# init terra conf
$ terraform init
# chk plan
$ terraform plan
# apply change to vault
$ terraform apply -auto-approve
```

auth method

```bash
# enable userpass auth method with Terraform
resource "vault_auth_backend" "userpass" {
  type = "userpass"
}

# create user in userpass
resource "vault_generic_endpoint" "danielle-user" {
   path                 = "auth/${vault_auth_backend.userpass.path}/users/danielle-vault-user"
   ignore_absent_fields = true
   data_json = <<EOT
{
   "token_policies": ["developer-vault-policy"],
   "password": "Vividness Itinerary Mumbo Reassure"
}
EOT
}

# assign default policy to user
resource "vault_policy" "developer-vault-policy" {
   name = "developer-vault-policy"

   policy = <<EOT
   path "dev-secrets/+/creds" {
   capabilities = ["create", "update"]
    }
    path "dev-secrets/+/creds" {
       capabilities = ["read"]
    }
    ## Vault TF provider requires ability to create a child token
    path "auth/token/create" {
       capabilities = ["create", "update", "sudo"]
    }
EOT
}

# chk on vault
$ vault policy list
```

static secrets

```bash
# create secret engine
resource "vault_mount" "dev-secrets" {
   path        = "dev-secrets"
   type        = "kv"
   options     = { version = "2" }
}

# chk
$ vault secrets list
```

developer conf

```bash
$ cd terraform/danielle/
$ export VAULT_TOKEN=root && export VAULT_ADDR='https://localhost:8200' && export VAULT_CACERT=<<YOUR_CA_LOC_HERE>>
$ terraform init

# hardcoded in provider.tf
$ variable login_username {
   type = string
   default = "danielle-vault-user"
}

# pass
$ terraform plan

# pass
$ terraform apply -auto-approve
```

developer cred

```bash
# var.login_username & var.login_password in provider.tf
provider "vault" {
   auth_login {
      path = "auth/userpass/login/${var.login_username}"
      parameters = {
         password = var.login_password
      }
   }
}
```

```bash
# main.tf
# create secret in secret engine
resource "vault_kv_secret_v2" "creds" {
  mount                      = "dev-secrets"
  name                       = "creds"
  data_json                  = jsonencode(
    {
       password = "Vividness Itinerary Mumbo Reassure",
    }
  )
}

# chk
$ vault read /dev-secrets/data/creds
```

clean up

```bash
$ pkill vault
$ unset VAULT_TOKEN && unset VAULT_ADDR && unset VAULT_CACERT
```

### [Static vs. Dynamic secrets](https://developer.hashicorp.com/vault/tutorials/get-started/understand-static-dynamic-secrets)

KV secrets engine is the **most commonly used** engine for static secrets.

set up lab

```bash
$ vault server -dev -dev-root-token-id root -dev-tls
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_CACERT='/tmp/vault-tls2181733608/vault-ca.pem'
$ export VAULT_TOKEN=root
```

```bash
$ vault secret list
$ vault secrets enable -path=kvv2 kv-v2
$ vault secrets disable kvv2
```

While *Dynamic secrets* **do not exist until read**, revoked after use.

vault 提供的数据库插件，配置连接数据，使用 HTTP API 创建 dynamic cred。

```bash
$ docker pull postgres:latest
$ docker run \
    --detach \
    --name learn-postgres \
    -e POSTGRES_USER=root \
    -e POSTGRES_PASSWORD=rootpassword \
    -p 5432:5432 \
    --rm \
    postgres
$ docker ps -f name=learn-postgres --format "table {{.Names}}\t{{.Status}}"

# create role
$ docker exec -i \
    learn-postgres \
    psql -U root -c "CREATE ROLE \"ro\" NOINHERIT;"
    
# grant read all tables to role
$ docker exec -i \
  learn-postgres \
  psql -U root -c "GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"ro\";"
```

```bash
# create db plugin
$ vault secrets enable database
$ export POSTGRES_URL="127.0.0.1:5432"

# conf conn cred
$ vault write database/config/postgresql \
  plugin_name=postgresql-database-plugin \
  connection_url="postgresql://{{username}}:{{password}}@$POSTGRES_URL/postgres?sslmode=disable" \
  allowed_roles=readonly \
  username="root" \
  password="rootpassword"

# help
$ vault path-help database/config/postgresql

# read
$ vault read database/config/postgresql
```

create a role

```bash
# create api request payload
$ tee readonly.sql <<EOF
CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}' INHERIT;
GRANT ro TO "{{name}}";
EOF

# craete role
$ vault write database/roles/readonly \
    db_name=postgresql \
    creation_statements=@readonly.sql \
    default_ttl=1h \
    max_ttl=24h
```

request db cred & validate

```bash
# get username
$ vault read database/creds/readonly


$ docker exec -i \
  learn-postgres \
  psql -U root -c "SELECT usename, valuntil FROM pg_user;"
```

clean up

```bash
$ unset VAULT_TOKEN && unset VAULT_ADDR && unset VAULT_CACERT && unset POSTGRES_URL
$ docker stop $(docker ps -f name=learn-postgres -q)
$ pkill vault
```