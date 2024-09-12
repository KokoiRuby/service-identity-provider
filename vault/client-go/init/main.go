package main

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const VAULT_INTERNAL_SERVICE_NAME = "vault-internal"

var (
	ctx context.Context

	clientK *kubernetes.Clientset
	clientV *vault.Client

	nsByte, _ = os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns        = string(nsByte)

	// env
	vaultFQDNs []string

	// sync
	wg sync.WaitGroup
)

func init() {
	ctx = context.Background()

	// setup client to k8s
	kubeConf, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	clientK = kubernetes.NewForConfigOrDie(kubeConf)

	// vault-0,vault-1,vault-2
	vaultCluster := os.Getenv("VAULT_CLUSTER")
	domain := os.Getenv("KUBERNETES_CLUSTER_DOMAIN")
	for _, v := range strings.Split(vaultCluster, ",") {
		// VAULT_FQDNS=vault-0.sip.svc.cluster.local,vault-1.sip.svc.cluster.local,vault-2.sip.svc.cluster.local
		vaultFQDNs = append(vaultFQDNs, v+"."+VAULT_INTERNAL_SERVICE_NAME+"."+ns+"."+"svc."+domain)
	}

	// setup client to vault-0
	clientV, err = getClientV(vaultFQDNs[0])
	if err != nil {
		return
	}

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
			unSeal()
			os.Exit(0)
		} else {
			log.Println("Vault is already unsealed.")
			os.Exit(0)
		}
	} else {
		initVault()
		os.Exit(0)
	}
}

// initVault performs init, unseal, raft join
func initVault() {
	wg.Add(3)
	defer wg.Wait()

	log.Println("Vault is not initialized yet. Initializing...")
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

	unsealKey := initResp.Data["keys_base64"].([]interface{})[0]
	// TODO: not safe, find another place to keep root token
	rootToken := initResp.Data["root_token"]

	// vault operator unseal vault-0/1/2
	go unSealWorker(vaultFQDNs[0], unsealKey.(string), &wg, false)
	go unSealWorker(vaultFQDNs[1], unsealKey.(string), &wg, true)
	go unSealWorker(vaultFQDNs[2], unsealKey.(string), &wg, true)

	// vault operator raft join
	// API is not available yet
	// https://github.com/hashicorp/vault-client-go/blob/main/docs/SystemApi.md
	// Here we use retry join
	// https://developer.hashicorp.com/vault/docs/platform/k8s/helm/examples/ha-tls

	nsByte, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns := string(nsByte)

	// persist to secrets
	log.Println("Create secret for Vault unseal key.")
	wantedSecretForUnsealKey := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-unseal-key",
		},
		StringData: map[string]string{
			"key": unsealKey.(string),
		},
		Type: corev1.SecretTypeOpaque,
	}

	_, err = clientK.CoreV1().Secrets(ns).Create(ctx, wantedSecretForUnsealKey, metav1.CreateOptions{})
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			log.Fatalf("Namespace %v not found\n", ns)
		case errors.IsAlreadyExists(err):
			log.Println("Secret vault-unseal-key already exists. Updating...")
			existedSecretForUnsealKey, _ := clientK.CoreV1().Secrets(ns).Get(ctx, "vault-unseal-key", metav1.GetOptions{})
			existedSecretForUnsealKey.Data["key"] = []byte(unsealKey.(string))
			_, err = clientK.CoreV1().Secrets(ns).Update(context.TODO(), existedSecretForUnsealKey, metav1.UpdateOptions{})
			if err != nil {
				log.Fatal(err)
			}
		case errors.IsInvalid(err):
			log.Fatal("Secret spec is invalid.\n")
		default:
			log.Fatal(err)
		}
	}

	// TODO: not safe, find another place to keep root token
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

	_, err = clientK.CoreV1().Secrets(ns).Create(ctx, wantedSecretForRootToken, metav1.CreateOptions{})
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			log.Fatalf("Namespace %v not found\n", ns)
		case errors.IsAlreadyExists(err):
			log.Println("Secret vault-root-token already exists. Updating...")
			existedSecretForRootToken, _ := clientK.CoreV1().Secrets(ns).Get(ctx, "vault-root-token", metav1.GetOptions{})
			existedSecretForRootToken.Data["token"] = []byte(rootToken.(string))
			_, err = clientK.CoreV1().Secrets(ns).Update(context.TODO(), existedSecretForRootToken, metav1.UpdateOptions{})
			if err != nil {
				log.Fatal(err)
			}
		case errors.IsInvalid(err):
			log.Fatal("Secret spec is invalid.\n")
		default:
			log.Fatal(err)
		}
	}

}

func unSeal() {
	wg.Add(3)
	defer wg.Wait()

	log.Println("Vault is sealed. Unsealing...")
	// get unseal key secret
	secretOfUnsealKey, err := clientK.CoreV1().Secrets(ns).Get(ctx, "vault-unseal-key", metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error getting secret: %v", err)
	}
	unsealKey := string(secretOfUnsealKey.Data["key"])

	// vault operator unseal vault-0/1/2
	go unSealWorker(vaultFQDNs[0], unsealKey, &wg, false)
	go unSealWorker(vaultFQDNs[1], unsealKey, &wg, false)
	go unSealWorker(vaultFQDNs[2], unsealKey, &wg, false)
}

func unSealWorker(fqdn, unsealKey string, wg *sync.WaitGroup, needSleep bool) {
	defer wg.Done()

	tmpClient, err := getClientV(fqdn)
	if err != nil {
		return
	}
	// TODO: a better way? u cannot detect if it's initialized right here, it's controlled by vault-2
	// [INFO] core: seal configuration missing, not initialized - every 5s
	if needSleep {
		time.Sleep(8 * time.Second)
	}
	log.Printf("%v is sealed. Unsealing...\n", fqdn)
	_, err = tmpClient.System.Unseal(ctx,
		schema.UnsealRequest{
			Key: unsealKey,
		})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v is now unsealed.\n", fqdn)
}

func getClientV(fqdn string) (*vault.Client, error) {
	tls := vault.TLSConfiguration{}
	tls.ServerCertificate.FromFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	log.Println("Setup client to " + fqdn)
	return vault.New(
		vault.WithAddress("https://"+fqdn+":8200"),
		vault.WithTLS(tls),
	)
}
