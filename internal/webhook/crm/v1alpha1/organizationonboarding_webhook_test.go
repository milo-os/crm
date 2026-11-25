// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

func sentOnboarding() (*crmv1alpha1.OrganizationOnboarding, *crmv1alpha1.OrganizationOnboarding) {
	old := &crmv1alpha1.OrganizationOnboarding{
		ObjectMeta: metav1.ObjectMeta{Name: "acme-onboarding"},
		Spec: crmv1alpha1.OrganizationOnboardingSpec{
			CompanyRef:     crmv1alpha1.CompanyReference{Name: "acme"},
			OpportunityRef: crmv1alpha1.OpportunityReference{Name: "acme-sale"},
			Billing:        crmv1alpha1.OnboardingBilling{Currency: crmv1alpha1.BillingCurrencyUSD},
		},
		Status: crmv1alpha1.OrganizationOnboardingStatus{Phase: crmv1alpha1.OnboardingPhaseInvited},
	}
	updated := old.DeepCopy()
	updated.Spec.OrganizationDetails = &crmv1alpha1.OnboardingOrganizationDetails{DisplayName: "Acme Corp"}
	return old, updated
}

func TestOrganizationOnboardingWebhookCreate(t *testing.T) {
	webhook := &organizationOnboardingWebhook{reader: emptyReader(t)}
	onboarding, _ := sentOnboarding()
	_, err := webhook.ValidateCreate(context.Background(), onboarding)
	assertInvalid(t, err, "OrganizationOnboarding", "acme-onboarding")
}

func TestOrganizationOnboardingWebhookChecksStatus(t *testing.T) {
	webhook := &organizationOnboardingWebhook{reader: emptyReader(t)}
	old, _ := sentOnboarding()
	updated := old.DeepCopy()
	updated.Status.OrganizationRef = &crmv1alpha1.OrganizationReference{Name: "org-8f3k2"}
	_, err := webhook.ValidateUpdate(context.Background(), old, updated)
	assertInvalid(t, err, "OrganizationOnboarding", "acme-onboarding")
}
