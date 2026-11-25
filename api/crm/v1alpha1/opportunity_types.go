// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// OpportunityStage is where an Opportunity is in the sales process. Staff
// move it by hand; completing onboarding never changes it, and onboarding can
// start at any stage.
//
// +kubebuilder:validation:Enum=Prospecting;Qualification;Proposal;Negotiation;ClosedWon;ClosedLost
type OpportunityStage string

const (
	// OpportunityStageProspecting is a potential sale sales has identified
	// but not yet qualified.
	OpportunityStageProspecting OpportunityStage = "Prospecting"
	// OpportunityStageQualification is a sale sales is confirming is a real
	// fit: need, budget, and decision makers.
	OpportunityStageQualification OpportunityStage = "Qualification"
	// OpportunityStageProposal is a sale with an offer sent to the customer.
	OpportunityStageProposal OpportunityStage = "Proposal"
	// OpportunityStageNegotiation is a sale whose terms are being agreed.
	OpportunityStageNegotiation OpportunityStage = "Negotiation"
	// OpportunityStageClosedWon is a sale the customer agreed to.
	OpportunityStageClosedWon OpportunityStage = "ClosedWon"
	// OpportunityStageClosedLost is a sale that will not happen.
	OpportunityStageClosedLost OpportunityStage = "ClosedLost"
)

// OpportunitySpec is a potential sale with a Company.
type OpportunitySpec struct {
	// DisplayName is the Opportunity's human-readable name.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:XValidation:rule="!self.matches('^[[:space:]]*$')",message="must not be blank"
	DisplayName string `json:"displayName"`

	// Description is a short summary of what the sale is about, such as what
	// the customer wants to buy. Keep it current by editing it in place.
	// Interactions and follow-ups belong in ClusterNotes, not here: see
	// Opportunity.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=2000
	Description string `json:"description,omitempty"`

	// CompanyRef is the Company the sale is with. It must exist. A Company
	// can have many Opportunities. It can't be changed: a sale to another
	// Company is a new Opportunity, so the onboardings that name this one
	// stay with its Company.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="companyRef can't be changed; a sale to another Company is a new Opportunity"
	CompanyRef CompanyReference `json:"companyRef"`

	// PersonRefs are the People involved in the sale. Each must exist and
	// work for the Opportunity's Company.
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=50
	PersonRefs []PersonReference `json:"personRefs,omitempty"`

	// OwnerRef is the staff Milo User who owns the sale. The User must exist.
	// The owner is notified when onboarding started from this Opportunity
	// completes.
	//
	// +kubebuilder:validation:Required
	OwnerRef UserReference `json:"ownerRef"`

	// Stage is where the sale is in the sales process.
	//
	// +optional
	// +kubebuilder:default=Prospecting
	Stage OpportunityStage `json:"stage,omitempty"`
}

// Opportunity is a potential sale with a Company that sales tracks through
// its stages. It is cluster-scoped and managed only by Milo staff.
//
// Track the history of a sale, such as calls, meetings, and follow-ups, with
// ClusterNotes (notes.miloapis.com), the Milo platform primitive for notes on
// cluster-scoped resources. Set the note's subjectRef to the Opportunity
// (apiGroup crm.miloapis.com, kind Opportunity, and its name) and use
// nextAction, nextActionTime, and followUp to track what happens next. Milo
// deletes a note when its subject is deleted. CRM resources have no notes
// fields of their own.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Company",type=string,JSONPath=`.spec.companyRef.name`
// +kubebuilder:printcolumn:name="Stage",type=string,JSONPath=`.spec.stage`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.ownerRef.name`
// +kubebuilder:printcolumn:name="People",type=string,JSONPath=`.spec.personRefs[*].name`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Opportunity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec OpportunitySpec `json:"spec"`
}

// OpportunityList contains a list of Opportunities.
//
// +kubebuilder:object:root=true
type OpportunityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Opportunity `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(GroupVersion, &Opportunity{}, &OpportunityList{})
		metav1.AddToGroupVersion(scheme, GroupVersion)
		return nil
	})
}
