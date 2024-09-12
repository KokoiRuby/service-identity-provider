package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net"
	"os"
	"time"
)

const (
	VAULT_INTERNAL_SERVICE_NAME = "vault-internal"
	VAULT_ACTIVE_SERVICE_NAME   = "vault-active"
)

var (
	ctx       context.Context
	clientset *kubernetes.Clientset
	// ENV
	ns, clusterName string
)

func init() {
	ctx = context.Background()

	// ENV
	ns = os.Getenv("VAULT_K8S_NAMESPACE")
	clusterName = os.Getenv("K8S_CLUSTER_NAME")
	//ns = "sip"
	//clusterName = "cluster.local"

	// in-cluster kubeconfig
	//config, err := getConfig()
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Error creating Kubernetes config: %v\n", err)
		os.Exit(1)
	}

	// build clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	// gen private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Fatalf("failed to marshal private key: %s", err)
	}
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	// gen csr
	csr, err := x509.CreateCertificateRequest(
		rand.Reader,
		&x509.CertificateRequest{
			Subject: pkix.Name{
				Organization: []string{"system:nodes"},
				CommonName:   "system:node:" + VAULT_INTERNAL_SERVICE_NAME + "." + ns + ".svc",
			},
			SignatureAlgorithm: x509.ECDSAWithSHA256,
			DNSNames: []string{
				VAULT_ACTIVE_SERVICE_NAME,
				"*." + VAULT_INTERNAL_SERVICE_NAME,
				"*." + VAULT_INTERNAL_SERVICE_NAME + "." + ns,
				"*." + VAULT_INTERNAL_SERVICE_NAME + "." + ns + ".svc.",
				"*." + VAULT_INTERNAL_SERVICE_NAME + "." + ns + ".svc." + clusterName,
			},
			IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		},
		priv)
	if err != nil {
		log.Fatalf("Failed to create CSR: %s", err)
	}

	// 1 year
	// not greater than --cluster-signing-duration of controller-manager
	expirationSeconds := int32(31557600)
	csrToK8s := &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault.svc",
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			SignerName:        "kubernetes.io/kubelet-serving",
			ExpirationSeconds: &expirationSeconds,
			Request: pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE REQUEST",
				Bytes: csr,
			}),
			Usages: []certificatesv1.KeyUsage{
				certificatesv1.UsageDigitalSignature,
				certificatesv1.UsageKeyEncipherment,
				certificatesv1.UsageServerAuth,
			},
		},
	}

	// issue
	_, err = clientset.CertificatesV1().CertificateSigningRequests().Create(ctx, csrToK8s, metav1.CreateOptions{})
	if err != nil {
		switch {
		case errors.IsAlreadyExists(err):
			log.Println("CSR Already exists")
			err = clientset.CertificatesV1().CertificateSigningRequests().Delete(ctx, "vault.svc", metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Error deleting CertificateSigningRequest: %v\n", err)
				os.Exit(1)
			}
			_, err = clientset.CertificatesV1().CertificateSigningRequests().Create(ctx, csrToK8s, metav1.CreateOptions{})
			if err != nil {
				log.Printf("Error creating CertificateSigningRequest: %v\n", err)
			}
		default:
			log.Fatalf("Error creating CertificateSigningRequest: %v\n", err)
		}
	}

	// get csr to be approved
	csrToBeApproved, err := clientset.CertificatesV1().CertificateSigningRequests().Get(ctx, "vault.svc", metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting CertificateSigningRequest: %v\n", err)
		os.Exit(1)
	}

	// approve csr
	csrToBeApproved.Status.Conditions = append(csrToBeApproved.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
		Type:           certificatesv1.CertificateApproved,
		Reason:         "ManuallyApproved",
		Message:        "This CSR was approved manually",
		LastUpdateTime: metav1.Now(),
		Status:         corev1.ConditionTrue,
	})
	_, err = clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(ctx, "vault.svc", csrToBeApproved, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Error approving CertificateSigningRequest: %v\n", err)
		os.Exit(1)
	}

	// wait for cert gen
	time.Sleep(1 * time.Second)

	// parse approved csr
	csrApproved, err := clientset.CertificatesV1().CertificateSigningRequests().Get(ctx, "vault.svc", metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting CertificateSigningRequest: %v\n", err)
		os.Exit(1)
	}

	//persist keypair to secret
	wantedSecretForVaultTLS := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault-ha-tls",
			Namespace: ns,
		},
		Data: map[string][]byte{
			"vault.crt": csrApproved.Status.Certificate,
			"vault.key": privPem,
		},
		Type: corev1.SecretTypeOpaque,
	}

	_, err = clientset.CoreV1().Secrets(ns).Create(ctx, wantedSecretForVaultTLS, metav1.CreateOptions{})
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			log.Fatalf("Namespace %v not found\n", ns)
		case errors.IsAlreadyExists(err):
			log.Println("Secret vault-ha-tls already exists. Updating...")
			existedSecretForVaultTLS, _ := clientset.CoreV1().Secrets(ns).Get(ctx, "vault-ha-tls", metav1.GetOptions{})
			existedSecretForVaultTLS.Data["vault.crt"] = csrApproved.Status.Certificate
			existedSecretForVaultTLS.Data["vault.key"] = privPem
			_, err = clientset.CoreV1().Secrets(ns).Update(context.TODO(), existedSecretForVaultTLS, metav1.UpdateOptions{})
			if err != nil {
				log.Fatal(err)
			}
		case errors.IsInvalid(err):
			log.Fatal("Secret spec is invalid.\n")
		default:
			log.Fatal(err)
		}
	}

	// delete approved csr
	err = clientset.CertificatesV1().CertificateSigningRequests().Delete(ctx, "vault.svc", metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error deleting CertificateSigningRequest: %v\n", err)
		os.Exit(1)
	}

}

func getConfig() (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
}
