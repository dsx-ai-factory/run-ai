package reconcile

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

// StepResult is a helper type to encapsulate the result of a reconcile step.
// It is used to return the result of a reconcile step and to short-circuit the reconcile flow.
type StepResult struct {
	result            ctrl.Result
	Err               error
	continueReconcile bool
}

func (r StepResult) Result() (ctrl.Result, error) {
	return r.result, r.Err
}

func (r StepResult) HasError() bool {
	return r.Err != nil
}

func (r StepResult) ShortCircuitReconcileFlow() bool {
	return !r.continueReconcile
}

func Continue() StepResult {
	return StepResult{
		result:            ctrl.Result{},
		continueReconcile: true,
	}
}

func WithError(err error) StepResult {
	return StepResult{
		result:            ctrl.Result{},
		Err:               err,
		continueReconcile: false,
	}
}

func DoNotRequeue() StepResult {
	return StepResult{
		continueReconcile: false,
		result:            ctrl.Result{},
	}
}
