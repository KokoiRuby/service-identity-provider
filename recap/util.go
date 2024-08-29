package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
)

func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	return serialNumber, nil
}

func parsePemCert(file string) (*x509.Certificate, error) {
	certPem, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read cert file: %s", err)
	}

	blk, _ := pem.Decode(certPem)
	if blk == nil {
		log.Fatalf("failed to decode %s", file)
	}

	return x509.ParseCertificate(blk.Bytes)
}

func parseECKey(file string) (*ecdsa.PrivateKey, error) {
	ECKeyPem, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read key file: %s", err)
	}

	blk, _ := pem.Decode(ECKeyPem)
	if blk == nil {
		log.Fatalf("failed to decode %s", file)
	}
	return x509.ParseECPrivateKey(blk.Bytes)
}

func certToFile(cert []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create cert file: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close cert file: %s", err)
		}
	}(file)
	return pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
}

func KeyToFile(key interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create key file: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close cert file: %s", err)
		}
	}(file)

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		keyBytes, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			log.Fatalf("Unable to marshal ECDSA private key: %v", err)
		}
		return pem.Encode(file, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
		// TODO: more cases
	default:
		return fmt.Errorf("unsupported key type")
	}
}

func csrToFile(csr []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create csr file: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close csr file: %s", err)
		}
	}(file)
	return pem.Encode(file, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr})
}

func buildServerCertChain(serverCert string, intermCerts ...string) error {
	file, err := os.Create("server-chain.pem")
	if err != nil {
		log.Fatalf("failed to create cert chain file: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close cert chain file: %s", err)
		}
	}(file)

	var certChain []byte

	serverCertPEM, err := os.ReadFile(serverCert)
	if err != nil {
		log.Fatalf("failed to read server cert file: %s", err)
	}

	for _, intermCert := range intermCerts {
		intermCert, err := os.ReadFile(intermCert)
		if err != nil {
			log.Fatalf("failed to read intermediate cert file: %s", err)
		}
		certChain = append(serverCertPEM, intermCert...)
	}

	//fmt.Println(string(certChain))

	err = os.WriteFile("./artifacts/server-chain.pem", certChain, 0600)
	if err != nil {
		log.Fatalf("failed to write server cert chain file: %s", err)
	}

	return nil
}

func buildClientCertChain(clientCert string, intermCerts ...string) error {
	file, err := os.Create("server-chain.pem")
	if err != nil {
		log.Fatalf("failed to create cert chain file: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close cert chain file: %s", err)
		}
	}(file)

	var certChain []byte

	serverCertPEM, err := os.ReadFile(clientCert)
	if err != nil {
		log.Fatalf("failed to read server cert file: %s", err)
	}

	for _, intermCert := range intermCerts {
		intermCert, err := os.ReadFile(intermCert)
		if err != nil {
			log.Fatalf("failed to read intermediate cert file: %s", err)
		}
		certChain = append(serverCertPEM, intermCert...)
	}

	//fmt.Println(string(certChain))

	err = os.WriteFile("./artifacts/client-chain.pem", certChain, 0600)
	if err != nil {
		log.Fatalf("failed to write client cert chain file: %s", err)
	}

	return nil
}
