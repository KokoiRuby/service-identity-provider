package main

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"log"
	"os"

	"github.com/hashicorp/vault-client-go/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	ctx     context.Context
	clientV *vault.Client
	clientK *kubernetes.Clientset
)

func init() {
	ctx = context.Background()

	// setup client to vault
	log.Println("Setup client to vault.")
	var err error
	clientV, err = vault.New(
		vault.WithAddress("http://127.0.0.1:8200"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// setup client to k8s
	//kubeConf, err := rest.InClusterConfig()
	kubeConf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}
	clientK = kubernetes.NewForConfigOrDie(kubeConf)

}

func main() {
	// vault status
	status, err := clientV.System.ReadHealthStatus(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if status.Data["initialized"].(bool) {
		log.Println("Vault is already initialized.")
		if status.Data["sealed"].(bool) {
			log.Println("Vault is sealed. Unsealing...")
			unSeal()
			log.Println("Vault is now unsealed.")
			os.Exit(0)
		} else {
			log.Println("Vault is already unsealed.")
			os.Exit(0)
		}
	} else {
		log.Println("Vault is not initialized yet. Initializing...")
		initVault()
		log.Println("Vault is now unsealed.")
		os.Exit(0)
	}
}

func getConfig() (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
}

func initVault() {
	// vault operator init -key-shares=1 -key-threshold=1
	initResp, err := clientV.System.Initialize(
		ctx,
		schema.InitializeRequest{
			SecretShares:    1,
			SecretThreshold: 1,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Vault is now initialized.")

	// persist unseal key & root token into secret
	// read from resp
	unsealKey := initResp.Data["keys_base64"].([]interface{})[0]
	rootToken := initResp.Data["root_token"]

	// in-cluster
	nsByte, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns := string(nsByte)

	// test
	ns = "sip"

	log.Println("Create secret for Vault unseal key")
	wantedSecretForUnsealKey := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-unseal-key",
		},
		StringData: map[string]string{
			"key": unsealKey.(string),
		},
		Type: corev1.SecretTypeOpaque,
	}

	createdSecretForUnsealKey, err := clientK.CoreV1().Secrets(ns).Create(ctx, wantedSecretForUnsealKey, metav1.CreateOptions{})
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			log.Fatalf("Namespace %v not found\n", ns)
		case errors.IsAlreadyExists(err):
			log.Println("Secret vault-unseal-key already exists. Updating...")
			createdSecretForUnsealKey.Data["key"] = []byte(unsealKey.(string))
			_, err = clientK.CoreV1().Secrets(ns).Update(context.TODO(), createdSecretForUnsealKey, metav1.UpdateOptions{})
			if err != nil {
				log.Fatal(err)
			}
		case errors.IsInvalid(err):
			log.Fatal("Secret spec is invalid\n")
		default:
			log.Fatal(err)
		}
	}

	log.Println("Create secret for Vault root token")
	wantedSecretForRootToken := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault-root-token",
			Namespace: ns,
		},
		StringData: map[string]string{
			"token": rootToken.(string),
		},
		Type: corev1.SecretTypeOpaque,
	}

	createdSecretForRootToken, err := clientK.CoreV1().Secrets(ns).Create(ctx, wantedSecretForRootToken, metav1.CreateOptions{})
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			log.Fatalf("Namespace %v not found\n", ns)
		case errors.IsAlreadyExists(err):
			log.Println("Secret vault-root-token already exists.")
			createdSecretForRootToken.Data["key"] = []byte(rootToken.(string))
			_, err = clientK.CoreV1().Secrets(ns).Update(context.TODO(), createdSecretForRootToken, metav1.UpdateOptions{})
			if err != nil {
				log.Fatal(err)
			}
		case errors.IsInvalid(err):
			log.Fatal("Secret spec is invalid\n")
		default:
			log.Fatal(err)
		}
	}

	_ = createdSecretForRootToken

	// vault operator unseal
	log.Println("Vault is sealed. Unsealing...")
	_, err = clientV.System.Unseal(ctx,
		schema.UnsealRequest{
			Key: unsealKey.(string),
		})
	if err != nil {
		log.Fatal(err)
	}
}

func unSeal() {
	// in-cluster
	nsByte, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns := string(nsByte)

	// test
	ns = "sip"

	// get unseal key secret
	secretOfUnsealKey, err := clientK.CoreV1().Secrets(ns).Get(ctx, "vault-unseal-key", metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error getting secret: %v", err)
	}
	unsealKey := string(secretOfUnsealKey.Data["key"])

	// vault operator unseal
	_, err = clientV.System.Unseal(ctx,
		schema.UnsealRequest{
			Key: unsealKey,
		})
	if err != nil {
		log.Fatal(err)
	}
}
