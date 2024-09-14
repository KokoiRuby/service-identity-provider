package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return
	}

	// no TLS
	go func() {
		r1 := gin.Default()
		err := r1.SetTrustedProxies([]string{"127.0.0.1"})
		if err != nil {
			return
		}

		r1.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "no tls")
		})

		// http://localhost:8080/
		err = r1.Run(":8080")
		if err != nil {
			return
		}
	}()

	// TLS
	go func() {
		r2 := gin.Default()
		err := r2.SetTrustedProxies([]string{"127.0.0.1"})
		if err != nil {
			return
		}

		r2.GET("/tls", func(c *gin.Context) {
			c.String(http.StatusOK, "tls")
		})

		// https://localhost:8088/tls
		err = r2.RunTLS(":8088", "./keypair/tls.crt", "./keypair/tls.key")
		if err != nil {
			return
		}

	}()

	// mTLS
	r.GET("/mtls", func(c *gin.Context) {
		c.String(http.StatusOK, "mtls")
	})

	caCert, err := os.ReadFile("./ca/client-ca.pem")
	if err != nil {
		log.Fatalf("failed to read CA cert: %s", err)
	}

	// add ca cert to pool
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// tls conf
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert, // need to verify client cert
		MinVersion: tls.VersionTLS12,
	}

	// load server cert
	// if chained = interm cert + server cert
	// tls: private key does not match public key
	// but if u only keep server cert
	// tls: failed to verify certificate: x509: certificate signed by unknown authority
	// Solution: have to changed the order of interm cert & server cert, must be: server cert + interm cert
	cert, err := tls.LoadX509KeyPair("./keypair/tls.crt", "./keypair/tls.key")
	if err != nil {
		log.Fatalf("failed to load server key pair: %s", err)
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	server := &http.Server{
		Addr:      ":8089",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	// https://localhost:8089/mtls
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}
