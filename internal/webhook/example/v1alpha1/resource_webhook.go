// SPDX-License-Identifier: AGPL-3.0-only

package webhook

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	examplev1alpha1 "go.miloapis.com/crm/api/example/v1alpha1"
)

var resourceLog = logf.Log.WithName("resource-webhook")

// SetupWebhookWithManager registers the Resource webhook with the manager.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	webhook := &resourceWebhook{}

	return ctrl.NewWebhookManagedBy(mgr, &examplev1alpha1.Resource{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-example-miloapis-com-v1alpha1-resource,mutating=true,failurePolicy=fail,sideEffects=None,groups=example.miloapis.com,resources=resources,verbs=create;update,versions=v1alpha1,name=mresource.kb.io,admissionReviewVersions=v1

// +kubebuilder:webhook:path=/validate-example-miloapis-com-v1alpha1-resource,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.miloapis.com,resources=resources,verbs=create;update;delete,versions=v1alpha1,name=vresource.kb.io,admissionReviewVersions=v1

type resourceWebhook struct{}

var _ admission.Defaulter[*examplev1alpha1.Resource] = &resourceWebhook{}
var _ admission.Validator[*examplev1alpha1.Resource] = &resourceWebhook{}

// Default implements admission.Defaulter.
func (r *resourceWebhook) Default(ctx context.Context, resource *examplev1alpha1.Resource) error {
	resourceLog.Info("defaulting", "name", resource.GetName())

	// TODO: add defaulting logic here

	return nil
}

// ValidateCreate implements admission.Validator.
func (r *resourceWebhook) ValidateCreate(ctx context.Context, resource *examplev1alpha1.Resource) (admission.Warnings, error) {
	resourceLog.Info("validating create", "name", resource.GetName())

	// TODO: add validation logic here

	return nil, nil
}

// ValidateUpdate implements admission.Validator.
func (r *resourceWebhook) ValidateUpdate(ctx context.Context, oldResource, newResource *examplev1alpha1.Resource) (admission.Warnings, error) {
	resourceLog.Info("validating update", "name", newResource.GetName())

	// TODO: add validation logic here

	return nil, nil
}

// ValidateDelete implements admission.Validator.
func (r *resourceWebhook) ValidateDelete(ctx context.Context, resource *examplev1alpha1.Resource) (admission.Warnings, error) {
	resourceLog.Info("validating delete", "name", resource.GetName())

	// TODO: add validation logic here

	return nil, nil
}
