// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// InvitationPhase is where an OnboardingInvitation is in its lifecycle.
//
// +kubebuilder:validation:Enum=Pending;Sent;Redeemed;Expired;Revoked
type InvitationPhase string

const (
	// InvitationPhasePending is an invitation CRM created but hasn't
	// delivered yet.
	InvitationPhasePending InvitationPhase = "Pending"
	// InvitationPhaseSent is an invitation delivered to the recipient and not
	// yet redeemed.
	InvitationPhaseSent InvitationPhase = "Sent"
	// InvitationPhaseRedeemed is an invitation the recipient redeemed. They
	// can work on the onboarding while it is the current invitation.
	InvitationPhaseRedeemed InvitationPhase = "Redeemed"
	// InvitationPhaseExpired is an invitation whose link expired before it
	// was redeemed. The onboarding keeps its progress.
	InvitationPhaseExpired InvitationPhase = "Expired"
	// InvitationPhaseRevoked is an invitation that no longer works because it
	// was cancelled or replaced, or its onboarding was cancelled. The
	// Revoked condition says which.
	InvitationPhaseRevoked InvitationPhase = "Revoked"
)

// OnboardingInvitationSpec is who the invitation is for, captured when it was
// sent. CRM's send action fills it in when it creates the invitation; no
// client sets it, and none of it can change afterwards. Everything CRM does
// with the invitation is in status.
//
// TODO(#11): reject an onboarding or Person that doesn't exist, and a Person
// whose Company isn't the onboarding's Company.
type OnboardingInvitationSpec struct {
	// OnboardingRef is the onboarding the invitation is for. It can't be
	// changed. Resending or replacing an invitation continues the same
	// onboarding.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="onboardingRef can't be changed"
	OnboardingRef OrganizationOnboardingReference `json:"onboardingRef"`

	// PersonRef is the Person invited. It can't be changed; to invite someone
	// else, the onboarding's invitation is replaced.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="personRef can't be changed"
	PersonRef PersonReference `json:"personRef"`

	// RecipientEmail is the Person's email when the invitation was sent. CRM
	// copies it from the Person as it creates the invitation. It can't be
	// changed, so changing the Person later never redirects an invitation
	// already sent. Only a User whose verified email matches it can redeem
	// the invitation.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=254
	// +kubebuilder:validation:XValidation:rule="self.matches('^[^@[:space:]]+@[^@[:space:]]+[.][^@[:space:]]+$')",message="must be an email address, such as jane@example.com"
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="recipientEmail can't be changed"
	RecipientEmail string `json:"recipientEmail"`
}

// OnboardingInvitationStatus is what CRM has recorded about the invitation.
// Only CRM writes it; staff roles can read it but not change it, so clients
// can't set the inviter or claim a redemption.
//
// The invitation holds no link, credential, or anything that could be used
// as one. CRM keeps only a verifier for the link, in restricted storage
// outside this resource.
//
// TODO(#14): record agreement acceptance here or on the onboarding: which
// agreement version, the redeeming User, and when. It belongs to this
// recipient, so a replacement starts without it.
type OnboardingInvitationStatus struct {
	// Phase is where the invitation is in its lifecycle. The conditions give
	// the details, such as why it was revoked.
	//
	// TODO(#13, #14): set by the send, resend, replace, cancel, and redeem
	// actions, and by the invitation controller on delivery and expiry.
	//
	// +optional
	Phase InvitationPhase `json:"phase,omitempty"`

	// InviterRef is the staff User who sent the invitation, taken from the
	// authenticated request, never from the request body.
	//
	// TODO(#13): set by the send action.
	//
	// +optional
	InviterRef *UserReference `json:"inviterRef,omitempty"`

	// SentTime is when the invitation was sent.
	//
	// TODO(#13): set by the send action.
	//
	// +optional
	SentTime *metav1.Time `json:"sentTime,omitempty"`

	// ExpiresAt is when the invitation link stops working, seven days after
	// it was sent. An expired invitation doesn't erase the onboarding's
	// progress.
	//
	// TODO(#13): set by the send action. Whether a resend moves it or issues a
	// new invitation is decided with the action API (#12).
	//
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// RedeemedUserRef is the Milo User who redeemed the invitation: the
	// signed-in User whose verified email matched recipientEmail. While this
	// is the onboarding's current invitation, only this User can work on the
	// onboarding, and on completion this User becomes the organization's
	// first administrator.
	//
	// TODO(#14): set by the redeem action.
	//
	// +optional
	RedeemedUserRef *UserReference `json:"redeemedUserRef,omitempty"`

	// RedeemedTime is when the invitation was redeemed.
	//
	// TODO(#14): set by the redeem action.
	//
	// +optional
	RedeemedTime *metav1.Time `json:"redeemedTime,omitempty"`

	// Conditions report delivery, redemption, expiration, and revocation, for
	// example when the invitation is replaced or the onboarding is cancelled.
	//
	// TODO(#13, #14): define the condition types once the actions are built.
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

// OnboardingInvitation is one recipient's invitation to complete an
// OrganizationOnboarding. CRM creates it when staff send the onboarding; staff
// can view invitations but not create or change them. It is cluster-scoped and
// visible only to Milo staff.
//
// An onboarding can have many invitations over time, one per send or
// replacement, but only its status.currentInvitationRef can be used. List an
// onboarding's or a Person's invitations with a field selector on
// spec.onboardingRef.name or spec.personRef.name.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:selectablefield:JSONPath=`.spec.onboardingRef.name`
// +kubebuilder:selectablefield:JSONPath=`.spec.personRef.name`
// +kubebuilder:printcolumn:name="Onboarding",type=string,JSONPath=`.spec.onboardingRef.name`
// +kubebuilder:printcolumn:name="Person",type=string,JSONPath=`.spec.personRef.name`
// +kubebuilder:printcolumn:name="Recipient",type=string,JSONPath=`.spec.recipientEmail`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Expires",type=date,JSONPath=`.status.expiresAt`
// +kubebuilder:printcolumn:name="Redeemed By",type=string,JSONPath=`.status.redeemedUserRef.name`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type OnboardingInvitation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OnboardingInvitationSpec   `json:"spec"`
	Status OnboardingInvitationStatus `json:"status,omitempty"`
}

// OnboardingInvitationList contains a list of OnboardingInvitations.
//
// +kubebuilder:object:root=true
type OnboardingInvitationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OnboardingInvitation `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(GroupVersion, &OnboardingInvitation{}, &OnboardingInvitationList{})
		metav1.AddToGroupVersion(scheme, GroupVersion)
		return nil
	})
}
