// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	examplev1alpha1 "go.miloapis.com/crm/api/example/v1alpha1"
	validation "go.miloapis.com/crm/internal/validation/example"
)

// ResourceDefaultDescription is set on Resources created without a
// description.
const ResourceDefaultDescription = "No description provided."

var resourceLog = logf.Log.WithName("resource-webhook")

// SetupResourceWebhookWithManager registers the Resource webhook with the
// manager.
func SetupResourceWebhookWithManager(mgr ctrl.Manager) error {
	webhook := &resourceWebhook{}

	return ctrl.NewWebhookManagedBy(mgr, &examplev1alpha1.Resource{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-example-miloapis-com-v1alpha1-resource,mutating=true,failurePolicy=fail,sideEffects=None,groups=example.miloapis.com,resources=resources,verbs=create;update,versions=v1alpha1,name=mresource.kb.io,admissionReviewVersions=v1,serviceName=crm-webhook

// +kubebuilder:webhook:path=/validate-example-miloapis-com-v1alpha1-resource,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.miloapis.com,resources=resources,verbs=create;update;delete,versions=v1alpha1,name=vresource.kb.io,admissionReviewVersions=v1,serviceName=crm-webhook

type resourceWebhook struct{}

var _ admission.Defaulter[*examplev1alpha1.Resource] = &resourceWebhook{}
var _ admission.Validator[*examplev1alpha1.Resource] = &resourceWebhook{}

// Default implements webhook.CustomDefaulter. It fills in a description
// when none is given.
func (r *resourceWebhook) Default(ctx context.Context, resource *examplev1alpha1.Resource) error {
	resourceLog.Info("defaulting", "name", resource.GetName())

	if resource.Spec.Description == "" {
		resource.Spec.Description = ResourceDefaultDescription
	}
	return nil
}

// ValidateCreate implements webhook.CustomValidator.
func (r *resourceWebhook) ValidateCreate(ctx context.Context, resource *examplev1alpha1.Resource) (admission.Warnings, error) {
	resourceLog.Info("validating create", "name", resource.GetName())

	if errs := validation.ValidateResourceCreate(resource); len(errs) > 0 {
		return nil, apierrors.NewInvalid(resourceGroupKind(), resource.Name, errs)
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator.
func (r *resourceWebhook) ValidateUpdate(ctx context.Context, oldResource, newResource *examplev1alpha1.Resource) (admission.Warnings, error) {
	resourceLog.Info("validating update", "name", newResource.GetName())

	if errs := validation.ValidateResourceUpdate(oldResource, newResource); len(errs) > 0 {
		return nil, apierrors.NewInvalid(resourceGroupKind(), newResource.Name, errs)
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator. It refuses deletion
// while deletion protection is enabled.
func (r *resourceWebhook) ValidateDelete(ctx context.Context, resource *examplev1alpha1.Resource) (admission.Warnings, error) {
	resourceLog.Info("validating delete", "name", resource.GetName())

	if errs := validation.ValidateResourceDelete(resource); len(errs) > 0 {
		return nil, apierrors.NewInvalid(resourceGroupKind(), resource.Name, errs)
	}
	return nil, nil
}

func resourceGroupKind() schema.GroupKind {
	return examplev1alpha1.GroupVersion.WithKind("Resource").GroupKind()
}
