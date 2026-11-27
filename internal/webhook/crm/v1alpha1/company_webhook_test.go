// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

func TestCompanyWebhook(t *testing.T) {
	webhook := &companyWebhook{reader: emptyReader(t)}
	plain := &crmv1alpha1.Company{ObjectMeta: metav1.ObjectMeta{Name: "acme"}, Spec: crmv1alpha1.CompanySpec{DisplayName: "Acme"}}
	linked := plain.DeepCopy()
	linked.Spec.OrganizationRefs = []crmv1alpha1.OrganizationReference{{Name: "org-8f3k2"}}

	if _, err := webhook.ValidateCreate(context.Background(), plain); err != nil {
		t.Errorf("create without organizations: unexpected error %v", err)
	}
	_, err := webhook.ValidateCreate(context.Background(), linked)
	assertInvalid(t, err, "Company", "acme")
	_, err = webhook.ValidateUpdate(context.Background(), plain, linked)
	assertInvalid(t, err, "Company", "acme")
}
