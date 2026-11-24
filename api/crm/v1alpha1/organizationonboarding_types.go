// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// OnboardingPhase is where an OrganizationOnboarding is in its lifecycle.
//
// +kubebuilder:validation:Enum=Draft;Invited;InProgress;Completing;Complete;Cancelled
type OnboardingPhase string

const (
	// OnboardingPhaseDraft is an onboarding staff are still preparing. No
	// invitation has been sent.
	OnboardingPhaseDraft OnboardingPhase = "Draft"
	// OnboardingPhaseInvited is an onboarding whose invitation has been sent
	// but not yet redeemed.
	OnboardingPhaseInvited OnboardingPhase = "Invited"
	// OnboardingPhaseInProgress is an onboarding the customer has opened and
	// is completing.
	OnboardingPhaseInProgress OnboardingPhase = "InProgress"
	// OnboardingPhaseCompleting is an onboarding the customer has submitted,
	// while CRM finishes setting up the organization, billing, and
	// administrator access.
	OnboardingPhaseCompleting OnboardingPhase = "Completing"
	// OnboardingPhaseComplete is an onboarding whose organization is ready to
	// use.
	OnboardingPhaseComplete OnboardingPhase = "Complete"
	// OnboardingPhaseCancelled is an onboarding staff ended before it
	// completed. It is final; staff start a new onboarding if the deal
	// resumes.
	OnboardingPhaseCancelled OnboardingPhase = "Cancelled"
)

// BillingCurrency is an ISO 4217 currency code Billing can bill in.
//
// +kubebuilder:validation:Enum=USD
type BillingCurrency string

const (
	// BillingCurrencyUSD is US dollars.
	BillingCurrencyUSD BillingCurrency = "USD"
)

// PaymentArrangement is how the customer will pay.
//
// +kubebuilder:validation:Enum=Card;Invoice
type PaymentArrangement string

const (
	// PaymentArrangementCard means the customer adds a payment method during
	// onboarding.
	PaymentArrangementCard PaymentArrangement = "Card"
	// PaymentArrangementInvoice means the customer is invoiced on Billing's
	// standard terms and adds no payment method. CRM grants it with an
	// Invoice BillingArrangement.
	PaymentArrangementInvoice PaymentArrangement = "Invoice"
)

// OrganizationOnboardingSpec is what staff prepare, and the customer later
// completes, to set up one organization.
//
// TODO(#11): once the onboarding is sent, reject staff changes to the spec.
// Staff-prepared details, the currency, and the payment arrangement are frozen
// at send; after that only CRM writes the spec, saving the customer's
// corrections.
type OrganizationOnboardingSpec struct {
	// CompanyRef is the Company the organization is for. It can't be changed.
	//
	// TODO(#11): reject a Company that doesn't exist.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="companyRef can't be changed"
	CompanyRef CompanyReference `json:"companyRef"`

	// OpportunityRef is the Opportunity the onboarding started from. It can't
	// be changed. An Opportunity can have more than one onboarding, each
	// setting up its own organization.
	//
	// TODO(#11): reject an Opportunity that doesn't exist or whose companyRef
	// is not this onboarding's Company.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="opportunityRef can't be changed"
	OpportunityRef OpportunityReference `json:"opportunityRef"`

	// OrganizationDetails are the details of the organization to create.
	// Staff fill in what they know; the customer completes the rest.
	//
	// +optional
	OrganizationDetails *OnboardingOrganizationDetails `json:"organizationDetails,omitempty"`

	// Billing is how the organization will be billed.
	//
	// +kubebuilder:validation:Required
	Billing OnboardingBilling `json:"billing"`
}

// OnboardingOrganizationDetails are the details CRM sets on the Milo
// Organization (resourcemanager.miloapis.com). They follow the Organization's
// fields; all are optional while the onboarding is a draft.
type OnboardingOrganizationDetails struct {
	// DisplayName is the organization's human-readable name.
	//
	// TODO(#14): on submit, require it or fill it in from the contact's
	// business name, then their name, as the cloud portal does.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:XValidation:rule="!self.matches('^[[:space:]]*$')",message="must not be blank"
	DisplayName string `json:"displayName,omitempty"`

	// ContactInfo is who the organization is and how to reach them.
	//
	// +optional
	ContactInfo *OnboardingContactInfo `json:"contactInfo,omitempty"`
}

