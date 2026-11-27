// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

func TestPersonWebhook(t *testing.T) {
	webhook := &personWebhook{reader: emptyReader(t)}
	person := &crmv1alpha1.Person{
		ObjectMeta: metav1.ObjectMeta{Name: "jane"},
		Spec:       crmv1alpha1.PersonSpec{GivenName: "Jane", Email: "jane@example.com", CompanyRef: crmv1alpha1.CompanyReference{Name: "acme"}},
	}
	linked := person.DeepCopy()
	linked.Status.UserRef = &crmv1alpha1.UserReference{Name: "999"}

	_, err := webhook.ValidateCreate(context.Background(), person)
	assertInvalid(t, err, "Person", "jane")
	if _, err := webhook.ValidateUpdate(context.Background(), person, person); err != nil {
		t.Errorf("unchanged update: unexpected error %v", err)
	}
	_, err = webhook.ValidateUpdate(context.Background(), person, linked)
	assertInvalid(t, err, "Person", "jane")
}
