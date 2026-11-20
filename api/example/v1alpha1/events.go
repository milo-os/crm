// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import "strings"

// Event reasons for the example Resource activity timeline.
//
// The example controller records these as Kubernetes Events on the Resource,
// and the example ActivityPolicy (config/components/activity-policies) matches
// on them. An Event carries its detail in annotations with the
// EventAnnotationPrefix, not in its message: the policy renders the summary
// from those annotations, and the message is a plain fallback for kubectl.
//
// An Event reaches the owning organization's timeline only when it carries the
// scope annotations from OrganizationScopeAnnotations. Events without them stay
// on the platform timeline, which only staff can read.
//
// Emitters record each transition once. Requeues must not record a second
// event for a transition already recorded.
const (
	// EventReasonResourceReady is recorded on a Resource when it first
	// becomes Ready. Annotations: description.
	EventReasonResourceReady = "ResourceReady"
)

// EventAnnotationPrefix prefixes every example Event detail annotation.
const EventAnnotationPrefix = "event.example.miloapis.com/"

// EventAnnotationDescription is the Resource's description when the Event was
// recorded.
const EventAnnotationDescription = EventAnnotationPrefix + "description"

// Scope annotations that route an Event to an organization's timeline.
const (
	ScopeTypeAnnotation = "platform.miloapis.com/scope.type"
	ScopeNameAnnotation = "platform.miloapis.com/scope.name"
)

const organizationEventNamespacePrefix = "organization-"

// OrganizationScopeAnnotations returns the scope annotations that put an Event
// about a resource in namespace on its organization's activity timeline.
// Controllers write Events as service accounts, which Milo doesn't scope, so
// without these an Event only reaches the platform timeline. It returns nil
// for a namespace that isn't an organization's.
func OrganizationScopeAnnotations(namespace string) map[string]string {
	org, ok := strings.CutPrefix(namespace, organizationEventNamespacePrefix)
	if !ok || org == "" {
		return nil
	}
	return map[string]string{
		ScopeTypeAnnotation: "Organization",
		ScopeNameAnnotation: org,
	}
}
