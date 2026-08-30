package actions

import (
	"sort"
	"strings"
)

// MapToValues renders a map of Helm --set overrides into a sorted key=value
// slice so the resulting helm command is deterministic.
func MapToValues(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

// MergeValues combines our default --set overrides with user-provided custom
// values. When a key appears in both, the user's value replaces the default in
// place and only that single entry is emitted (Helm honors the last --set, but
// emitting both is confusing). New user keys are appended in sorted order.
func MergeValues(defaults []string, custom map[string]string) []string {
	out := make([]string, 0, len(defaults)+len(custom))
	applied := make(map[string]bool, len(custom))
	for _, d := range defaults {
		if i := strings.IndexByte(d, '='); i >= 0 {
			if v, ok := custom[d[:i]]; ok {
				out = append(out, d[:i]+"="+v)
				applied[d[:i]] = true
				continue
			}
		}
		out = append(out, d)
	}
	extra := make([]string, 0, len(custom))
	for k, v := range custom {
		if !applied[k] {
			extra = append(extra, k+"="+v)
		}
	}
	sort.Strings(extra)
	return append(out, extra...)
}
