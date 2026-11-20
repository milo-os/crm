// SPDX-License-Identifier: AGPL-3.0-only

package example

import (
	"maps"

	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	examplev1alpha1 "go.miloapis.com/crm/api/example/v1alpha1"
)

// eventRecorderName is the component the example timeline Events are reported
// by. The activity timeline shows it as the actor.
const eventRecorderName = "crm-controller"

// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// recordTimelineEvent records a timeline Event on obj with the given detail
// annotations, scoped to obj's organization so its members see it. A nil
// recorder (unit tests that don't care about events) records nothing.
func recordTimelineEvent(
	recorder record.EventRecorder,
	obj client.Object,
	detail map[string]string,
	eventType, reason, message string,
) {
	if recorder == nil {
		return
	}
	annotations := examplev1alpha1.OrganizationScopeAnnotations(obj.GetNamespace())
	if annotations == nil {
		annotations = map[string]string{}
	}
	maps.Copy(annotations, detail)
	recorder.AnnotatedEventf(obj, annotations, eventType, reason, "%s", message)
}

// isReady reports whether the Ready condition is True.
func isReady(c *metav1.Condition) bool {
	return c != nil && c.Status == metav1.ConditionTrue
}

// readyCondition returns a copy of the Resource's Ready condition, or nil.
func readyCondition(resource *examplev1alpha1.Resource) *metav1.Condition {
	c := apimeta.FindStatusCondition(resource.Status.Conditions, ConditionTypeReady)
	if c == nil {
		return nil
	}
	cp := *c
	return &cp
}

// recordReadyTransition records ResourceReady when the Resource became Ready
// since before. Call it only after the status write succeeded, so a conflict
// and retry can't record the same transition twice.
func recordReadyTransition(recorder record.EventRecorder, resource *examplev1alpha1.Resource, before *metav1.Condition) {
	if isReady(before) || !isReady(readyCondition(resource)) {
		return
	}
	var detail map[string]string
	if resource.Spec.Description != "" {
		detail = map[string]string{examplev1alpha1.EventAnnotationDescription: resource.Spec.Description}
	}
	recordTimelineEvent(recorder, resource, detail, corev1.EventTypeNormal,
		examplev1alpha1.EventReasonResourceReady, "Resource is ready.")
}
