package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocalWebserverExchangeCodeGetter", func() {
	var (
		codeAuthGetter  *LocalWebserverExchangeCodeGetter
		successRedirect string
		address         string
	)

	BeforeEach(func() {
		successRedirect = "https://localhost:60212/success"
		address = "localhost:14000"

		// Create a test server that simulates the authorization server

		time.Sleep(100 * time.Millisecond)
		codeAuthGetter = NewLocalWebserverExchangeCodeGetter(address, successRedirect)
	})

	AfterEach(func() {
		if codeAuthGetter.server != nil {
			codeAuthGetter.server.Close()
		}

		// make it extra sure the server is closed
		time.Sleep(100 * time.Millisecond)
	})

	Context("Start", func() {
		It("should start the server successfully", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := codeAuthGetter.Start(ctx)
			Expect(err).NotTo(HaveOccurred())

			resp, err := http.Get("http://" + address + readyCheckPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("GetExchangeCode", func() {
		It("should return the exchange code successfully", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := codeAuthGetter.Start(ctx)
			Expect(err).NotTo(HaveOccurred())
			defer GinkgoRecover()
			go func() {
				time.Sleep(1 * time.Second)
				_, err := http.Get("http://" + address + "/?code=abc123")
				Expect(err).To(HaveOccurred())
				var urlError *url.Error
				errors.As(err, &urlError)
				// checks that the redirect URL is correct
				Expect(urlError.URL).To(Equal(successRedirect))
			}()

			code, err := codeAuthGetter.GetExchangeCode(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal("abc123"))
		})

		It("should return an error when the context is canceled", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := codeAuthGetter.Start(ctx)
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel = context.WithCancel(context.Background())
			cancel()

			code, err := codeAuthGetter.GetExchangeCode(ctx)
			Expect(code).To(BeEmpty())
			Expect(err).To(MatchError(context.Canceled))
		})
	})

	Describe("codeGetterHandler", func() {
		var (
			respCh  chan *codeExchangeResult
			handler *codeGetterHandler
		)

		BeforeEach(func() {
			respCh = make(chan *codeExchangeResult, 1)
			handler = newCodeGetterHandler(respCh, successRedirect, "")
		})

		Context("waitForCode", func() {
			It("should send a success response when the code is present", func() {
				req, err := http.NewRequest(http.MethodGet, "/?code=abc123", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				handler.waitForCode(rr, req)

				result := <-respCh
				Expect(result.Code).To(Equal("abc123"))
				Expect(result.Err).NotTo(HaveOccurred())
				Expect(rr.Code).To(Equal(http.StatusFound))
				Expect(rr.Header().Get("Location")).To(Equal(successRedirect))
			})

			It("should send a failed response when an error is present", func() {
				req, err := http.NewRequest(http.MethodGet, "/?error=access_denied&error_description=User+denied+access", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				handler.waitForCode(rr, req)

				result := <-respCh
				Expect(result.Code).To(BeEmpty())
				Expect(result.Err).To(MatchError(fmt.Sprintf(authServerError, "access_denied", "User denied access")))
				Expect(rr.Code).To(Equal(http.StatusUnauthorized))
			})

			It("should return a 404 response for unknown paths", func() {
				req, err := http.NewRequest(http.MethodGet, "/unknown", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				handler.waitForCode(rr, req)

				Expect(rr.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("handleSuccess", func() {
			It("should return the code in the response and redirect", func() {
				req, err := http.NewRequest(http.MethodGet, "/?code=abc123", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := handler.handleSuccess(rr, req)

				Expect(result.Code).To(Equal("abc123"))
				Expect(result.Err).NotTo(HaveOccurred())
				Expect(rr.Code).To(Equal(http.StatusFound))
				Expect(rr.Header().Get("Location")).To(Equal(successRedirect))
			})
		})

		Context("handleFailed", func() {
			It("should return an error with the provided error details", func() {
				req, err := http.NewRequest(http.MethodGet, "/?error=access_denied&error_description=User+denied+access", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := handler.handleFailed(rr, req)

				Expect(result.Code).To(BeEmpty())
				Expect(result.Err).To(MatchError(fmt.Sprintf(authServerError, "access_denied", "User denied access")))
				Expect(rr.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("state validation", func() {
			It("should succeed when state matches expected state", func() {
				stateHandler := newCodeGetterHandler(respCh, successRedirect, "expected-state")
				req, err := http.NewRequest(http.MethodGet, "/?code=abc123&state=expected-state", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := stateHandler.handleSuccess(rr, req)

				Expect(result.Code).To(Equal("abc123"))
				Expect(result.Err).NotTo(HaveOccurred())
				Expect(rr.Code).To(Equal(http.StatusFound))
			})

			It("should fail when state does not match expected state", func() {
				stateHandler := newCodeGetterHandler(respCh, successRedirect, "expected-state")
				req, err := http.NewRequest(http.MethodGet, "/?code=abc123&state=wrong-state", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := stateHandler.handleSuccess(rr, req)

				Expect(result.Code).To(BeEmpty())
				Expect(result.Err).To(MatchError(stateMismatchError))
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("should fail when state is missing from callback but expected", func() {
				stateHandler := newCodeGetterHandler(respCh, successRedirect, "expected-state")
				req, err := http.NewRequest(http.MethodGet, "/?code=abc123", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := stateHandler.handleSuccess(rr, req)

				Expect(result.Code).To(BeEmpty())
				Expect(result.Err).To(MatchError(stateMismatchError))
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("should skip state validation when no expected state is set", func() {
				noStateHandler := newCodeGetterHandler(respCh, successRedirect, "")
				req, err := http.NewRequest(http.MethodGet, "/?code=abc123", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := noStateHandler.handleSuccess(rr, req)

				Expect(result.Code).To(Equal("abc123"))
				Expect(result.Err).NotTo(HaveOccurred())
			})

			It("should fail handleFailed when state does not match expected state", func() {
				stateHandler := newCodeGetterHandler(respCh, successRedirect, "expected-state")
				req, err := http.NewRequest(http.MethodGet, "/?error=access_denied&error_description=User+denied+access&state=wrong-state", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := stateHandler.handleFailed(rr, req)

				Expect(result.Code).To(BeEmpty())
				Expect(result.Err).To(MatchError(stateMismatchError))
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("should pass handleFailed when state matches expected state", func() {
				stateHandler := newCodeGetterHandler(respCh, successRedirect, "expected-state")
				req, err := http.NewRequest(http.MethodGet, "/?error=access_denied&error_description=User+denied+access&state=expected-state", nil)
				Expect(err).NotTo(HaveOccurred())

				rr := httptest.NewRecorder()
				result := stateHandler.handleFailed(rr, req)

				Expect(result.Code).To(BeEmpty())
				Expect(result.Err).To(MatchError(fmt.Sprintf(authServerError, "access_denied", "User denied access")))
				Expect(rr.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
