package controller_utils

import (
	"fmt"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// SetControllerReference adds an owner as a controller owner reference to obj
func SetControllerReference(owner, resource client.Object, scheme *runtime.Scheme) error {
	if owner.GetNamespace() == "" || owner.GetNamespace() == resource.GetNamespace() {
		// Owner is cluster scoped, or both owner and obj are in the same namespace
		// either way, owner reference can be set
		return controllerutil.SetControllerReference(owner, resource, scheme)
	} else {
		// Owner is namespaced and obj is not in its namespace,
		// either cluster scoped, or in a different namespace,
		// so owner annotations are used instead
		return SetOwnerAnnotations(owner, []client.Object{resource})
	}
}

// SetOwnerReference adds an owner as a owner reference to obj
func SetOwnerReference(owner, resource client.Object, scheme *runtime.Scheme) error {
	if owner.GetNamespace() == "" || owner.GetNamespace() == resource.GetNamespace() {
		// Owner is cluster scoped, or both owner and obj are in the same namespace
		// either way, owner reference can be set
		return controllerutil.SetOwnerReference(owner, resource, scheme)
	} else {
		// Owner is namespaced and obj is not in its namespace,
		// either cluster scoped, or in a different namespace,
		// so owner annotations are used instead
		return SetOwnerAnnotations(owner, []client.Object{resource})
	}
}

// SetControllerReferenceForApplicableResources sets owner references for resources in the same namespace as the owner
func SetControllerReferenceForApplicableResources(owner client.Object, objects []client.Object, scheme *runtime.Scheme) error {
	for _, object := range objects {
		if err := SetControllerReference(owner, object, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference for object %s/%s: %w",
				object.GetNamespace(), object.GetName(), err)
		}
	}
	return nil
}

// SetOwnerReferenceForApplicableResources sets owner references for resources in the same namespace as the owner
func SetOwnerReferenceForApplicableResources(owner client.Object, objects []client.Object, scheme *runtime.Scheme) error {
	for _, object := range objects {
		if err := SetOwnerReference(owner, object, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference for object %s/%s: %w",
				object.GetNamespace(), object.GetName(), err)
		}
	}
	return nil
}

// SetOwnerAnnotations helps in adding 'NamespacedNameAnnotation' and 'TypeAnnotation' to object based on
// the values obtained from owner. The object gets the annotations from owner's namespace, name, group
// and kind. In other terms, object can be said to be the dependent having annotations from the owner.
// When a watch is set on the object, the annotations help to identify the owner and trigger reconciliation.
// Annotations are ALWAYS overwritten.
func SetOwnerAnnotations(owner client.Object, objects []client.Object) error {
	if owner.GetName() == "" {
		return fmt.Errorf("%T does not have a name, cannot call SetOwnerAnnotations", owner)
	}

	ownerGK := owner.GetObjectKind().GroupVersionKind().GroupKind()

	if ownerGK.Kind == "" {
		return fmt.Errorf("owner %s Kind not found, cannot call SetOwnerAnnotations", owner.GetName())
	}

	for _, object := range objects {
		annotations := object.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}

		annotations[runaiv1.NamespacedNameAnnotation] = fmt.Sprintf("%s/%s", owner.GetNamespace(), owner.GetName())
		annotations[runaiv1.TypeAnnotation] = ownerGK.String()

		object.SetAnnotations(annotations)
	}

	return nil
}

func IsOwnerOf(ownerGK metav1.GroupKind, ownerNamespace, ownerName string, objAnnotations map[string]string) bool {
	// Check if the object has owner annotations
	if objAnnotations == nil {
		return false
	}

	// Get the owner's namespaced name from annotations
	ownerNamespacedName, exists := objAnnotations[runaiv1.NamespacedNameAnnotation]
	if !exists {
		return false
	}

	// Get the owner's type from annotations
	ownerType, exists := objAnnotations[runaiv1.TypeAnnotation]
	if !exists {
		return false
	}

	// Check if the namespaced name matches
	expectedNamespacedName := fmt.Sprintf("%s/%s", ownerNamespace, ownerName)
	if ownerNamespacedName != expectedNamespacedName {
		return false
	}

	// Check if the type matches
	if ownerGK.String() != ownerType {
		return false
	}

	return true
}

func IsOwnedByMe(obj client.Object, ownerAPIVersion, ownerKind string) bool {
	if obj == nil {
		return false
	}

	for _, ref := range obj.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller &&
			ref.APIVersion == ownerAPIVersion &&
			ref.Kind == ownerKind {
			return true
		}
	}

	return false
}
