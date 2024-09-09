/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/KokoiRuby/service-identity-provider/operator/internal-client-ca/api/v1alpha1"
	"github.com/hashicorp/vault-client-go"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/hashicorp/vault-client-go/schema"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	clientCATTL   = "876000h"
	clientCertTTL = "168h"
	clientCARole  = "client-ca"

	patchBody = []byte(`{"client_flag":false,"server_flag":false}`)
)

// InternalClientCAReconciler reconciles a InternalClientCA object
type InternalClientCAReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ++
	ClientV *vault.Client
	ClientH *http.Client
}

// +kubebuilder:rbac:groups=sip.sec.com,resources=internalclientcas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sip.sec.com,resources=internalclientcas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sip.sec.com,resources=internalclientcas/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InternalClientCA object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *InternalClientCAReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here

	// 1. Load CR
	var intClientCA v1alpha1.InternalClientCA
	if err := r.Get(ctx, req.NamespacedName, &intClientCA); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("InternalClientCA resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to get InternalClientCA")
		return ctrl.Result{}, err
	}

	// 2. Parse CR
	clientCAPath := "sip-client-ca/" + intClientCA.Spec.Certificate.Subject.CN + "-ca/"

	// 3. create pki for client authn
	listResp, err := r.ClientV.System.MountsListSecretsEngines(ctx)
	if err != nil {
		logger.Error(err, "failed to list secrets engine")
		return ctrl.Result{}, err
	}

	// TODO: Robustness
	if _, ok := listResp.Data[clientCAPath]; !ok {
		logger.Info("Secrets engine " + clientCAPath + " not found. Enabling...")
		cert, err := r.enablePKIClientCA(ctx, clientCAPath)
		if err != nil {
			logger.Error(err, "failed to enable client ca pki")
			return ctrl.Result{}, err
		}
		secret, err := r.buildSecretFrom(&intClientCA, cert)
		if err != nil {
			logger.Error(err, "failed to define Secret for InternalClientCA")
			return ctrl.Result{}, err
		}
		err = r.Client.Create(ctx, secret)
		if err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	} else {
		logger.Info("Secrets engine " + clientCAPath + " is already enabled.")
		caCert, err := r.ClientV.Secrets.PkiReadCaPem(ctx, vault.WithMountPath(clientCAPath))
		if err != nil {
			logger.Error(err, "failed to get ca cert of "+clientCAPath+" secrets engine")
			return ctrl.Result{}, err
		}
		secret, err := r.buildSecretFrom(&intClientCA, caCert.Data.Certificate)
		if err != nil {
			logger.Error(err, "failed to create secret from InternalClientCA")
			return ctrl.Result{}, err
		}
		err = r.Client.Create(ctx, secret)
		if err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}
}

func (r *InternalClientCAReconciler) enablePKIClientCA(ctx context.Context, path string) (cert string, err error) {
	logger := log.FromContext(ctx)

	logger.Info("1. Enable " + path + " secrets engine.")
	_, err = r.ClientV.System.MountsEnableSecretsEngine(ctx, path,
		schema.MountsEnableSecretsEngineRequest{
			Type:        "pki",
			Description: "CA certificate backend created by sip for client authn",
		})
	if err != nil {
		logger.Error(err, "Failed to enable "+path+" secrets engine.")
		return "", err
	}

	logger.Info("2. Set up TTL for " + path + " secrets engine.")
	_, err = r.ClientV.System.MountsTuneConfigurationParameters(ctx, path,
		schema.MountsTuneConfigurationParametersRequest{
			MaxLeaseTtl: clientCATTL,
		})
	if err != nil {
		logger.Error(err, "Failed to set up TTL for "+path+" secrets engine.")
		return "", err
	}

	logger.Info("3. Configure Root CA keypair.")
	clientCAKeyPair, err := r.ClientV.Secrets.PkiGenerateRoot(ctx, "internal",
		schema.PkiGenerateRootRequest{
			CommonName: "sip Internal Root CA",
			KeyType:    "ec",
			KeyBits:    256,
			Ttl:        clientCATTL,
		},
		vault.WithMountPath(path))
	if err != nil {
		logger.Error(err, "Failed to configure CA pair for "+path+" secrets engine.")
		return "", err
	}

	logger.Info("4. Update CRL location & issuing certificates for " + path + " secrets engine.")
	_, err = r.ClientV.Secrets.PkiConfigureUrls(ctx,
		schema.PkiConfigureUrlsRequest{
			IssuingCertificates:   []string{"https://vault:8200/v1/" + path + "/ca"},
			CrlDistributionPoints: []string{"https://vault:8200/v1/" + path + "/crl"},
		},
		vault.WithMountPath(path))
	if err != nil {
		logger.Error(err, "Failed to update crl & issue cert &  for "+path+" secrets engine.")
		return "", err
	}

	logger.Info("5. Configure a role for " + path + " secrets engine to issue certificates.")
	_, err = r.ClientV.Secrets.PkiWriteRole(ctx, clientCARole,
		schema.PkiWriteRoleRequest{
			AllowAnyName:                  true,
			KeyType:                       "ec",
			KeyBits:                       256,
			KeyUsage:                      []string{"CertSign,CRLSign"},
			EnforceHostnames:              false,
			BasicConstraintsValidForNonCa: true,
			MaxTtl:                        clientCertTTL,
		},
		vault.WithMountPath(path))
	if err != nil {
		logger.Error(err, "Failed to configure role for "+path+" secrets engine.")
		return "", err
	}

	// TODO: how to pass token & set http token in header, fetch again?
	kubeConf, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	clientK := kubernetes.NewForConfigOrDie(kubeConf)
	nsByte, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns := string(nsByte)
	existedSecretForRootToken, err := clientK.CoreV1().Secrets(ns).Get(context.Background(), "vault-root-token", metav1.GetOptions{})
	if err != nil {
		return "", nil
	}
	rootToken := string(existedSecretForRootToken.Data["token"])
	disableFlag(path, clientCARole, rootToken)

	return clientCAKeyPair.Data.Certificate, nil
}

func (r *InternalClientCAReconciler) buildSecretFrom(intClientCA *v1alpha1.InternalClientCA, cert string) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intClientCA.Spec.Secret.Name,
			Namespace: intClientCA.Namespace,
		},
		StringData: map[string]string{
			intClientCA.Spec.Secret.CertName: cert,
		},
		Type: corev1.SecretTypeOpaque,
	}

	// set owner ref
	if err := ctrl.SetControllerReference(intClientCA, secret, r.Scheme); err != nil {
		return nil, err
	}

	return secret, nil

}

func disableFlag(pki, role, token string) {
	req, err := http.NewRequest("PATCH", "http://127.0.0.1:8200/v1/"+pki+"/roles/"+role, bytes.NewBuffer(patchBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/merge-patch+json")
	// TODO: how to set root token in a better way...
	req.Header.Set("X-Vault-Token", token)

	clientH := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := clientH.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InternalClientCAReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.InternalClientCA{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
