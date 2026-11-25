// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

func ownerRef(kind, name string) metav1.OwnerReference {
	return metav1.OwnerReference{APIVersion: crmv1alpha1.GroupVersion.String(), Kind: kind, Name: name, UID: types.UID("uid-" + name)}
}

func TestOwnerReferences(t *testing.T) {
	ctx := context.Background()
	c := records(t)
	owned := func(p *crmv1alpha1.Person, refs ...metav1.OwnerReference) *crmv1alpha1.Person {
		p.OwnerReferences = refs
		return p
	}

	t.Run("create owned by its Company", func(t *testing.T) {
		assertErrors(t, ValidatePersonCreate(ctx, c, owned(person("ann", "acme"), ownerRef("Company", "acme"))))
	})
	t.Run("create with a second owner", func(t *testing.T) {
		assertErrors(t, ValidatePersonCreate(ctx, c, owned(person("ann", "acme"), ownerRef("Company", "acme"), ownerRef("Opportunity", "acme-sale"))),
			`metadata.ownerReferences[1]: Forbidden: the record can only be owned by its Company acme`)
	})
	t.Run("create owned by another Company", func(t *testing.T) {
		assertErrors(t, ValidatePersonCreate(ctx, c, owned(person("ann", "acme"), ownerRef("Company", "globex"))),
			`metadata.ownerReferences[0]: Forbidden`)
	})
	t.Run("update keeping the owner", func(t *testing.T) {
		old := owned(person("jane", "acme"), ownerRef("Company", "acme"))
		updated := old.DeepCopy()
		updated.Spec.Email = "jane.doe@example.com"
		assertErrors(t, ValidatePersonUpdate(ctx, c, old, updated))
	})
	t.Run("update removing the owner", func(t *testing.T) {
		old := owned(person("jane", "acme"), ownerRef("Company", "acme"))
		assertErrors(t, ValidatePersonUpdate(ctx, c, old, owned(person("jane", "acme"))),
			`metadata.ownerReferences: Forbidden: owner references can't be changed`)
	})
	t.Run("update adding an owner", func(t *testing.T) {
		old := owned(person("jane", "acme"), ownerRef("Company", "acme"))
		updated := owned(person("jane", "acme"), ownerRef("Company", "acme"), ownerRef("Company", "globex"))
		assertErrors(t, ValidatePersonUpdate(ctx, c, old, updated), `owner references can't be changed`)
	})
	t.Run("Opportunity update removing the owner", func(t *testing.T) {
		old := opportunity("acme-sale", "acme", "jane")
		old.OwnerReferences = []metav1.OwnerReference{ownerRef("Company", "acme")}
		assertErrors(t, ValidateOpportunityUpdate(ctx, c, old, opportunity("acme-sale", "acme", "jane")), `owner references can't be changed`)
	})
	t.Run("onboarding update removing the owner", func(t *testing.T) {
		old := onboarding("acme-onboarding", "acme", "acme-sale", "")
		old.OwnerReferences = []metav1.OwnerReference{ownerRef("Company", "acme")}
		assertErrors(t, ValidateOrganizationOnboardingUpdate(ctx, c, old, onboarding("acme-onboarding", "acme", "acme-sale", "")), `owner references can't be changed`)
	})
	t.Run("invitation owned by its Person instead of its onboarding", func(t *testing.T) {
		i := invitation("acme-onboarding", "jane")
		i.OwnerReferences = []metav1.OwnerReference{ownerRef("Person", "jane")}
		assertErrors(t, ValidateOnboardingInvitationCreate(ctx, c, i), `can only be owned by its OrganizationOnboarding acme-onboarding`)
	})
	t.Run("invitation update removing the owner", func(t *testing.T) {
		old := invitation("acme-onboarding", "jane")
		old.OwnerReferences = []metav1.OwnerReference{ownerRef("OrganizationOnboarding", "acme-onboarding")}
		assertErrors(t, ValidateOnboardingInvitationUpdate(ctx, c, old, invitation("acme-onboarding", "jane")), `owner references can't be changed`)
	})
}
