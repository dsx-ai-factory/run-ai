package services

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// ruleIssues accumulates validation findings discovered while walking a
// policy's rules and while resolving its breakdown.
//
// Set maxErrors > 0 to cap the number of accumulated issues; once the cap
// is reached, every add* method becomes a no-op and full() returns true so
// callers can short-circuit further work. maxErrors of 0 means unlimited.
//
// The same collector is shared across GeneratePolicyBreakdown and
// validateRules so that:
//   - issues from breakdown generation (rules whose field/specSelector
//     doesn't resolve in the mapping) and validation (everything else) are
//     reported uniformly to the caller, and
//   - the maxErrors cap is honored across both phases.
type ruleIssues struct {
	items     []*api.Error
	maxErrors int
}

// newRuleIssues creates a collector with the given maxErrors cap.
// Pass 0 for unlimited; pass 1 to recover the legacy "fail on first issue"
// behavior used by create/update/get.
func newRuleIssues(maxErrors int) *ruleIssues {
	return &ruleIssues{
		maxErrors: maxErrors,
	}
}

// full reports whether the collector has reached its maxErrors cap. Callers
// (validateRules, the breakdown generator) check this between phases to
// stop processing once the cap is hit, since further add calls would just
// be discarded anyway.
func (r *ruleIssues) full() bool {
	return r.maxErrors > 0 && len(r.items) >= r.maxErrors
}

// add appends a non-nil *api.Error produced by an existing helper. No-op
// when err is nil or the collector is already full().
func (r *ruleIssues) add(err *api.Error) {
	if err == nil || r.full() {
		return
	}
	r.items = append(r.items, err)
}

// addf creates a new validation error from the formatted message and appends
// it. This is the LogBrokenRule analog: instead of writing to the logger, the
// finding is captured in the slice. No-op when the collector is full().
func (r *ruleIssues) addf(format string, args ...any) {
	if r.full() {
		return
	}
	r.items = append(r.items, api.NewValidationError(fmt.Sprintf(format, args...)))
}

// addWithDetails appends a validation error that carries a Details payload
// (e.g. "valid specSelectors: …"). No-op when the collector is full().
func (r *ruleIssues) addWithDetails(message, details string) {
	if r.full() {
		return
	}
	r.items = append(r.items, validationErrorWithDetails(message, details))
}

// toError returns the first issue collected, or nil when none were
// collected. Callers that want every issue (planned for a future "show all
// findings" API) should iterate all() directly.
func (r *ruleIssues) toError() *api.Error {
	if len(r.items) == 0 {
		return nil
	}
	return r.items[0]
}

// all returns every collected issue in insertion order. Currently used by
// tests to assert accumulation across phases; reserved for a future API
// that surfaces every finding to the user instead of just the first.
func (r *ruleIssues) all() []*api.Error {
	return r.items
}
