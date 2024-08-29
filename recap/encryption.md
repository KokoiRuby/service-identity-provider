## Key Algorithms

### Symmetric

加解密使用同一个密钥 :smile: 加解密速度快 :cry: 传输过程不安全，易破解，不易管理。一般密钥的长度越长，越安全。

|      | Desc                            | Prod                    |
| ---- | ------------------------------- | ----------------------- |
| DES  | Data Encryption Standard        | ❌                       |
| 3DES | Triple Data Encryption Standard | ❌                       |
| AES  | Advanced Encryption Standard    | :ballot_box_with_check: |

### Asymmetric

加解密使用不同的密钥，分为公钥 public 和私钥 private，使用其中一个加密，只能用另一个解密。

|      | Desc                                    | Prod                    |
| ---- | --------------------------------------- | ----------------------- |
| RSA  | Ron Rivest, Adi Shamir, Leonard Adleman | :ballot_box_with_check: |
| DSA  | Digital Signature Algorithm             | :ballot_box_with_check: |
| ECC  | Elliptic Curves Cryptography            | :ballot_box_with_check: |

:smile: 中间人无法获取接收方的私钥，所以无法解密，**保证消息的安全性**。

:cry: 阻止不了中间人对信息进行篡改。

#### Digest

通过将消息进行哈希运算得到长度固定的唯一值。位数越长，越安全，开销越大。

:smile: 若中间人进行篡改，Digest 必然会发生改变。接收方可通过校验 Digest 判断数据是否被篡改，**保证消息的完整性**。

:cry: 即便如此，接受方也无法确认发送方的真实性，有可能被中间人伪装。

|      | Desc                  | Prod                    |
| ---- | --------------------- | ----------------------- |
| MD5  | Message Digest 5      | ❌                       |
| SHA  | Secure Hash Algorithm | :ballot_box_with_check: |

#### Digital Certificate

:smile: 身份认证，由 CA 颁发，**保证消息的真实性**。

通常证书格式 X.509

- `Version Number` 证书版本号
- `Serial Number` 证书序列号，由 CA 分配
- `Signature Algorithm ID` 数字签名算法的标识
- `Issuer` 证书颁发者
- `Validity Period` 证书的有效期
- `Subject` 证书持有者
- `Subject Public Key` 证书持有者的公钥信息
- (Optional) `Issuer Unique Identifier` 颁发者的唯一标识
- (Optional) `Subject Unique Identifier` 证书持有者唯一标识
- (Optional) `Extensions` 扩展：密钥用途，证书持有者别名
- `Signature` **数字签名：通过哈希生成 Digest，再用 CA 的私钥进行加密，附在证书的最后。**



!["A sample certificate layout"](https://azure.github.io/IoTTrainingPack/modules/Certificates101/media/Certificates_6.png)

**信任链 Chain of Trust**

- 根证书 Root Certificate 是 CA 的自签证书

- 中间证书 Intermetidate Certificates 是由根证书签发的一系列证书，这些证书可以用来签发其他中间证书 or 签发终端实体证书

- 实体（网站/应用）证书由中间证书签发。

- 如果信任了根证书，那么链上所有的中间证书都会被认证。

  

![A sample certificate chain of trust](https://azure.github.io/IoTTrainingPack/modules/Certificates101/media/Certificates_7.jpeg)



**流程 Flow**

1. 向 CA 申请证书
   - **申请人**生成一对公钥和私钥
   - **申请人**向 CA 提交 CSR，包括公钥 & 身份信息
2. CA 验证
   - 验证通过，返回 X.509 证书，其中包含了**申请人**的公钥，身份信息以及 CA 签名
3. 加密数据
   - 通信双方交换证书
   - 使用 CA 公钥验证 CA 签名并相互认证身份
   - **发送方**使用**接收方**的公钥加密数据
4. 解密数据
   - **接收方**使用自己的私钥解密数据



**KeyUsage** 指定证书公钥的使用方式，即证书公钥可以用于哪些目的：

- `KeyUsageDigitalSignature` 可用于数字签名，验证数据的完整性和身份认证
- `KeyUsageContentCommitment` 可用于内容承诺，通常用于签署内容摘要，以确保内容的完整性和不可否认性
- `KeyUsageKeyEncipherment ` 可用于密钥加密，即用于加密会话密钥等数据
- `KeyUsageDataEncipherment` 可用于数据加密，即用于对数据进行加密
- `KeyUsageKeyAgreement ` 可用于密钥协商
- `KeyUsageCertSign` 可用于签发其他证书
- `KeyUsageCRLSign ` 可以用于对证书作废列表 (CRL) 进行签名
- `KeyUsageEncipherOnly` 可用于数据加密，但不能用于数字签名。
- `KeyUsageDecipherOnly ` 可用于数据解密，但不能用于数字签名。

```go
type KeyUsage int

const (
	KeyUsageDigitalSignature KeyUsage = 1 << iota
	KeyUsageContentCommitment
	KeyUsageKeyEncipherment
	KeyUsageDataEncipherment
	KeyUsageKeyAgreement
	KeyUsageCertSign
	KeyUsageCRLSign
	KeyUsageEncipherOnly
	KeyUsageDecipherOnly
)
```

**ExtKeyUsage** 指定证书公钥的**扩展**使用方式

- `ExtKeyUsageAny` 该证书可以用于任何目的
- `ExtKeyUsageServerAuth` 可用于服务器认证，用于验证服务器的身份
- `ExtKeyUsageClientAuth` 可用于客户端认证，用于验证客户端的身份
- `ExtKeyUsageCodeSigning` 可用于代码签名
- `ExtKeyUsageEmailProtection `可以用于电子邮件保护
- `ExtKeyUsageIPSECEndSystem `可用于 IPsec 终端系统
- `ExtKeyUsageIPSECTunnel `可用于 IPsec 隧道
- `ExtKeyUsageIPSECUser `可用于 IPsec 用户
- `ExtKeyUsageTimeStamping` 可用于对数据和文件进行时间戳签名
- `ExtKeyUsageOCSPSigning `可用于在线证书状态协议 (OCSP) 签名
- `ExtKeyUsageMicrosoftServerGatedCrypto` 可用于 Microsoft 服务器门控加密
- `ExtKeyUsageNetscapeServerGatedCrypto` 可用于 Netscape 服务器门控加密
- `ExtKeyUsageMicrosoftCommercialCodeSigning` 可用于 Microsoft 商业代码签名
- `ExtKeyUsageMicrosoftKernelCodeSigning` 可用于 Microsoft 内核代码签名

```go
type ExtKeyUsage int

const (
	ExtKeyUsageAny ExtKeyUsage = iota
	ExtKeyUsageServerAuth
	ExtKeyUsageClientAuth
	ExtKeyUsageCodeSigning
	ExtKeyUsageEmailProtection
	ExtKeyUsageIPSECEndSystem
	ExtKeyUsageIPSECTunnel
	ExtKeyUsageIPSECUser
	ExtKeyUsageTimeStamping
	ExtKeyUsageOCSPSigning
	ExtKeyUsageMicrosoftServerGatedCrypto
	ExtKeyUsageNetscapeServerGatedCrypto
	ExtKeyUsageMicrosoftCommercialCodeSigning
	ExtKeyUsageMicrosoftKernelCodeSigning
)
```

