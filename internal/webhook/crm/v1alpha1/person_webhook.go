// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
	validation "go.miloapis.com/crm/internal/validation/crm"
)

// +kubebuilder:webhook:path=/validate-crm-miloapis-com-v1alpha1-person,mutating=false,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=people;people/status,verbs=create;update,versions=v1alpha1,name=vperson.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook

// personWebhook checks that a Person's Company exists, and that the User CRM
// links them to exists. It also sees status writes, where CRM sets the link.
type personWebhook struct {
	reader client.Reader
}

var _ admission.Validator[*crmv1alpha1.Person] = &personWebhook{}

func (w *personWebhook) ValidateCreate(ctx context.Context, person *crmv1alpha1.Person) (admission.Warnings, error) {
	return nil, invalid("Person", person.Name, validation.ValidatePersonCreate(ctx, w.reader, person))
}

func (w *personWebhook) ValidateUpdate(ctx context.Context, oldPerson, newPerson *crmv1alpha1.Person) (admission.Warnings, error) {
	return nil, invalid("Person", newPerson.Name, validation.ValidatePersonUpdate(ctx, w.reader, oldPerson, newPerson))
}

// ValidateDelete isn't called: deleting a Person is always allowed.
// Opportunities that name them keep the link. Their invitations are owned by
// their onboardings, so CRM's controller deletes them (TODO(#13)).
func (w *personWebhook) ValidateDelete(context.Context, *crmv1alpha1.Person) (admission.Warnings, error) {
	return nil, nil
}
