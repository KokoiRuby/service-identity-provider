package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"net"
	"time"
)

func genClientCert() {
	clientPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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
			Organization:       []string{"client.com"},
			OrganizationalUnit: []string{"test"},
			CommonName:         "test.client.com",
		},
		// SAN
		DNSNames:    []string{"test.client.com", "localhost", "service-consumer"},
		IPAddresses: []net.IP{net.IPv4(192, 168, 60, 20), net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IsCA:        false,
	}

	// optional
	// gen csr
	clientCSR, err := x509.CreateCertificateRequest(
		rand.Reader,
		&x509.CertificateRequest{
			Subject:   template.Subject,
			PublicKey: clientPriv.Public(),
		},
		clientPriv)
	if err != nil {
		log.Fatalf("Failed to create CSR: %s", err)
	}

	// write csr to file
	err = csrToFile(clientCSR, "./artifacts/client.csr")
	if err != nil {
		log.Fatalf("Failed to write csr file: %s", err)
	}

	// parse interm ca cert
	caCert, err := parsePemCert("./artifacts/intermediate.pem")
	if err != nil {
		log.Fatalf("failed to parse intermediate.pem file: %s", err)
	}

	// parse interm ca key
	caKey, err := parseECKey("./artifacts/intermediate-key.pem")
	if err != nil {
		log.Fatalf("failed to parse intermediate-key.pem file: %s", err)
	}

	// sign by ca
	clientCert, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		caCert,
		clientPriv.Public(),
		caKey)
	if err != nil {
		panic(err)
	}

	// write cert to file
	err = certToFile(clientCert, "./artifacts/client.pem")
	if err != nil {
		log.Fatalf("failed to write cert file: %s", err)
	}

	// write key to file
	err = KeyToFile(clientPriv, "./artifacts/client-key.pem")
	if err != nil {
		log.Fatalf("failed to write key file: %s", err)
	}
}
