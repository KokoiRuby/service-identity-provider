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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sipv1alpha1 "github.com/KokoiRuby/service-identity-provider/operator/internal-cert/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InternalCertificateReconciler reconciles a InternalCertificate object
type InternalCertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
		logger.Error(err, "unable to get InternalCertificate")
		return ctrl.Result{}, err
	}

	// 2. Parse CR
	certificate := intCert.Spec.Certificate
	secret := intCert.Spec.Secret

	logger.Info("Parsing InternalCertificate",
		"Issuer", certificate.Issuer.Reference,
		"Subject CN", certificate.Subject.CN,
		"Extended Key Usage", certificate.ExtendedKeyUsage,
		"Secret Name", secret.Name,
		"Secret Key Name", secret.KeyName,
		"Secret Certiticate Name", secret.CertName)

	// 3. vault api & get keypair info

	// 4. persist into secret
	s, err := r.secretFromIntCert(&intCert)
	if err != nil {
		logger.Error(err, "failed to define Secret for InternalCertificate")
		return ctrl.Result{}, err
	}

	err = r.Client.Create(ctx, s)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return ctrl.Result{}, fmt.Errorf("failed to create Secret: %v", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *InternalCertificateReconciler) secretFromIntCert(intCert *sipv1alpha1.InternalCertificate) (*corev1.Secret, error) {
	key := "this is a key"
	cert := "this is a cert"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intCert.Spec.Secret.Name,
			Namespace: intCert.Namespace,
		},
		Data: map[string][]byte{
			"tls.key": []byte(key),
			"tls.crt": []byte(cert),
		},
		Type: corev1.SecretTypeTLS,
	}

	// set owner ref
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
