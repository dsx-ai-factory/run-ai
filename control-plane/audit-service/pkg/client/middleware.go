package client

import (
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

func Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Note: event when writer is not initialized, we want to provide the API with a valid audit log object
		ctx := r.Context()
		al := NewAuditLog()
		ctx = CtxWithAudit(ctx, al)

		if !IsInitialized() {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		rwc := NewResponseWriterCapture(w)
		al.WithRequest(r).WithCtx(ctx)

		defer doAudit(rwc, r, al)()

		next.ServeHTTP(rwc, r.WithContext(ctx))
	}
}

func doAudit(w *ResponseStatusCodeCapture, r *http.Request, al *AuditLog) func() {
	return func() {
		err := al.WithStatusCode(w.StatusCode).Write()
		if err != nil {
			logging.Errorf("Failed sending audit for URL %q: %v", r.URL.String(), err)
		} else {
			logging.Debugf("Successfully sent audit to service for URL %q", r.URL.String())
		}
	}
}
