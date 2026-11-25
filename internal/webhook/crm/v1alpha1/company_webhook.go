// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
	validation "go.miloapis.com/crm/internal/validation/crm"
)

// +kubebuilder:webhook:path=/validate-crm-miloapis-com-v1alpha1-company,mutating=false,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=companies,verbs=create;update,versions=v1alpha1,name=vcompany.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook

// companyWebhook checks that the organizations a Company links exist.
type companyWebhook struct {
	reader client.Reader
}

var _ admission.Validator[*crmv1alpha1.Company] = &companyWebhook{}

func (w *companyWebhook) ValidateCreate(ctx context.Context, company *crmv1alpha1.Company) (admission.Warnings, error) {
	return nil, invalid("Company", company.Name, validation.ValidateCompanyCreate(ctx, w.reader, company))
}

func (w *companyWebhook) ValidateUpdate(ctx context.Context, oldCompany, newCompany *crmv1alpha1.Company) (admission.Warnings, error) {
	return nil, invalid("Company", newCompany.Name, validation.ValidateCompanyUpdate(ctx, w.reader, oldCompany, newCompany))
}

// ValidateDelete isn't called: deleting a Company is always allowed. Milo's
// garbage collector then deletes the People, Opportunities, and onboardings
// it owns, and the onboardings' invitations. Its organizations are never
// touched.
func (w *companyWebhook) ValidateDelete(context.Context, *crmv1alpha1.Company) (admission.Warnings, error) {
	return nil, nil
}
