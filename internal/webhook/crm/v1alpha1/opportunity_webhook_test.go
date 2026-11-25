// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

func TestOpportunityWebhook(t *testing.T) {
	webhook := &opportunityWebhook{reader: emptyReader(t)}
	opportunity := &crmv1alpha1.Opportunity{
		ObjectMeta: metav1.ObjectMeta{Name: "acme-sale"},
		Spec: crmv1alpha1.OpportunitySpec{
			DisplayName: "Acme sale",
			CompanyRef:  crmv1alpha1.CompanyReference{Name: "acme"},
			OwnerRef:    crmv1alpha1.UserReference{Name: "340583683847097153"},
		},
	}
	withPerson := opportunity.DeepCopy()
	withPerson.Spec.PersonRefs = []crmv1alpha1.PersonReference{{Name: "jane"}}

	_, err := webhook.ValidateCreate(context.Background(), opportunity)
	assertInvalid(t, err, "Opportunity", "acme-sale")
	if _, err := webhook.ValidateUpdate(context.Background(), opportunity, opportunity); err != nil {
		t.Errorf("unchanged update: unexpected error %v", err)
	}
	_, err = webhook.ValidateUpdate(context.Background(), opportunity, withPerson)
	assertInvalid(t, err, "Opportunity", "acme-sale")
}
