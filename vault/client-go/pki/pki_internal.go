package main

//
//import (
//	"bytes"
//	"context"
//	"crypto/tls"
//	"fmt"
//	"io"
//	"log"
//	"net/http"
//
//	"github.com/hashicorp/vault-client-go"
//	"github.com/hashicorp/vault-client-go/schema"
//)
//
//// 任何以 "服务提供方" 或 "服务消费者" 为前缀的工件（artifact）都是特定于该微服务的，并且不会与其他微服务共享。受信任的根服务器 CA 证书由服务消费者共享。同样，中间服务器 CA 证书由服务提供方共享（尽管它是服务特定证书链的一部分）。
//
//// 左侧的工件用于服务器身份验证场景，而右侧的工件用于客户端身份验证场景。通过结合两种场景，可以实现双向认证（mutual authentication）。
//
//// 在某些特殊情况下，服务消费者可能需要从受信任的根服务器 CA 申请客户端证书，或服务提供方从服务消费者服务器 CA 申请服务器证书。用例并不是绝对分开的，它们可以根据服务的具体需求进行混合。这可以通过使用内部证书 API 实现。为了更好地理解，可以参考图 3 中的关系图，并将图中每个 "提供方" 替换为 "消费者"，每个 "服务器" 替换为 "客户端"。
//
//var (
//	ctx = context.Background()
//	//
//	//	// parsed from InternalClientCA CR
//	cn           = "service-provider"
//	clientCAPath = "sip-client-ca/" + cn
//	rootCAPath   = "sip-root-ca"
//	intermCAPath = "sip-interm-ca"
//	rootToken    = "hvs.nEibi98ZxL4Low38AEvNDXNg"
//	//
//	client *vault.Client
//	//
//	patchBody = []byte(`{"client_flag":false,"server_flag":false}`)
//	//	//patchBody = []byte(`{"data":{"client_flag":false,"server_flag":false}}`)
//)
//
//func main() {
//	// 0. Setup client to Vault
//	tls := vault.TLSConfiguration{}
//	tls.ServerCertificate.FromFile = "../../tls/vault-cert.pem"
//
//	client, err := vault.New(
//		vault.WithAddress("https://127.0.0.1:8200"),
//		vault.WithTLS(tls),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 0. AuthN against a root token (insecure)
//	if err := client.SetToken(rootToken); err != nil {
//		log.Fatal(err)
//	}
//
//	// managed by operator
//	//clientCA(client)
//
//	// managed by init
//	serverRootCA(client)
//	serverIntermCA(client)
//
//	// clean up
//	cleanup(client)
//}
//
//func clientCA(client *vault.Client) {
//	// 1. Enable PKI secrets engine.
//	_, err := client.System.MountsEnableSecretsEngine(ctx, clientCAPath,
//		schema.MountsEnableSecretsEngineRequest{
//			Type:        "pki",
//			Description: "CA certificate backend created by sip for client authn",
//		})
//	if err != nil {
//		if !vault.IsErrorStatus(err, http.StatusBadRequest) {
//			log.Fatal(err)
//		}
//		log.Println("path is already in use at sip-client-ca/service-provider-ca/")
//	}
//
//	// 2. Set TTL.
//	maxTTL := "87600h"
//	_, err = client.System.MountsTuneConfigurationParameters(ctx, clientCAPath,
//		schema.MountsTuneConfigurationParametersRequest{
//			MaxLeaseTtl: maxTTL,
//		})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 3. Configure CA keypair.
//	_, err = client.Secrets.PkiGenerateRoot(ctx, "internal",
//		schema.PkiGenerateRootRequest{
//			CommonName: "sip Internal Root CA",
//			KeyType:    "ec",
//			KeyBits:    256,
//			Ttl:        "87600h",
//		},
//		vault.WithMountPath(clientCAPath))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 4. Update CRL location & issuing certificates, can be updated in the future.
//	_, err = client.Secrets.PkiConfigureUrls(ctx,
//		schema.PkiConfigureUrlsRequest{
//			IssuingCertificates:   []string{"http://127.0.0.1:8200/v1/" + clientCAPath + "/ca"},
//			CrlDistributionPoints: []string{"http://127.0.0.1:8200/v1/" + clientCAPath + "/crl"},
//		},
//		vault.WithMountPath(clientCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	// 5. Configure a role that maps a name in Vault to a procedure for generating a certificate.
//	_, err = client.Secrets.PkiWriteRole(ctx, "client-ca",
//		schema.PkiWriteRoleRequest{
//			AllowAnyName:     true,
//			KeyType:          "ec",
//			KeyBits:          256,
//			KeyUsage:         []string{"DigitalSignature"},
//			ServerFlag:       false,
//			ClientFlag:       false,
//			ExtKeyUsage:      []string{"ClientAuth"},
//			EnforceHostnames: false,
//			MaxTtl:           "168h",
//		},
//		vault.WithMountPath(clientCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	// 6. Issue certificates
//	resp, err := client.Secrets.PkiIssueWithRole(ctx, "client-ca",
//		schema.PkiIssueWithRoleRequest{
//			CommonName: "service-provider",
//		},
//		vault.WithMountPath(clientCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	log.Println(resp.Data.Certificate)
//	log.Println(resp.Data.PrivateKey)
//
//}
//
//func serverRootCA(client *vault.Client) {
//	//listResp, err := client.System.MountsListSecretsEngines(ctx)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//log.Println("List resp:", listResp.Data)
//	//if _, ok := listResp.Data["sip-root-ca/"]; !ok {
//	//	log.Println("Secrets engine sip-root-ca/ not found.")
//	//}
//	//log.Println("Secrets engine sip-root-ca/ is found.")
//
//	// 1. Enable PKI secrets engine.
//	log.Println("1. Enable PKI secrets engine.")
//	_, err = client.System.MountsEnableSecretsEngine(ctx, rootCAPath,
//		schema.MountsEnableSecretsEngineRequest{
//			Type:        "pki",
//			Description: "CA certificate backend created by sip for server authn",
//		})
//	if err != nil {
//		if !vault.IsErrorStatus(err, http.StatusBadRequest) {
//			log.Fatal(err)
//		}
//		log.Println("path is already in use at sip-root-ca/")
//	}
//
//	// 2. Set TTL.
//	log.Println("2. Enable PKI secrets engine.")
//	maxTTL := "87600h"
//	_, err = client.System.MountsTuneConfigurationParameters(ctx, rootCAPath,
//		schema.MountsTuneConfigurationParametersRequest{
//			MaxLeaseTtl: maxTTL,
//		})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 3. Configure CA keypair.
//	log.Println("3. Configure CA keypair.")
//	_, err = client.Secrets.PkiGenerateRoot(ctx, "internal",
//		schema.PkiGenerateRootRequest{
//			CommonName: "sip Internal Root CA",
//			KeyType:    "ec",
//			KeyBits:    256,
//			Ttl:        "87600h",
//		},
//		vault.WithMountPath(rootCAPath))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 4. Update CRL location & issuing certificates, can be updated in the future.
//	log.Println("4. Update CRL location & issuing certificates, can be updated in the future.")
//	_, err = client.Secrets.PkiConfigureUrls(ctx,
//		schema.PkiConfigureUrlsRequest{
//			IssuingCertificates:   []string{"http://127.0.0.1:8200/v1/" + rootCAPath + "/ca"},
//			CrlDistributionPoints: []string{"http://127.0.0.1:8200/v1/" + rootCAPath + "/crl"},
//		},
//		vault.WithMountPath(rootCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	// 5. Configure a role that maps a name in Vault to a procedure for generating a certificate.
//	log.Println("5. Configure a role that maps a name in Vault to a procedure for generating a certificate.")
//	rootCARole := "root-ca"
//	_, err = client.Secrets.PkiWriteRole(ctx, rootCARole,
//		schema.PkiWriteRoleRequest{
//			AllowAnyName: true,
//			KeyType:      "ec",
//			KeyBits:      256,
//			KeyUsage:     []string{"CertSign,CRLSign"},
//			// not effective, switching to HTTP
//			//ServerFlag:       false,
//			//ClientFlag:       false,
//			EnforceHostnames: false,
//			MaxTtl:           "876000h",
//		},
//		vault.WithMountPath(rootCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	// 5. set server/client flag to false by HTTP
//	disableFlag(rootCAPath, rootCARole)
//
//	//roleResp, err := client.Secrets.PkiReadRole(ctx, "root-ca", vault.WithMountPath(rootCAPath))
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//fmt.Println(roleResp.Data.ServerFlag)
//	//fmt.Println(roleResp.Data.ClientFlag)
//	//fmt.Println(roleResp.Data.KeyUsage)
//	//fmt.Println(roleResp.Data.ExtKeyUsage)
//
//	// 6. Issue certificates
//	log.Println("Issue certificates.")
//	_, err = client.Secrets.PkiIssueWithRole(ctx, "root-ca",
//		schema.PkiIssueWithRoleRequest{
//			CommonName: "service-provider",
//			//UriSans:    []string{"service-provider", "service-provider.sip", "service-provider.sip.svc", "service-provider.sip.svc.cluster.local"},
//			AltNames: "service-provider, service-provider.sip, service-provider.sip.svc, service-provider.sip.svc.cluster.local",
//		},
//		vault.WithMountPath(rootCAPath))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	//log.Println(issueResp.Data.CaChain)
//	//log.Println(issueResp.Data.CaChain[0])
//	//log.Println(issueResp.Data.CaChain[1])
//	//log.Println(issueResp.Data.CaChain[0] + issueResp.Data.CaChain[1])
//	//log.Println(issueResp.Data.Certificate)
//	//log.Println(issueResp.Data.PrivateKey)
//
//}
//
//func disableFlag(pki, role string) {
//	req, err := http.NewRequest("PATCH", "https://127.0.0.1:8200/v1/"+pki+"/roles/"+role, bytes.NewBuffer(patchBody))
//	if err != nil {
//		fmt.Println("Error creating request:", err)
//		return
//	}
//
//	req.Header.Set("Content-Type", "application/merge-patch+json")
//	req.Header.Set("X-Vault-Token", rootToken)
//
//	clientH := &http.Client{
//		Transport: &http.Transport{
//			TLSClientConfig: &tls.Config{
//				InsecureSkipVerify: true,
//			},
//		},
//	}
//	resp, err := clientH.Do(req)
//	if err != nil {
//		fmt.Println("Error sending request:", err)
//		return
//	}
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//
//		}
//	}(resp.Body)
//}
//
//func serverIntermCA(client *vault.Client) {
//	// 1. Enable PKI secrets engine.
//	_, err := client.System.MountsEnableSecretsEngine(ctx, intermCAPath,
//		schema.MountsEnableSecretsEngineRequest{
//			Type:        "pki",
//			Description: "CA certificate backend created by sip for server authn",
//		})
//	if err != nil {
//		if !vault.IsErrorStatus(err, http.StatusBadRequest) {
//			log.Fatal(err)
//		}
//		log.Println("path is already in use at sip-interm-ca")
//	}
//
//	// 2. Set TTL.
//	maxTTL := "168h"
//	_, err = client.System.MountsTuneConfigurationParameters(ctx, intermCAPath,
//		schema.MountsTuneConfigurationParametersRequest{
//			MaxLeaseTtl: maxTTL,
//		})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 3. Configure CA keypair. No need for Intermediate CA.
//
//	// 4. Update CRL location & issuing certificates, can be updated in the future.
//	_, err = client.Secrets.PkiConfigureUrls(ctx,
//		schema.PkiConfigureUrlsRequest{
//			IssuingCertificates:   []string{"http://127.0.0.1:8200/v1/" + intermCAPath + "/ca"},
//			CrlDistributionPoints: []string{"http://127.0.0.1:8200/v1/" + intermCAPath + "/crl"},
//		},
//		vault.WithMountPath(intermCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	// 5. Configure a role that maps a name in Vault to a procedure for generating a certificate.
//	intermCARole := "interm-ca"
//	_, err = client.Secrets.PkiWriteRole(ctx, intermCARole,
//		schema.PkiWriteRoleRequest{
//			AllowAnyName: true,
//			KeyType:      "ec",
//			KeyBits:      256,
//			KeyUsage:     []string{"DigitalSignature"},
//			//ServerFlag:       false,
//			//ClientFlag:       false,
//			ExtKeyUsage:      []string{"ServerAuth"},
//			EnforceHostnames: false,
//			MaxTtl:           "168h",
//		},
//		vault.WithMountPath(intermCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	// 5. set server/client flag to false by HTTP
//	disableFlag(intermCAPath, intermCARole)
//
//	roleResp, err := client.Secrets.PkiReadRole(ctx, "interm-ca", vault.WithMountPath(intermCAPath))
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(roleResp.Data.ServerFlag)
//	fmt.Println(roleResp.Data.ClientFlag)
//	fmt.Println(roleResp.Data.KeyUsage)
//	fmt.Println(roleResp.Data.ExtKeyUsage)
//
//	// 6. Generate intermediate CA CSR.
//	csrResp, err := client.Secrets.PkiGenerateIntermediate(
//		context.Background(),
//		"internal",
//		schema.PkiGenerateIntermediateRequest{
//			CommonName:          "sip Internal Intermediate CA",
//			KeyType:             "ec",
//			KeyBits:             256,
//			AddBasicConstraints: true,
//			Ttl:                 "168h",
//		},
//		vault.WithMountPath(intermCAPath),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	//log.Println(csrResp.Data.Csr)
//
//	// 7. Signed by Root CA.
//	signResp, err := client.Secrets.PkiRootSignIntermediate(
//		context.Background(),
//		schema.PkiRootSignIntermediateRequest{
//			Csr:          csrResp.Data.Csr,
//			Format:       "pem",
//			UseCsrValues: true,
//			Ttl:          "876000h",
//		},
//		vault.WithMountPath(rootCAPath),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// log.Println(signResp.Data.CaChain)
//	// log.Println(signResp.Data.CaChain[0])
//	// log.Println(signResp.Data.CaChain[1])
//	// log.Println(signResp.Data.CaChain[0] + signResp.Data.CaChain[1])
//	log.Println(signResp.Data.Certificate)
//
//	// 8. Set the intermediate CA signing certificate to the root-signed certificate.
//	_, err = client.Secrets.PkiSetSignedIntermediate(
//		context.Background(),
//		schema.PkiSetSignedIntermediateRequest{
//			// Certificate: signResp.Data.CaChain[0] + signResp.Data.CaChain[1],
//			Certificate: signResp.Data.Certificate,
//		},
//		vault.WithMountPath(intermCAPath),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 9. Issue certificates
//	issueResp, err := client.Secrets.PkiIssueWithRole(ctx, "interm-ca",
//		schema.PkiIssueWithRoleRequest{
//			CommonName: "service-provider",
//			AltNames:   "service-provider, service-provider.sip, service-provider.sip.svc, service-provider.sip.svc.cluster.local",
//		},
//		vault.WithMountPath(intermCAPath))
//	if err != nil {
//		log.Fatal()
//	}
//
//	//log.Println(signResp.Data.CaChain)
//	//log.Println(signResp.Data.CaChain[0])
//	//log.Println(signResp.Data.CaChain[1])
//	//log.Println(signResp.Data.CaChain[0] + signResp.Data.CaChain[1])
//	log.Println(issueResp.Data.Certificate)
//	//log.Println(issueResp.Data.PrivateKey)
//}
//
//func cleanup(client *vault.Client) {
//	//_, err := client.System.AuthDisableMethod(
//	//	ctx,
//	//	clientCAPath,
//	//)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	_, err := client.System.AuthDisableMethod(
//		ctx,
//		rootCAPath,
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	_, err = client.System.AuthDisableMethod(
//		ctx,
//		intermCAPath,
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
