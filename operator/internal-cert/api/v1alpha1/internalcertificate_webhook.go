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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var internalcertificatelog = logf.Log.WithName("internalcertificate-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *InternalCertificate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-sip-sec-com-v1alpha1-internalcertificate,mutating=true,failurePolicy=fail,sideEffects=None,groups=sip.sec.com,resources=internalcertificates,verbs=create;update,versions=v1alpha1,name=minternalcertificate.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &InternalCertificate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *InternalCertificate) Default() {
	internalcertificatelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-sip-sec-com-v1alpha1-internalcertificate,mutating=false,failurePolicy=fail,sideEffects=None,groups=sip.sec.com,resources=internalcertificates,verbs=create;update,versions=v1alpha1,name=vinternalcertificate.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &InternalCertificate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *InternalCertificate) ValidateCreate() (admission.Warnings, error) {
	internalcertificatelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, r.validateInternalCertificate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *InternalCertificate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	internalcertificatelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, r.validateInternalCertificate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *InternalCertificate) ValidateDelete() (admission.Warnings, error) {
	internalcertificatelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *InternalCertificate) validateInternalCertificate() error {
	var allErrs field.ErrorList
	if err := r.validateInternalCertificateSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{
		Group: "sip.sec.com",
		Kind:  "InternalCertificate",
	}, r.Name, allErrs)
}

func (r *InternalCertificate) validateInternalCertificateSpec() *field.Error {
	return validateCertificate(r.Spec.Certificate, field.NewPath("spec").Child("certificate"))
}

func validateCertificate(certificate Certificate, fldPath *field.Path) *field.Error {
	if certificate.ServerAuth == certificate.ClientAuth {
		return field.Invalid(fldPath, certificate, "clientAuth and serverAuth cannot both be true or both be false")
	}
	return nil
}
