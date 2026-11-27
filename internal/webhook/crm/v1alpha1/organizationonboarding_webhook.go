// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
	validation "go.miloapis.com/crm/internal/validation/crm"
)

// +kubebuilder:webhook:path=/validate-crm-miloapis-com-v1alpha1-organizationonboarding,mutating=false,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=organizationonboardings;organizationonboardings/status,verbs=create;update,versions=v1alpha1,name=vorganizationonboarding.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook

// organizationOnboardingWebhook checks an onboarding's Company and
// Opportunity, and the references CRM records in status, which it sees
// through status writes.
type organizationOnboardingWebhook struct {
	reader client.Reader
}

var _ admission.Validator[*crmv1alpha1.OrganizationOnboarding] = &organizationOnboardingWebhook{}

func (w *organizationOnboardingWebhook) ValidateCreate(ctx context.Context, onboarding *crmv1alpha1.OrganizationOnboarding) (admission.Warnings, error) {
	return nil, invalid("OrganizationOnboarding", onboarding.Name, validation.ValidateOrganizationOnboardingCreate(ctx, w.reader, onboarding))
}

func (w *organizationOnboardingWebhook) ValidateUpdate(ctx context.Context, oldOnboarding, newOnboarding *crmv1alpha1.OrganizationOnboarding) (admission.Warnings, error) {
	return nil, invalid("OrganizationOnboarding", newOnboarding.Name, validation.ValidateOrganizationOnboardingUpdate(ctx, w.reader, oldOnboarding, newOnboarding))
}

// ValidateDelete isn't called: deleting an onboarding is always allowed.
// Milo's garbage collector then deletes its invitations. The organization and
// billing account it created are never touched.
func (w *organizationOnboardingWebhook) ValidateDelete(context.Context, *crmv1alpha1.OrganizationOnboarding) (admission.Warnings, error) {
	return nil, nil
}
