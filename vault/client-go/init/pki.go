package main

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	tls := vault.TLSConfiguration{}
	tls.ServerCertificate.FromFile = "../tls/vault-cert.pem"

	client, err := vault.New(
		vault.WithAddress("https://localhost:8200"),
		vault.WithTLS(tls),
	)
	if err != nil {
		log.Fatal(err)
	}

	// authenticate with a root token (insecure)
	if err := client.SetToken("hvs.zlC0lWZ3ReP0UxZr5U63yXCB"); err != nil {
		log.Fatal(err)
	}

	// 1. Enable PKI secrets engine.
	path := "sip-client-ca/service-provider-ca"
	_, err = client.System.MountsEnableSecretsEngine(ctx, path,
		schema.MountsEnableSecretsEngineRequest{
			Type:        "pki",
			Description: "CA certificate backend created by sip for client authn",
		})
	if err != nil {
		if !vault.IsErrorStatus(err, http.StatusBadRequest) {
			log.Fatal(err)
		}
		log.Println("path is already in use at sip-client-ca/service-provider-ca/")
	}

	// 2. Set TTL.
	maxTTL := "87600h"
	_, err = client.System.MountsTuneConfigurationParameters(ctx, path,
		schema.MountsTuneConfigurationParametersRequest{
			MaxLeaseTtl: maxTTL,
		})

	// 3. Configure CA keypair.
	_, err = client.Secrets.PkiGenerateRoot(ctx, "internal",
		schema.PkiGenerateRootRequest{
			CommonName: "service-provider",
			KeyType:    "ec",
			KeyBits:    256,
			Ttl:        "87600h",
		},
		vault.WithMountPath(path))

	// 4. Update CRL location & issuing certificates, can be updated in the future.
	_, err = client.Secrets.PkiConfigureUrls(ctx,
		schema.PkiConfigureUrlsRequest{
			IssuingCertificates:   []string{"http://127.0.0.1:8200/v1/sip-client-ca/service-provider-ca/ca"},
			CrlDistributionPoints: []string{"http://127.0.0.1:8200/v1/sip-client-ca/service-provider-ca/crl"},
		},
		vault.WithMountPath(path))

	// 5. Configure a role that maps a name in Vault to a procedure for generating a certificate.
	_, err = client.Secrets.PkiWriteRole(ctx, "client-ca",
		schema.PkiWriteRoleRequest{
			KeyType:          "ec",
			KeyBits:          256,
			KeyUsage:         []string{"DigitalSignature"},
			ServerFlag:       false,
			ClientFlag:       false,
			ExtKeyUsage:      []string{"ClientAuth"},
			AllowedDomains:   []string{"service-provider", "cluster.local"},
			AllowSubdomains:  true,
			EnforceHostnames: false,
			MaxTtl:           "168h",
		},
		vault.WithMountPath(path))

	// 6. Issue certificates
	resp, err := client.Secrets.PkiIssueWithRole(ctx, "client-ca",
		schema.PkiIssueWithRoleRequest{
			CommonName: "service-provider.cluster.local",
		},
		vault.WithMountPath(path))

	log.Println(resp.Data.Certificate)
	log.Println(resp.Data.PrivateKey)

	// 7. Clean up

}
