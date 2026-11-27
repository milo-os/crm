// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
	validation "go.miloapis.com/crm/internal/validation/crm"
)

// +kubebuilder:webhook:path=/validate-crm-miloapis-com-v1alpha1-opportunity,mutating=false,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=opportunities,verbs=create;update,versions=v1alpha1,name=vopportunity.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook

// opportunityWebhook checks that an Opportunity's Company and People exist,
// and that the People work for its Company.
type opportunityWebhook struct {
	reader client.Reader
}

var _ admission.Validator[*crmv1alpha1.Opportunity] = &opportunityWebhook{}

func (w *opportunityWebhook) ValidateCreate(ctx context.Context, opportunity *crmv1alpha1.Opportunity) (admission.Warnings, error) {
	return nil, invalid("Opportunity", opportunity.Name, validation.ValidateOpportunityCreate(ctx, w.reader, opportunity))
}

func (w *opportunityWebhook) ValidateUpdate(ctx context.Context, oldOpportunity, newOpportunity *crmv1alpha1.Opportunity) (admission.Warnings, error) {
	return nil, invalid("Opportunity", newOpportunity.Name, validation.ValidateOpportunityUpdate(ctx, w.reader, oldOpportunity, newOpportunity))
}

// ValidateDelete isn't called: deleting an Opportunity is always allowed.
// Onboardings that name it keep the link; they are owned by the Company.
func (w *opportunityWebhook) ValidateDelete(context.Context, *crmv1alpha1.Opportunity) (admission.Warnings, error) {
	return nil, nil
}
