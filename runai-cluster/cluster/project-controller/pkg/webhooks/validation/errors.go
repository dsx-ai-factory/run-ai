package validation

import (
	"errors"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// internalError marks a failure that is caused by the cluster/backend (e.g. an
// API server call failing) rather than by the validated object violating a
// rule. The handler maps these to an HTTP 500 admission error so transient
// outages surface as retryable failures instead of policy denials.
type internalError struct {
	err error
}

func (e *internalError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *internalError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func newInternalError(err error) *internalError {
	return &internalError{err: err}
}

// toAdmissionResponse maps a validation error to an admission response: backend
// failures become a retryable HTTP 500 error, while spec violations become a
// denial whose message reaches the caller verbatim. errors.As traverses the
// multierror chain, so an internal error anywhere in an aggregated result wins
// (an outage should be retried, not reported as a spec violation).
func toAdmissionResponse(err error) admission.Response {
	if _, ok := errors.AsType[*internalError](err); ok {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.Denied(err.Error())
}
