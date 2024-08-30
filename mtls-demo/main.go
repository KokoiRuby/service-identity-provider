package main

func main() {
	genCA()

	genIntermCSRandCert()

	genServerCert()
	err := buildServerCertChain("./artifacts/server.pem", "./artifacts/intermediate.pem")
	if err != nil {
		return
	}

	genClientCert()
	err = buildClientCertChain("./artifacts/client.pem", "./artifacts/intermediate.pem")
	if err != nil {
		return
	}

}
