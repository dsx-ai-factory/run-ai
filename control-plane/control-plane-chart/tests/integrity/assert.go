package integrity

import (
	"fmt"
	"sort"
	"strings"
)

type Violation struct {
	Reference Reference
	Reason    string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s container=%s field=%s: %s",
		v.Reference.Service, v.Reference.Container, v.Reference.Field, v.Reason)
}

func validateReferences(refs []Reference, p *Producers) []Violation {
	var violations []Violation
	for _, r := range refs {
		if r.Secret == "" {
			violations = append(violations, Violation{
				Reference: r,
				Reason:    "empty secret name -- broken chart wiring",
			})
			continue
		}
		if p.isExternal(r.Secret) {
			continue
		}
		if r.Key == "*" {
			if !p.hasSecret(r.Secret) {
				violations = append(violations, Violation{
					Reference: r,
					Reason:    fmt.Sprintf("no producer declares secret %q", r.Secret),
				})
			}
			continue
		}
		if !p.hasKey(r.Secret, r.Key) {
			violations = append(violations, Violation{
				Reference: r,
				Reason: fmt.Sprintf("secret %q has no key %q (declared: %s)",
					r.Secret, r.Key, sortedKeys(p, r.Secret)),
			})
		}
	}
	return violations
}

func sortedKeys(p *Producers, secret string) string {
	keys, ok := p.Keys[secret]
	if !ok {
		return "<none>"
	}
	out := make([]string, 0, len(keys))
	for k := range keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}
