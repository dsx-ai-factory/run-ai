package common

import "reflect"

// DiffMaps reports field-level changes between two flat snapshots as
// path -> {before, after}. A key present on only one side is reported with the
// missing side as nil. Intended for write tools that show a before/after diff.
func DiffMaps(before, after map[string]any) map[string]any {
	const keyBefore, keyAfter = "before", "after"
	diff := map[string]any{}
	for k, bv := range before {
		av, ok := after[k]
		if !ok {
			diff[k] = map[string]any{keyBefore: bv, keyAfter: nil}
		} else if !equalValues(bv, av) {
			diff[k] = map[string]any{keyBefore: bv, keyAfter: av}
		}
	}
	for k, av := range after {
		if _, ok := before[k]; !ok {
			diff[k] = map[string]any{keyBefore: nil, keyAfter: av}
		}
	}
	return diff
}

func equalValues(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