// OnboardingBilling is the billing for the organization: the currency and
// payment arrangement staff choose, the billing contact, tax registrations,
// and invoice details.
//
// +kubebuilder:validation:XValidation:rule="!has(self.invoice) || self.paymentArrangement == 'Invoice'",message="invoice details can only be set when paymentArrangement is Invoice"
type OnboardingBilling struct {
	// Currency is the billing account's currency. Staff choose it and the
	// customer can't change it. Billing only bills in USD today; more
	// currencies are added here when Billing supports them.
	//
	// +kubebuilder:validation:Required
	Currency BillingCurrency `json:"currency"`

	// PaymentArrangement is how the customer will pay: Card or Invoice. Staff
	// choose it and the customer can't change it. Defaults to Card.
	//
	// +optional
	// +kubebuilder:default=Card
	PaymentArrangement PaymentArrangement `json:"paymentArrangement,omitempty"`

	// ContactInfo is the billing contact and the address invoices are issued
	// to. It follows the BillingAccount's contactInfo
	// (billing.miloapis.com). CRM never stores payment credentials.
	//
	// +optional
	ContactInfo *OnboardingBillingContactInfo `json:"contactInfo,omitempty"`

	// TaxIDs are the organization's tax registrations, such as a VAT number,
	// printed on invoices. They follow the BillingAccount's taxIds: one
	// entry per type.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=10
	TaxIDs []TaxID `json:"taxIds,omitempty"`

	// Invoice holds details for invoices when paymentArrangement is Invoice.
	// CRM sets them on the Invoice BillingArrangement it creates. Billing's
	// standard terms always apply; custom terms are out of scope.
	//
	// TODO(#14): decide whether the customer can correct these when saving
	// details, or only staff before sending.
	//
	// +optional
	Invoice *OnboardingInvoiceDetails `json:"invoice,omitempty"`
}

// TaxID is a tax registration. It follows the BillingAccount's taxIds.
type TaxID struct {
	// Type is the tax registration scheme, written as
	// <jurisdiction>_<scheme>, such as gb_vat, eu_vat, or us_ein.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-z]{2}_[a-z][a-z_]*$')",message="must be <jurisdiction>_<scheme>, such as gb_vat, eu_vat, or us_ein"
	Type string `json:"type"`

	// Value is the registration number, such as GB123456789.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:XValidation:rule="!self.matches('^[[:space:]]*$')",message="must not be blank"
	Value string `json:"value"`
}

// OnboardingInvoiceDetails are details printed on or attached to invoices.
// They follow the invoice terms on Billing's BillingArrangement, without the
// terms themselves: no net days or credit limit.
type OnboardingInvoiceDetails struct {
	// PurchaseOrderNumber is the customer's purchase order number, printed
	// on invoices.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=64
	PurchaseOrderNumber string `json:"purchaseOrderNumber,omitempty"`

	// AgreementReference identifies a signed MSA or order form. Without one,
	// the standard terms of service apply.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=128
	AgreementReference string `json:"agreementReference,omitempty"`

	// AccountsPayableEmail receives invoices in addition to the billing
	// contact.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=254
	// +kubebuilder:validation:XValidation:rule="self.matches('^[^@[:space:]]+@[^@[:space:]]+[.][^@[:space:]]+$')",message="must be an email address, such as jane@example.com"
	AccountsPayableEmail string `json:"accountsPayableEmail,omitempty"`
}

// OnboardingContactInfo is who an organization is and how to reach them. It
// follows the Organization's contactInfo.
//
// TODO(#14): require name and email on submit. Milo only marks an
// organization ready once its contact has both.
type OnboardingContactInfo struct {
	// Name is the display name of the primary contact.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name,omitempty"`

	// Email is the primary contact email.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=254
	// +kubebuilder:validation:XValidation:rule="self.matches('^[^@[:space:]]+@[^@[:space:]]+[.][^@[:space:]]+$')",message="must be an email address, such as jane@example.com"
	Email string `json:"email,omitempty"`

	// BusinessName is the legal entity or company name.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=256
	BusinessName string `json:"businessName,omitempty"`

	// Address is the organization's postal address.
	//
	// +optional
	Address *PostalAddress `json:"address,omitempty"`
}

// OnboardingBillingContactInfo is the billing contact. It follows the
// BillingAccount's contactInfo.
type OnboardingBillingContactInfo struct {
	// Name is the individual billing contact.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name,omitempty"`

	// Email is the billing contact's email. It receives billing
	// notifications, and invoices when invoiceEmails is empty.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=254
	// +kubebuilder:validation:XValidation:rule="self.matches('^[^@[:space:]]+@[^@[:space:]]+[.][^@[:space:]]+$')",message="must be an email address, such as jane@example.com"
	Email string `json:"email,omitempty"`

	// BusinessName is the legal entity that pays, printed on invoices.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=256
	BusinessName string `json:"businessName,omitempty"`

	// InvoiceEmails is where invoices are emailed. It allows one entry:
	// BillingAccount accepts up to 10, but Billing only ever sends to the
	// first, so a second would be silently ignored.
	//
	// +optional
	// +listType=set
	// +kubebuilder:validation:MaxItems=1
	// +kubebuilder:validation:items:MaxLength=254
	// +kubebuilder:validation:items:XValidation:rule="self.matches('^[^@[:space:]]+@[^@[:space:]]+[.][^@[:space:]]+$')",message="must be an email address, such as jane@example.com"
	InvoiceEmails []string `json:"invoiceEmails,omitempty"`

	// Address is the postal address invoices are issued to.
	//
	// +optional
	Address *PostalAddress `json:"address,omitempty"`
}

