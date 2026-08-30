package topology

import workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

// ActiveAnnotationValue returns the effective value of the named annotation.
// Later entries override earlier entries, and an excluded entry removes the
// value until a later active entry with the same name reactivates it.
func ActiveAnnotationValue(annotations []workloadsApi.Annotation, name string) (string, bool) {
	var value string
	found := false

	for _, annotation := range annotations {
		if annotation.GetName() != name {
			continue
		}

		if annotation.GetExclude() {
			value = ""
			found = false
			continue
		}

		value = annotation.GetValue()
		found = true
	}

	return value, found
}
