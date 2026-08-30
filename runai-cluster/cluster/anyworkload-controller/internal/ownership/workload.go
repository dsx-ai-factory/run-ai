package ownership

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const submitUsage = "Submit"

type legacyWorkload struct {
	object *unstructured.Unstructured
}

func decodeLegacyWorkload(raw []byte) (*legacyWorkload, error) {
	object := &unstructured.Unstructured{}
	if err := json.Unmarshal(raw, &object.Object); err != nil {
		return nil, err
	}
	return &legacyWorkload{object: object}, nil
}

func (w *legacyWorkload) usage() string {
	usage, _, _ := unstructured.NestedString(w.object.Object, "spec", "usage")
	return usage
}

func (w *legacyWorkload) namespace() string {
	return w.object.GetNamespace()
}

func (w *legacyWorkload) ownerUID() string {
	ownerUID, _, _ := unstructured.NestedString(w.object.Object, "spec", "userId", "value")
	return ownerUID
}

func (w *legacyWorkload) restoreUserID(stored *legacyWorkload) error {
	userID, found, err := unstructured.NestedFieldCopy(stored.object.Object, "spec", "userId")
	if err != nil {
		return err
	}
	if !found {
		unstructured.RemoveNestedField(w.object.Object, "spec", "userId")
		return nil
	}
	return unstructured.SetNestedField(w.object.Object, userID, "spec", "userId")
}

func (w *legacyWorkload) marshal() ([]byte, error) {
	return json.Marshal(w.object.Object)
}
