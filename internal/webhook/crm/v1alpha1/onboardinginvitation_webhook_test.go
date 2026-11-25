// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

func TestOnboardingInvitationWebhook(t *testing.T) {
	webhook := &onboardingInvitationWebhook{reader: emptyReader(t)}
	invitation := &crmv1alpha1.OnboardingInvitation{
		ObjectMeta: metav1.ObjectMeta{Name: "acme-jane"},
		Spec: crmv1alpha1.OnboardingInvitationSpec{
			OnboardingRef:  crmv1alpha1.OrganizationOnboardingReference{Name: "acme-onboarding"},
			PersonRef:      crmv1alpha1.PersonReference{Name: "jane"},
			RecipientEmail: "jane@example.com",
		},
	}
	redeemed := invitation.DeepCopy()
	redeemed.Status.RedeemedUserRef = &crmv1alpha1.UserReference{Name: "999"}

	_, err := webhook.ValidateCreate(context.Background(), invitation)
	assertInvalid(t, err, "OnboardingInvitation", "acme-jane")
	if _, err := webhook.ValidateUpdate(context.Background(), invitation, invitation); err != nil {
		t.Errorf("unchanged update: unexpected error %v", err)
	}
	_, err = webhook.ValidateUpdate(context.Background(), invitation, redeemed)
	assertInvalid(t, err, "OnboardingInvitation", "acme-jane")
}
