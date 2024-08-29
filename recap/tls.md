## TLS

传输层安全性 Transport Layer Security，是目前广泛采用的安全协议，主要应用于 Web 应用。

TLS 实现功能：**数据加密** + **身份认证** + **数据完整性**。

**如何使用 TLS？** 服务器安装 TLS 证书。

**如何工作？**

1. 协商 TLS 版本
2. 决定使用的密码套件 suite
3. 使用服务器 TLS 证书验证服务器身份
4. 生成会话密钥进行加密

最新版本 **TLS 1.3** since 2018 更快更安全，握手只需一次往返 (Hello & Key Exchange 合并)

1. 基于 False Start，在握手前就开始传输数据。
2. TLS 会话恢复，若此前已经连接过网站，那么就不再需要进行握手，实现往返次数为 0。



![img](https://cyberhoot.com/wp-content/uploads/2020/02/ssl2buy-tls12-13.jpg)

### Suite

TLS 套件是加密算法的组合，由以下部分组成：

- **密钥交换算法 Key Exchange Algorithm** 

- **数字签名算法 Digital Signature Algorithm** 身份认证

- **对称加密算法 Symmetric Encryption Algorithm** 会话加密

- **消息认证码算法 Message Authentication Code Algorithm** 摘要/哈希算法

  

![TLS Essentials 10: TLS cipher suites explained](https://i.ytimg.com/vi/mFdDap9A9-Q/maxresdefault.jpg)