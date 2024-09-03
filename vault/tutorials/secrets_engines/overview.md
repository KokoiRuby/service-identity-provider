[Secrets engines](https://developer.hashicorp.com/vault/docs/secrets) are components which **store, generate, or encrypt data**.

Secrets engines are enabled at a **path** in Vault. 类 virtual FS 来向请求会被自动路由到匹配路径的 secrets engines。

**Lifecycle**

- [Enable](https://developer.hashicorp.com/vault/docs/commands/secrets/enable) 在指定路径下启用；默认路径根据 secrets engine 的类型，比如 aws，那么就在 `aws/`。
- [Disable](https://developer.hashicorp.com/vault/docs/commands/secrets/disable) 在指定路径下禁用；所有的 secrets 会被撤销，后端存储删除。
- [Move](https://developer.hashicorp.com/vault/docs/commands/secrets/move) 移动到另一个路径；所有的 secrets 会被撤销，后端存储依旧保留。
- [Tune](https://developer.hashicorp.com/vault/docs/commands/secrets/tune) 调整全局配置，比如 TTLs。

```bash
$ vault secrets path-help
```

**Barrier view**

[chroot](https://en.wikipedia.org/wiki/Chroot)-like；当一个 secrets engine 被启用时，会生成一个随机的 UUID。

每当 engine 写入后端存储时，都会前缀带上该 UUID。

**Default secrets engines**

- `cubbyhole/` 每个 token 私有的秘密存储引擎，只有拥有 token 持有者能访问其中的数据，保证数据的隔离性和安全性。
- `identity/` 管理和存储与实体 identity 身份相关的信息。
- `sys/` 管理 Vault 实例的控制、策略和调试。

