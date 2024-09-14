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
	"context"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sipv1alpha1 "github.com/KokoiRuby/service-identity-provider/operator/internal-cert/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const VAULT_INTERNAL_SERVICE_NAME = "vault-internal"

// InternalCertificateReconciler reconciles a InternalCertificate object
type InternalCertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ++
	ClientV *vault.Client
	ClientH *http.Client
}

// +kubebuilder:rbac:groups=sip.sec.com,resources=internalcertificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sip.sec.com,resources=internalcertificates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sip.sec.com,resources=internalcertificates/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InternalCertificate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *InternalCertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here

	// 1. Load CR
	var intCert sipv1alpha1.InternalCertificate
	if err := r.Get(ctx, req.NamespacedName, &intCert); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("InternalCertificate resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to get InternalCertificate")
		return ctrl.Result{}, err
	}

	// 2. Parse CR and determine which pki to call
	// TODO: ++Validation Webhook, so we could skip if cond, to be verified
	if intCert.Spec.Certificate.ExtendedKeyUsage.ClientAuth {
		// client authn pki
		cn := intCert.Spec.Certificate.Subject.CN
		issuer := intCert.Spec.Certificate.Issuer.Reference

		logger.Info("Issue CA from Client CA")
		issueResp, err := r.ClientV.Secrets.PkiIssueWithRole(ctx, "client-ca",
			schema.PkiIssueWithRoleRequest{
				CommonName: cn,
				AltNames:   cn + ", " + cn + ".sip, " + cn + ".sip.svc, " + cn + ".sip.svc.cluster.local",
			},
			vault.WithMountPath("sip-client-ca/"+issuer))
		if err != nil {
			logger.Error(err, "Unable to issue client-ca")
			return ctrl.Result{}, err
		}

		logger.Info("Build secret from InternalCert")
		secret, err := r.buildSecretFrom(&intCert, issueResp)
		if err != nil {
			logger.Error(err, "Unable to build secret from InternalCertificate")
			return ctrl.Result{}, err
		}
		logger.Info("Created secret", "secret", secret.Name)
		logger.Info("Created secret", "secret", secret.Data["tls.crt"])
		logger.Info("Created secret", "secret", secret.Data["tls.key"])

		logger.Info("Create secret")
		err = r.Client.Create(ctx, secret)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				logger.Info("Secret already exists. Updating...")
				err = r.Client.Update(ctx, secret)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
			logger.Error(err, "Unable to create secret")
			return ctrl.Result{}, err
		}

	} else if intCert.Spec.Certificate.ExtendedKeyUsage.ServerAuth {

		// check if secret exists
		var secret corev1.Secret
		objKey := types.NamespacedName{
			Namespace: req.Namespace,
			Name:      intCert.Spec.Secret.Name,
		}
		err := r.Client.Get(ctx, objKey, &secret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Secret does not exist. Creating...")

				// server authn pki
				cn := intCert.Spec.Certificate.Subject.CN
				issueResp, err := r.ClientV.Secrets.PkiIssueWithRole(ctx, "interm-ca",
					schema.PkiIssueWithRoleRequest{
						CommonName: cn,
						AltNames:   cn + ", " + cn + ".sip, " + cn + ".sip.svc, " + cn + ".sip.svc.cluster.local",
					},
					vault.WithMountPath("sip-interm-ca"))
				if err != nil {
					logger.Error(err, "Unable to issue interm-ca")
					return ctrl.Result{}, err
				}

				// build secret from
				// TODO: ++ ca chain
				secret, err := r.buildSecretFrom(&intCert, issueResp)
				if err != nil {
					logger.Error(err, "Unable to build secret from InternalCertificate")
					return ctrl.Result{}, err
				}

				// create secret
				err = r.Client.Create(ctx, secret)
				if err != nil {
					logger.Error(err, "Unable to create secret")
					return ctrl.Result{}, err
				}
			}
		}
		return ctrl.Result{}, nil
	} else {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *InternalCertificateReconciler) buildSecretFrom(intCert *sipv1alpha1.InternalCertificate, issueResp *vault.Response[schema.PkiIssueWithRoleResponse]) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intCert.Spec.Secret.Name,
			Namespace: intCert.Namespace,
		},
		StringData: map[string]string{
			intCert.Spec.Secret.KeyName: issueResp.Data.PrivateKey,
			// TODO: ++ ca chain
			intCert.Spec.Secret.CertName: issueResp.Data.CaChain[0] + "\n" + issueResp.Data.Certificate,
		},
		// Type: corev1.SecretTypeTLS
		// KubeAPIWarningLogger    tls: private key does not match public key if ++ ca chain
		Type: corev1.SecretTypeOpaque,
	}

	if err := ctrl.SetControllerReference(intCert, secret, r.Scheme); err != nil {
		return nil, err
	}

	return secret, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InternalCertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sipv1alpha1.InternalCertificate{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
