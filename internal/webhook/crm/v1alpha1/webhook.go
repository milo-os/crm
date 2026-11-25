// SPDX-License-Identifier: AGPL-3.0-only

// Package v1alpha1 holds the admission webhooks for crm.miloapis.com
// resources. They only validate: every rule lives in internal/validation/crm,
// and every webhook fails closed, so a request is refused when the webhook
// can't be reached or a referenced record can't be read.
package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// The webhooks read the records a request references, including records
// other Milo services own.
// +kubebuilder:rbac:groups=crm.miloapis.com,resources=companies;people;opportunities;organizationonboardings;onboardinginvitations,verbs=get
// +kubebuilder:rbac:groups=resourcemanager.miloapis.com,resources=organizations,verbs=get
// +kubebuilder:rbac:groups=iam.miloapis.com,resources=users,verbs=get
// +kubebuilder:rbac:groups=billing.miloapis.com,resources=billingaccounts,verbs=get

// SetupWebhooksWithManager registers every CRM webhook with the manager.
// References are read with the manager's uncached API reader, so a record
// created a moment earlier is found.
func SetupWebhooksWithManager(mgr ctrl.Manager) error {
	reader := mgr.GetAPIReader()

	if err := ctrl.NewWebhookManagedBy(mgr, &crmv1alpha1.Company{}).
		WithValidator(&companyWebhook{reader: reader}).Complete(); err != nil {
		return err
	}
	if err := ctrl.NewWebhookManagedBy(mgr, &crmv1alpha1.Person{}).
		WithValidator(&personWebhook{reader: reader}).Complete(); err != nil {
		return err
	}
	if err := ctrl.NewWebhookManagedBy(mgr, &crmv1alpha1.Opportunity{}).
		WithValidator(&opportunityWebhook{reader: reader}).Complete(); err != nil {
		return err
	}
	if err := ctrl.NewWebhookManagedBy(mgr, &crmv1alpha1.OrganizationOnboarding{}).
		WithValidator(&organizationOnboardingWebhook{reader: reader}).Complete(); err != nil {
		return err
	}
	return ctrl.NewWebhookManagedBy(mgr, &crmv1alpha1.OnboardingInvitation{}).
		WithValidator(&onboardingInvitationWebhook{reader: reader}).Complete()
}

// invalid turns field errors into the API error the webhook returns, or nil
// when there are none.
func invalid(kind, name string, errs field.ErrorList) error {
	if len(errs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(crmv1alpha1.GroupVersion.WithKind(kind).GroupKind(), name, errs)
}