// PostalAddress is a postal address. It follows the address on the
// Organization and the BillingAccount.
type PostalAddress struct {
	// Country is the ISO 3166-1 alpha-2 country code, such as US or GB. It
	// is optional so a partly entered address can be saved, but the
	// Organization and the BillingAccount require it.
	//
	// TODO(#14): when an address is given, require its country on submit.
	//
	// +optional
	// +kubebuilder:validation:XValidation:rule="self.matches('^[A-Z]{2}$')",message="must be a two-letter country code, such as US or GB"
	Country string `json:"country,omitempty"`

	// Line1 is the first line of the street address.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Line1 string `json:"line1,omitempty"`

	// Line2 is the second line of the street address.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Line2 string `json:"line2,omitempty"`

	// City is the locality.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=128
	City string `json:"city,omitempty"`

	// Region is the state, province, or county.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=128
	Region string `json:"region,omitempty"`

	// PostalCode is the postal or zip code.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=32
	PostalCode string `json:"postalCode,omitempty"`
}

// OrganizationOnboardingStatus is the onboarding's progress, written only by
// CRM. Staff roles can read it but not change it, so clients can't set
// completion state. Fields may change once the action API is built.
//
// TODO(#14): record agreement acceptance: which agreement version, the
// authenticated User, and when. It belongs to the recipient, not to the
// reusable organization details, so a replacement recipient starts without
// it; it likely lives on the OnboardingInvitation (#10). The agreements and
// their versions are identified in #14.
type OrganizationOnboardingStatus struct {
	// Phase is where the onboarding is in its lifecycle. An onboarding with
	// no phase yet is a draft.
	//
	// TODO(#13, #14, #16): set by the send, redeem, submit, and cancel
	// actions, and by completion.
	//
	// +optional
	Phase OnboardingPhase `json:"phase,omitempty"`

	// CurrentInvitationRef is the onboarding's one current invitation. Only
	// this invitation can be redeemed or used; a resend or replacement moves
	// it, so older links stop working.
	//
	// TODO(#13): set by the send, resend, and replace actions.
	//
	// +optional
	CurrentInvitationRef *OnboardingInvitationReference `json:"currentInvitationRef,omitempty"`

	// InvoiceAuthorization records the staff User who chose Invoice billing
	// and when. It is the evidence that invoicing was authorized; a
	// customer can't grant it.
	//
	// TODO(#13): recorded by the send action from the authenticated staff
	// User when the payment arrangement is Invoice.
	//
	// +optional
	InvoiceAuthorization *InvoiceAuthorization `json:"invoiceAuthorization,omitempty"`

	// OrganizationRef is the organization this onboarding created. It stays
	// linked after completion; deleting the onboarding never deletes it.
	//
	// TODO(#15): set when CRM creates the organization.
	//
	// +optional
	OrganizationRef *OrganizationReference `json:"organizationRef,omitempty"`

	// BillingAccountRef is the BillingAccount this onboarding created, in the
	// organization's namespace. Deleting the onboarding never deletes it.
	//
	// TODO(#15): set when CRM creates the billing account.
	//
	// +optional
	BillingAccountRef *BillingAccountReference `json:"billingAccountRef,omitempty"`

	// CompletionTime is when the onboarding completed. It is kept as a
	// historical record even if the organization or billing changes later.
	//
	// TODO(#16): set when the onboarding completes.
	//
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions report progress and anything blocking it, such as billing
	// or organization readiness.
	//
	// TODO(#15, #16): define the condition types once the controller is
	// built.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the spec generation the status reflects.
	//
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// InvoiceAuthorization records who authorized Invoice billing and when.
type InvoiceAuthorization struct {
	// UserRef is the staff User who authorized invoicing.
	//
	// +kubebuilder:validation:Required
	UserRef UserReference `json:"userRef"`

	// Time is when invoicing was authorized.
	//
	// +kubebuilder:validation:Required
	Time metav1.Time `json:"time"`
}

// OrganizationOnboarding is the setup journey for one customer organization:
// the details sales prepares, the billing choices, and the progress toward
// completion. Resends and replacements of its invitation continue the same
// onboarding; only a new onboarding creates another organization. It is
// cluster-scoped and managed only by Milo staff.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Company",type=string,JSONPath=`.spec.companyRef.name`
// +kubebuilder:printcolumn:name="Opportunity",type=string,JSONPath=`.spec.opportunityRef.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Currency",type=string,JSONPath=`.spec.billing.currency`
// +kubebuilder:printcolumn:name="Payment",type=string,JSONPath=`.spec.billing.paymentArrangement`
// +kubebuilder:printcolumn:name="Organization",type=string,JSONPath=`.status.organizationRef.name`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type OrganizationOnboarding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationOnboardingSpec   `json:"spec"`
	Status OrganizationOnboardingStatus `json:"status,omitempty"`
}

// OrganizationOnboardingList contains a list of OrganizationOnboardings.
//
// +kubebuilder:object:root=true
type OrganizationOnboardingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrganizationOnboarding `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(GroupVersion, &OrganizationOnboarding{}, &OrganizationOnboardingList{})
		metav1.AddToGroupVersion(scheme, GroupVersion)
		return nil
	})
}
