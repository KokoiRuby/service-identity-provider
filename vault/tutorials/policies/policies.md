[Policies](https://developer.hashicorp.com/vault/docs/concepts/policies) provide a **declarative** way to **grant or forbid** access to certain paths and operations.

Policies are **deny by default (blacklist)**, so an empty policy grants no permission in the system.

**AuthN via auth method (as delegation) before AuthZ**

![Vault Auth Workflow](https://developer.hashicorp.com/_next/image?url=https%3A%2F%2Fcontent.hashicorp.com%2Fapi%2Fassets%3Fproduct%3Dvault%26version%3Drefs%252Fheads%252Frelease%252F1.17.x%26asset%3Dwebsite%252Fpublic%252Fimg%252Fvault-auth-workflow.svg%26width%3D669%26height%3D497&w=1920&q=75&dpl=dpl_AVrZmHXNKqb16fSStR7krRaRWA9X)

### Syntax

JSON-compatible [HCL](https://github.com/hashicorp/hcl)

```json
path "secret/foo" {
  capabilities = ["read"]
}

// This section grants all access on "secret/*". further restrictions can be
// applied to this broad policy, as shown below.
path "secret/*" {
  capabilities = ["create", "read", "update", "patch", "delete", "list"]
}

// Even though we allowed secret/*, this line explicitly denies
// secret/super-secret. this takes precedence.
path "secret/super-secret" {
  capabilities = ["deny"]
}

// Policies can also specify allowed, disallowed, and required parameters. 
// Here the key "secret/restricted" can only contain 
// "foo" (any value) and "bar" (one of "zip" or "zap").
path "secret/restricted" {
  capabilities = ["create"]
  allowed_parameters = {
    "foo" = []
    "bar" = ["zip", "zap"]
  }
}
```

glob pattern

```json
// Permit reading only "secret/foo". 
// An attached token cannot read "secret/food" or "secret/foo/bar".
path "secret/foo" {
  capabilities = ["read"]
}

// Permit reading everything under "secret/bar". 
// An attached token could read "secret/bar/zip", "secret/bar/zip/zap"
// But not "secret/bars/zip".
path "secret/bar/*" {
  capabilities = ["read"]
}

// Permit reading everything prefixed with "zip-".
// An attached token could read "secret/zip-zap" 
// Or "secret/zip-zap/zong", but not "secret/zip/zap
path "secret/zip-*" {
  capabilities = ["read"]
}

```

`+` can be used to denote any number of characters bounded within a single path segment

```json
// Permit reading the "teamb" path under any top-level path under secret/
path "secret/+/teamb" {
  capabilities = ["read"]
}

// Permit reading secret/foo/bar/teamb, secret/bar/foo/teamb, etc.
path "secret/+/+/teamb" {
  capabilities = ["read"]
}
```

### Matching

Potentially multiple matching policy paths, `P1` and `P2`.

1. If the first wildcard (`+`) or glob (`*`) occurs earlier in `P1`, `P1` is lower priority
2. If `P1` ends in `*` and `P2` doesn't, `P1` is lower priority
3. If `P1` has more `+` (wildcard) segments, `P1` is lower priority
4. If `P1` is shorter, it is lower priority
5. If `P1` is smaller lexicographically, it is lower priority

### Capabilities

- `create` POST/PUT
- `read` GET
- `update` POST/PUT
- `patch` PATCH
- `delete` DELETE
- `list` LIST
- `sudo` Allows access to paths that are *root-protected*.
- `deny` Disallows access. it takes precedence regardless of any other defined capabilities
- `subscribe` Allows subscribing to [events](https://developer.hashicorp.com/vault/docs/concepts/events) for the given path.

### [Templated policies](https://developer.hashicorp.com/vault/docs/concepts/policies#templated-policies)

Variable replacement in some policy strings with values available to the token.

```json
// Example
// policies against kv2 secret engine to a specific user in the token
path "secret/data/{{identity.entity.id}}/*" {
  capabilities = ["create", "update", "patch", "read", "delete"]
}

path "secret/metadata/{{identity.entity.id}}/*" {
  capabilities = ["list"]
}

```

### [Fine-grained control](https://developer.hashicorp.com/vault/docs/concepts/policies#fine-grained-control)

### Built-in

- `default` applies to all tokens by default.

```bash
$ vault read sys/policy/default
```



- `root`

### Managing

