package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"os"
)

var (
	ctx context.Context

	rootCAPath   = "sip-root-ca"
	intermCAPath = "sip-interm-ca"

	rootCATTL   = "876000h"
	intermCATTL = "168h"

	rootCARole   = "root-ca"
	intermCARole = "interm-ca"

	clientV *vault.Client
	clientK *kubernetes.Clientset

	patchBody = []byte(`{"client_flag":false,"server_flag":false}`)

	nsByte, _ = os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns        = string(nsByte)

	rootToken string
)

func init() {
	ctx = context.Background()

	// setup client to k8s
	kubeConf, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	clientK = kubernetes.NewForConfigOrDie(kubeConf)

	// vault root token
	existedSecretForRootToken, _ := clientK.CoreV1().Secrets(ns).Get(ctx, "vault-root-token", metav1.GetOptions{})
	rootToken = string(existedSecretForRootToken.Data["token"])

	// setup client to vault
	log.Println("Setup client to vault.")
	clientV, err = vault.New(
		vault.WithAddress("http://vault:8200"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// authn against root token
	if err := clientV.SetToken(rootToken); err != nil {
		log.Fatal(err)
	}
}

func main() {

	listResp, err := clientV.System.MountsListSecretsEngines(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Root CA
	if _, ok := listResp.Data["sip-root-ca/"]; !ok {
		log.Println("Secrets engine sip-root-ca/ not found. Enabling...")
		enablePKIRootCA()
	} else {
		log.Println("Secrets engine sip-root-ca/ is already enabled.")
	}

	// Intermediate CA
	if _, ok := listResp.Data["sip-interm-ca/"]; !ok {
		log.Println("Secrets engine sip-interm-ca/ not found. Enabling...")
		enablePKIIntermCA()
	} else {
		log.Println("Secrets engine sip-interm-ca/ is already enabled.")
	}

	log.Println("Exiting...")
	os.Exit(0)

}

func enablePKIRootCA() {
	log.Println("1. Enable sip-root-ca/ secrets engine.")
	_, _ = clientV.System.MountsEnableSecretsEngine(ctx, rootCAPath,
		schema.MountsEnableSecretsEngineRequest{
			Type:        "pki",
			Description: "CA certificate backend created by sip for server authn",
		})

	log.Println("2. Set up TTL for sip-root-ca/ secrets engine.")
	_, err := clientV.System.MountsTuneConfigurationParameters(ctx, rootCAPath,
		schema.MountsTuneConfigurationParametersRequest{
			MaxLeaseTtl: rootCATTL,
		})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("3. Configure Root CA keypair.")
	rootCAKeyPair, err := clientV.Secrets.PkiGenerateRoot(ctx, "internal",
		schema.PkiGenerateRootRequest{
			CommonName: "sip Internal Root CA",
			KeyType:    "ec",
			KeyBits:    256,
			Ttl:        rootCATTL,
		},
		vault.WithMountPath(rootCAPath))
	if err != nil {
		log.Fatal(err)
	}

	// persist root ca cert to secret
	go func() {
		log.Println("Create secret for Vault unseal key.")
		wantedSecretForUnsealKey := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "sip-root-ca",
			},
			StringData: map[string]string{
				"ca.pem": rootCAKeyPair.Data.Certificate,
			},
			Type: corev1.SecretTypeOpaque,
		}

		_, err = clientK.CoreV1().Secrets(ns).Create(ctx, wantedSecretForUnsealKey, metav1.CreateOptions{})
		if err != nil {
			switch {
			case errors.IsNotFound(err):
				log.Fatalf("Namespace %v not found\n", ns)
			case errors.IsAlreadyExists(err):
				log.Println("Secret sip-root-ca already exists. Updating...")
				existedSecretForUnsealKey, _ := clientK.CoreV1().Secrets(ns).Get(ctx, "sip-root-ca", metav1.GetOptions{})
				existedSecretForUnsealKey.Data["ca.pem"] = []byte(rootCAKeyPair.Data.Certificate)
				_, err = clientK.CoreV1().Secrets(ns).Update(context.TODO(), existedSecretForUnsealKey, metav1.UpdateOptions{})
				if err != nil {
					log.Fatal(err)
				}
			case errors.IsInvalid(err):
				log.Fatal("Secret spec is invalid.\n")
			default:
				log.Fatal(err)
			}
		}
	}()

	log.Println("4. Update CRL location & issuing certificates for sip-root-ca/ secrets engine.")
	_, err = clientV.Secrets.PkiConfigureUrls(ctx,
		schema.PkiConfigureUrlsRequest{
			IssuingCertificates:   []string{"https://vault:8200/v1/" + rootCAPath + "/ca"},
			CrlDistributionPoints: []string{"https://vault:8200/v1/" + rootCAPath + "/crl"},
		},
		vault.WithMountPath(rootCAPath))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("5. Configure a role for sip-root-ca/ secrets engine to issue certificates.")
	_, err = clientV.Secrets.PkiWriteRole(ctx, rootCARole,
		schema.PkiWriteRoleRequest{
			AllowAnyName:     true,
			KeyType:          "ec",
			KeyBits:          256,
			KeyUsage:         []string{"CertSign,CRLSign"},
			EnforceHostnames: false,
			MaxTtl:           rootCATTL,
		},
		vault.WithMountPath(rootCAPath))
	if err != nil {
		log.Fatal(err)
	}

	disableFlag(rootCAPath, rootCARole)
	//roleResp, err := clientV.Secrets.PkiReadRole(ctx, "root-ca", vault.WithMountPath(rootCAPath))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println(roleResp.Data.ServerFlag)
	//log.Println(roleResp.Data.ClientFlag)
	//log.Println(roleResp.Data.KeyUsage)
	//log.Println(roleResp.Data.ExtKeyUsage)

}

func enablePKIIntermCA() {
	log.Println("1. Enable sip-interm-ca/ secrets engine.")
	_, _ = clientV.System.MountsEnableSecretsEngine(ctx, intermCAPath,
		schema.MountsEnableSecretsEngineRequest{
			Type:        "pki",
			Description: "CA certificate backend created by sip for server authn",
		})

	log.Println("2. Set up TTL for sip-interm-ca/ secrets engine.")
	_, err := clientV.System.MountsTuneConfigurationParameters(ctx, intermCAPath,
		schema.MountsTuneConfigurationParametersRequest{
			MaxLeaseTtl: intermCATTL,
		})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("3. Update CRL location & issuing certificates for sip-interm-ca/ secrets engine.")
	_, err = clientV.Secrets.PkiConfigureUrls(ctx,
		schema.PkiConfigureUrlsRequest{
			IssuingCertificates:   []string{"https://vault:8200/v1/" + intermCAPath + "/ca"},
			CrlDistributionPoints: []string{"https://vault:8200/v1/" + intermCAPath + "/crl"},
		},
		vault.WithMountPath(intermCAPath))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("4. Configure a role for sip-interm-ca/ secrets engine to issue certificates.")
	_, err = clientV.Secrets.PkiWriteRole(ctx, intermCARole,
		schema.PkiWriteRoleRequest{
			AllowAnyName:     true,
			KeyType:          "ec",
			KeyBits:          256,
			KeyUsage:         []string{"DigitalSignature"},
			ExtKeyUsage:      []string{"ServerAuth"},
			EnforceHostnames: false,
			MaxTtl:           intermCATTL,
		},
		vault.WithMountPath(intermCAPath))
	if err != nil {
		log.Fatal()
	}

	disableFlag(intermCAPath, intermCARole)
	//roleResp, err := clientV.Secrets.PkiReadRole(ctx, "interm-ca", vault.WithMountPath(intermCAPath))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println(roleResp.Data.ServerFlag)
	//log.Println(roleResp.Data.ClientFlag)
	//log.Println(roleResp.Data.KeyUsage)
	//log.Println(roleResp.Data.ExtKeyUsage)

	log.Println("5. Generate intermediate CA CSR.")
	csrResp, err := clientV.Secrets.PkiGenerateIntermediate(
		context.Background(),
		"internal",
		schema.PkiGenerateIntermediateRequest{
			CommonName:          "sip Internal Intermediate CA",
			KeyType:             "ec",
			KeyBits:             256,
			AddBasicConstraints: true,
			Ttl:                 intermCATTL,
		},
		vault.WithMountPath(intermCAPath),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("5. Signed by Root CA.")
	signResp, err := clientV.Secrets.PkiRootSignIntermediate(
		context.Background(),
		schema.PkiRootSignIntermediateRequest{
			Csr:          csrResp.Data.Csr,
			Format:       "pem",
			UseCsrValues: true,
			Ttl:          rootCATTL,
		},
		vault.WithMountPath(rootCAPath),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("6. Set the intermediate CA signing certificate to the root-signed certificate.")
	_, err = clientV.Secrets.PkiSetSignedIntermediate(
		context.Background(),
		schema.PkiSetSignedIntermediateRequest{
			Certificate: signResp.Data.Certificate,
		},
		vault.WithMountPath(intermCAPath),
	)
	if err != nil {
		log.Fatal(err)
	}

	//issueResp, err := clientV.Secrets.PkiIssueWithRole(ctx, "interm-ca",
	//	schema.PkiIssueWithRoleRequest{
	//		CommonName: "service-provider",
	//		AltNames:   "service-provider, service-provider.sip, service-provider.sip.svc, service-provider.sip.svc.cluster.local",
	//	},
	//	vault.WithMountPath(intermCAPath))
	//if err != nil {
	//	log.Fatal()
	//}
	//
	//log.Println(issueResp.Data.CaChain[0])
	//log.Println(issueResp.Data.Certificate)
	//log.Println(issueResp.Data.PrivateKey)
}

func disableFlag(pki, role string) {
	req, err := http.NewRequest("PATCH", "http://vault:8200/v1/"+pki+"/roles/"+role, bytes.NewBuffer(patchBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/merge-patch+json")
	req.Header.Set("X-Vault-Token", rootToken)

	clientH := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := clientH.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
}
