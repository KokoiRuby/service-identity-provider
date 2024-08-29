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

func genIntermCSRandCert() {
	intermPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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
			Organization:       []string{"intermediate.com"},
			OrganizationalUnit: []string{"ca"},
			CommonName:         "ca.intermediate.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// parse ca cert
	caCert, err := parsePemCert("./artifacts/ca.pem")
	if err != nil {
		log.Fatalf("failed to parse ca.pem file: %s", err)
	}

	//var akiByte []byte
	//for _, ext := range caCert.Extensions {
	//	// AuthorityKeyIdentifier OID: 2.5.29.35
	//	if ext.Id.Equal([]int{2, 5, 29, 35}) {
	//		akiByte = ext.Value
	//		break
	//	}
	//}
	//
	//_ = akiByte

	// parse ca key
	caKey, err := parseECKey("./artifacts/ca-key.pem")
	if err != nil {
		log.Fatalf("failed to parse ca-key.pem file: %s", err)
	}

	// optional
	// gen csr
	intermediateCSR, err := x509.CreateCertificateRequest(
		rand.Reader,
		&x509.CertificateRequest{
			Subject:   template.Subject,
			PublicKey: intermPriv.Public(),
		},
		intermPriv)
	if err != nil {
		log.Fatalf("Failed to create CSR: %s", err)
	}

	// write csr to file
	err = csrToFile(intermediateCSR, "./artifacts/intermediate.csr")
	if err != nil {
		log.Fatalf("Failed to write csr file: %s", err)
	}

	// sign by ca
	cert, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		caCert,
		intermPriv.Public(),
		caKey)
	if err != nil {
		panic(err)
	}

	// write cert to file
	err = certToFile(cert, "./artifacts/intermediate.pem")
	if err != nil {
		log.Fatalf("failed to write cert file: %s", err)
	}

	// write key to file
	err = KeyToFile(intermPriv, "./artifacts/intermediate-key.pem")
	if err != nil {
		log.Fatalf("failed to write key file: %s", err)
	}

}
