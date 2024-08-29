package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"time"
)

func genCA() {
	// gen private key
	// priv, err := rsa.GenerateKey(rand.Reader, 2048)
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	// gen serial num
	serialNumber, err := generateSerialNumber()
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	// build cert template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Province:           []string{"SH"},
			Organization:       []string{"sip.com"},
			OrganizationalUnit: []string{"ca"},
			CommonName:         "ca.sip.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		AuthorityKeyId:        []byte("Private CA"),
	}

	// creates a new X.509 v3 certificate based on a template
	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	// write cert to file
	err = certToFile(cert, "./artifacts/ca.pem")
	if err != nil {
		log.Fatalf("failed to write cert file: %s", err)
	}

	// write key to file
	err = KeyToFile(priv, "./artifacts/ca-key.pem")
	if err != nil {
		log.Fatalf("failed to write key file: %s", err)
	}
}
