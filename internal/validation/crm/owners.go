// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// A new record is owned by the record it belongs to, which the defaulting
// webhook sets, so deleting the owner deletes it. These rules keep that
// ownership tree intact: no other owners on create, and no changes after.

var ownerReferencesPath = field.NewPath("metadata", "ownerReferences")

// validateOwnedOnlyBy rejects owner references other than the one to the
// record's owner. A second owner would keep the record alive after its owner
// is deleted, since the garbage collector waits for every owner to go.
func validateOwnedOnlyBy(refs []metav1.OwnerReference, kind, name string) field.ErrorList {
	var allErrs field.ErrorList
	for i, ref := range refs {
		if ref.APIVersion != crmv1alpha1.GroupVersion.String() || ref.Kind != kind || ref.Name != name {
			allErrs = append(allErrs, field.Forbidden(ownerReferencesPath.Index(i),
				"the record can only be owned by its "+kind+" "+name+", which CRM sets"))
		}
	}
	return allErrs
}

// validateOwnersUnchanged rejects any change to the owner references after
// create, so nobody can detach a record from its owner.
func validateOwnersUnchanged(oldRefs, newRefs []metav1.OwnerReference) field.ErrorList {
	if equality.Semantic.DeepEqual(oldRefs, newRefs) {
		return nil
	}
	return field.ErrorList{field.Forbidden(ownerReferencesPath, "owner references can't be changed; CRM sets them on create")}
}
