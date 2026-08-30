package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	readyCheckPath      = "/ready-internal-Very-uniq"
	serverStartError    = "could not start a server for broswer flow: %v"
	authServerError     = "authorization error from server: %s %s"
	stateMismatchError  = "authentication session is invalid"
	errorKey            = "error"
	codeKey             = "code"
	stateKey            = "state"
	defaultEndpoint     = "/"
	errorDescriptionMsg = "error_description"
	defaultStopTime     = 200 * time.Millisecond
	apiGetSleepTime     = 50 * time.Millisecond
	TextHtmlType        = "text/html"
)

type LocalWebserverExchangeCodeGetter struct {
	address                  string
	customSuccessRedirectUrl string
	expectedState            string
	tokenChan                chan *codeExchangeResult
	server                   *http.Server
}

func NewLocalWebserverExchangeCodeGetter(address, customSuccessRedirectUrl string) *LocalWebserverExchangeCodeGetter {
	return &LocalWebserverExchangeCodeGetter{
		address:                  address,
		customSuccessRedirectUrl: customSuccessRedirectUrl,
	}
}

func (l *LocalWebserverExchangeCodeGetter) SetExpectedState(state string) {
	l.expectedState = state
}

func (l *LocalWebserverExchangeCodeGetter) stop() {
	// give it a grace period to shutdown
	graceCtx, cancel := context.WithTimeout(context.Background(), defaultStopTime)
	defer cancel()
	err := l.server.Shutdown(graceCtx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		// force to close the server
		_ = l.server.Close()
	}
}

func (l *LocalWebserverExchangeCodeGetter) Start(ctx context.Context) error {
	return l.start(ctx)
}

func (l *LocalWebserverExchangeCodeGetter) start(ctx context.Context) error {
	mux := http.NewServeMux()
	l.tokenChan = make(chan *codeExchangeResult)
	handler := newCodeGetterHandler(l.tokenChan, l.customSuccessRedirectUrl, l.expectedState)

	// Create a server with a timeout
	l.server = &http.Server{ //nolint:gosec // this is for internal call, no security risk here.
		Addr:    l.address,
		Handler: mux,
	}
	mux.HandleFunc(readyCheckPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	errorChan := make(chan error)
	mux.HandleFunc(defaultEndpoint, handler.waitForCode)
	// add handler for ready check
	go func() {
		err := l.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorChan <- fmt.Errorf(serverStartError, err)
		}
	}()

	// check the server is ready
	readyUrl := &url.URL{
		Scheme: Http,
		Host:   l.address,
		Path:   readyCheckPath,
	}
	if err := waitForServerReady(ctx, readyUrl, errorChan); err != nil {
		l.stop()
		return err
	}
	return nil
}

func (l *LocalWebserverExchangeCodeGetter) GetExchangeCode(ctx context.Context) (string, error) {
	defer l.stop()
	for {
		select {
		case result := <-l.tokenChan:
			return result.Code, result.Err
		case <-ctx.Done():
			return "", context.Canceled
		default:
			time.Sleep(apiGetSleepTime)
		}
	}
}

type codeGetterHandler struct {
	respCh          chan<- *codeExchangeResult // channel to send a response to
	onceRespCh      sync.Once                  // ensure send once
	SuccessRedirect string
	expectedState   string
}

func newCodeGetterHandler(respCh chan<- *codeExchangeResult, successRedirect string, expectedState string) *codeGetterHandler {
	return &codeGetterHandler{
		respCh:          respCh,
		SuccessRedirect: successRedirect,
		expectedState:   expectedState,
	}
}

func (h *codeGetterHandler) validateState(r *http.Request) *codeExchangeResult {
	if h.expectedState == "" {
		return nil
	}
	callbackState := r.URL.Query().Get(stateKey)
	if callbackState == h.expectedState {
		return nil
	}
	return &codeExchangeResult{Err: errors.New(stateMismatchError)}
}

func (h *codeGetterHandler) waitForCode(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	switch {
	case r.Method == Get && q.Get(errorKey) != "":
		h.onceRespCh.Do(func() {
			h.respCh <- h.handleFailed(w, r)
		})
	case r.Method == Get && q.Get(codeKey) != "":
		h.onceRespCh.Do(func() {
			h.respCh <- h.handleSuccess(w, r)
		})
	default:
		http.NotFound(w, r)
	}
}

func (h *codeGetterHandler) handleSuccess(w http.ResponseWriter, r *http.Request) *codeExchangeResult {
	if result := h.validateState(r); result != nil {
		http.Error(w, result.Err.Error(), http.StatusBadRequest)
		return result
	}

	code := r.URL.Query().Get(codeKey)
	http.Redirect(w, r, h.SuccessRedirect, http.StatusFound)
	return &codeExchangeResult{Code: code}
}

func (h *codeGetterHandler) handleFailed(w http.ResponseWriter, r *http.Request) *codeExchangeResult {
	if result := h.validateState(r); result != nil {
		http.Error(w, result.Err.Error(), http.StatusBadRequest)
		return result
	}

	q := r.URL.Query()
	errorCode, errorDescription := q.Get(errorKey), q.Get(errorDescriptionMsg)
	err := fmt.Errorf(authServerError, errorCode, errorDescription)
	http.Error(w, err.Error(), http.StatusUnauthorized)
	return &codeExchangeResult{Err: err}
}

func waitForServerReady(ctx context.Context, readyUrl *url.URL, webserverErrChan chan error) error {
	ticker := time.NewTicker(checkServerInterval)
	defer ticker.Stop()
	for {
		select {
		case err := <-webserverErrChan:
			return err
		case <-ctx.Done():
			return errors.New(serverNotReady)
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, readyUrl.String(), nil)
			if err != nil {
				continue
			}
			resp, _ := http.DefaultClient.Do(req)
			if resp != nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		}
	}
}
