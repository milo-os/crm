// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
	validation "go.miloapis.com/crm/internal/validation/crm"
)

// +kubebuilder:webhook:path=/validate-crm-miloapis-com-v1alpha1-onboardinginvitation,mutating=false,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=onboardinginvitations;onboardinginvitations/status,verbs=create;update,versions=v1alpha1,name=vonboardinginvitation.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook

// onboardingInvitationWebhook checks an invitation's onboarding and Person,
// and that the Person works for the onboarding's Company. On update, which
// includes CRM's status writes, it checks the inviter and the User who
// redeemed it; the spec's schema rules keep it from changing.
type onboardingInvitationWebhook struct {
	reader client.Reader
}

var _ admission.Validator[*crmv1alpha1.OnboardingInvitation] = &onboardingInvitationWebhook{}

func (w *onboardingInvitationWebhook) ValidateCreate(ctx context.Context, invitation *crmv1alpha1.OnboardingInvitation) (admission.Warnings, error) {
	return nil, invalid("OnboardingInvitation", invitation.Name, validation.ValidateOnboardingInvitationCreate(ctx, w.reader, invitation))
}

func (w *onboardingInvitationWebhook) ValidateUpdate(ctx context.Context, oldInvitation, newInvitation *crmv1alpha1.OnboardingInvitation) (admission.Warnings, error) {
	return nil, invalid("OnboardingInvitation", newInvitation.Name, validation.ValidateOnboardingInvitationUpdate(ctx, w.reader, oldInvitation, newInvitation))
}

// ValidateDelete isn't called: deleting an invitation is always allowed.
func (w *onboardingInvitationWebhook) ValidateDelete(context.Context, *crmv1alpha1.OnboardingInvitation) (admission.Warnings, error) {
	return nil, nil
}
