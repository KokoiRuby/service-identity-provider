package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	err := httpNoTLS()
	err = httpsTLS()
	if err != nil {
		return
	}
	err = httpsMTLS()
	if err != nil {
		return
	}
}

func httpNoTLS() error {

	//client := &http.Client{
	//	Transport: &http.Transport{
	//		TLSClientConfig: &tls.Config{
	//			InsecureSkipVerify: true,
	//		},
	//	},
	//}

	r, err := http.Get("http://localhost:8080/")
	if err != nil {
		return err
	}

	body, err := io.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Response: %s\n", string(body))
	return nil
}

func httpsTLS() error {
	// ca to verify server cert
	caCert, err := os.ReadFile("./certs/ca.pem")
	if err != nil {
		return err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// create https client & supply ca pool & client key pair
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
				//CipherSuites: []uint16{
				//	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				//	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				//	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				//	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				//	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				//	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				//},
			},
		},
	}

	resp, err := client.Get("https://localhost:8088/tls")
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Response: %s\n", string(body))
	return nil
}

func httpsMTLS() error {
	// ca to verify server cert
	caCert, err := os.ReadFile("./certs/ca.pem")
	if err != nil {
		return err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// client key pair
	keyPair, err := tls.LoadX509KeyPair("./certs/client-chain.pem", "./certs/client-key.pem")
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	// create https client & supply ca pool & client key pair
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
				Certificates: []tls.Certificate{
					keyPair,
				},
				//CipherSuites: []uint16{
				//	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				//	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				//	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				//	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				//	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				//	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				//},
			},
		},
	}

	resp, err := client.Get("https://localhost:8089/mtls")
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Response: %s\n", string(body))
	return nil
}
