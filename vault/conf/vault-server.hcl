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
