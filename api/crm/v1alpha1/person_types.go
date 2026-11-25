// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PersonSpec is a sales contact at a Company. A Person is not a User: most
// contacts have no Milo account when sales adds them.
type PersonSpec struct {
	// GivenName is the Person's first name.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:XValidation:rule="!self.matches('^[[:space:]]*$')",message="must not be blank"
	GivenName string `json:"givenName"`

	// FamilyName is the Person's last name.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=128
	FamilyName string `json:"familyName,omitempty"`

	// Email is the Person's email address. Changing it does not redirect an
	// invitation already sent; invitations keep the address they were sent to.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=254
	// +kubebuilder:validation:XValidation:rule="self.matches('^[^@[:space:]]+@[^@[:space:]]+[.][^@[:space:]]+$')",message="must be an email address, such as jane@example.com"
	Email string `json:"email"`

	// CompanyRef is the Company the Person works for. It must exist. A
	// Company can have many People. It can't be changed: someone who moves to
	// another Company is a new Person there, so the Opportunities and
	// invitations that name them keep pointing at someone from their Company.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="companyRef can't be changed; someone who moves to another Company is a new Person"
	CompanyRef CompanyReference `json:"companyRef"`
}

// PersonStatus is what CRM has recorded about the Person.
//
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.userRef) || (has(self.userRef) && self.userRef == oldSelf.userRef)",message="userRef can't be changed or removed once set"
type PersonStatus struct {
	// UserRef links the Person to their Milo User, new or existing. CRM sets
	// it when the Person redeems an invitation, from the signed-in User making
	// that request, once their verified email matches the invitation. It is
	// status because only CRM writes it; staff roles can't update status.
	// The User must exist. Once set, it can't be changed or removed.
	//
	// +optional
	UserRef *UserReference `json:"userRef,omitempty"`
}

// Person is a sales contact at a Company, including people without a Milo
// account. It is cluster-scoped and managed only by Milo staff.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Given Name",type=string,JSONPath=`.spec.givenName`
// +kubebuilder:printcolumn:name="Family Name",type=string,JSONPath=`.spec.familyName`
// +kubebuilder:printcolumn:name="Email",type=string,JSONPath=`.spec.email`
// +kubebuilder:printcolumn:name="Company",type=string,JSONPath=`.spec.companyRef.name`
// +kubebuilder:printcolumn:name="User",type=string,JSONPath=`.status.userRef.name`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Person struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PersonSpec   `json:"spec"`
	Status PersonStatus `json:"status,omitempty"`
}

// PersonList contains a list of People.
//
// +kubebuilder:object:root=true
type PersonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Person `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(GroupVersion, &Person{}, &PersonList{})
		metav1.AddToGroupVersion(scheme, GroupVersion)
		return nil
	})
}
