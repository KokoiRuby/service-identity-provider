## [High Availability](https://developer.hashicorp.com/vault/docs/internals/high-availability)

HA mode to **minimize downtime** without affecting horizontal scalability.

Vault **IO limits** reside in **storage backend** rather than compute.

Note: storage is for persistence of encrypted data.

**Integrated Storage**:

- Raft (Cluster)
- File (Single-node)

**State** in HA:

- `Actice` processes all requests.
- `Standby` redirects all requests to Active.
  - Escalated to Active automatically if Active is down (sealed/failed/network partition).
  - Only *unsealed* Vault servers may act as a standby.

### [Architecture](https://developer.hashicorp.com/vault/tutorials/day-one-raft/raft-reference-architecture)

### [Vault HA cluster with integrated storage](https://developer.hashicorp.com/vault/tutorials/raft/raft-storage)

- **vault_1** is initialized and unsealed. 
  - The root token creates a transit key that enables the other Vaults auto-unseal. 
  - This Vault does not join the cluster.
- **vault_2** is initialized and unsealed. 
  - This Vault starts as the cluster leader. 
  - An example K/V-V2 secret is created.
- **vault_3** is only started. **You will join it to the cluster.**
- **vault_4** is only started. **You will join it to the cluster.**

![Scenario](https://developer.hashicorp.com/_next/image?url=https%3A%2F%2Fcontent.hashicorp.com%2Fapi%2Fassets%3Fproduct%3Dtutorials%26version%3Dmain%26asset%3Dpublic%252Fimg%252Fvault-raft-1.png%26width%3D1104%26height%3D564&w=3840&q=75&dpl=dpl_6LRbPJXp6xCiWFmaC7rrQSZkVi4v)

```bash
$ git clone https://github.com/hashicorp-education/learn-vault-raft.git
$ cd learn-vault-raft/raft-storage/local
$ chmod +x cluster.sh

$ ./cluster.sh create network
$ ./cluster.sh create config
$ ./cluster.sh setup vault_1
$ ./cluster.sh setup vault_2
$ ./cluster.sh setup vault_3
```

#### Create an HA cluster

**Examine the leader**

```bash
$ config-vault_2.hcl
$ export VAULT_ADDR="http://127.0.0.2:8200"
# vault_2 is the only node and is currently the leader
$ vault operator raft list-peers
# root token is saved to file during setup
$ cat root_token-vault_2 
```

**Join nodes to the cluster**

```bash
# vault_3
$ cd learn-vault-raft/raft-storage/local
$ export VAULT_ADDR="http://127.0.0.3:8200"
# join vault_2
$ vault operator raft join http://127.0.0.2:8200
# root token of vault_2
$ export VAULT_TOKEN=$(cat root_token-vault_2)
# vault_3 joined as follower
$ vault operator raft list-peers
# read secret
$ vault kv get kv/apikey
```

**Retry join**, if the connection details of all the nodes are known beforehand

```hcl
// # ++ in config-vault_4.hcl
storage "raft" {
   path    = "/root/go/src/vault/ha/learn-vault-raft/raft-storage/local/raft-vault_4/"
   node_id = "vault_4"
   // ++
   retry_join {
      leader_api_addr = "http://127.0.0.2:8200"
   }
   retry_join {
      leader_api_addr = "http://127.0.0.3:8200"
   }
}
```

```bash
$ ./cluster.sh start vault_4
$ export VAULT_ADDR="http://127.0.0.4:8200"
# vault_3 joined as follower
$ vault operator raft list-peers
# root token of vault_2
$ export VAULT_TOKEN=$(cat root_token-vault_2)
# patch 
$ vault kv patch kv/apikey expiration="365 days"
# chk from vault_3
$ vault kv get kv/apikey
```

**Data snapshots for recovery**

Note: Automated snapshots require **Vault Enterprise 1.6.0** or later.

```bash
# back to vault_2
$ vault operator raft snapshot save demo.snapshot
```

**Simulate loss of data**

```bash
# delete data
$ vault kv metadata delete kv/apikey
```

**Restore from snapshot**

```bash
$ vault operator raft snapshot restore demo.snapshot
# chk
$ vault kv get kv/apikey
```

**Resign from active duty**

```bash
$ vault operator step-down
# chk
$ vault operator raft list-peers
```

**Remove a cluster member** (any members can do?)

```bash
$ vault operator raft remove-peer vault_4
# chk
$ vault operator raft list-peers
```

**Add it back**

```bash
$ ./cluster.sh stop vault_4
# clear data
$ rm -rf raft-vault_4/*
$ ./cluster.sh start vault_4
```

**Recovery mode**, in the case of an outage caused by corrupt entries in the storage backend

```bash
# stop all
$ ./cluster.sh stop vault_2
$ ./cluster.sh stop vault_3
$ ./cluster.sh stop vault_4
# start vault_3 in recovery mode
$ VAULT_TOKEN=$(cat root_token-vault_1) VAULT_ADDR=http://127.0.0.3:8200 \
            vault server -recovery -config=config-vault_3.hcl
# generate otp
$ export VAULT_ADDR="http://127.0.0.3:8200"
$ vault operator generate-root -generate-otp -recovery-token
# start the generation of the recovery token with the otp
$ vault operator generate-root -init \
    -otp=wBxwwG26MeWV5qggDkOkPraw6LqC -recovery-token
# verify recovery key
$ cat recovery_key-vault_2
# create encoded token
$ vault operator generate-root -recovery-token
# complete the creation of a recovery token
$ vault operator generate-root \
  -decode=HzQKWUYudVo4NWAgczUSUwcsFRwKBCc5BwsZBQ \
  -otp=wBxwwG26MeWV5qggDkOkPraw6LqC \
  -recovery-token
# verify
$ VAULT_TOKEN=hvr.1iGluP7vFDu4CGZwZvFN1GhF vault list sys/raw/sys
```

**Resume normal operations**

```bash
$ ./cluster.sh start vault_3
$ ./cluster.sh start vault_2
```

**Clean up**

```bash
$ ./cluster.sh clean
```



