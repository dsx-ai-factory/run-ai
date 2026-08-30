package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

type SlackMockServer struct {
	endPointsValidators map[string]EndpointValidatorer
	server              *httptest.Server
	authCode            string
}

type EndpointValidatorer interface {
	Respond(w http.ResponseWriter, r *http.Request)
	SaveResponse(r *http.Request)
}

func NewSlackMockServer(authCode string) *SlackMockServer {
	mux := http.NewServeMux()
	sms := &SlackMockServer{
		authCode: authCode,
		endPointsValidators: map[string]EndpointValidatorer{
			"/api/oauth.v2.access": &DefaultOauthAccess{},
			"/oauth/v2/authorize":  &DefaultAuthorize{},
		},
	}

	mux.HandleFunc("/api/oauth.v2.access", func(w http.ResponseWriter, r *http.Request) {
		sms.endPointsValidators["/api/oauth.v2.access"].Respond(w, r)
	})

	mux.HandleFunc("/save-oauth-access-response", func(w http.ResponseWriter, r *http.Request) {
		sms.endPointsValidators["/api/oauth.v2.access"].SaveResponse(r)
	})

	mux.HandleFunc("/oauth/v2/authorize", func(w http.ResponseWriter, r *http.Request) {
		sms.endPointsValidators["/oauth/v2/authorize"].Respond(w, r)
	})

	mux.HandleFunc("/save-authorize-response", func(w http.ResponseWriter, r *http.Request) {
		sms.endPointsValidators["/oauth/v2/authorize"].SaveResponse(r)
	})

	sms.server = &httptest.Server{
		URL: ":8081",
		Config: &http.Server{
			Addr:    ":8081",
			Handler: mux,
		},
	}

	return sms
}

func (s *SlackMockServer) Start() {
	if err := s.server.Config.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Mock server failed: %v", err)
	}
	fmt.Println("Mock server listening on", "localhost:8081")
}

func (s *SlackMockServer) Close() {
	s.server.Close()
}

func parseBody[T any](request *http.Request) (*T, error) {
	t := new(T)
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func assertNotEmpty(val, message string) error {
	if val == "" {
		return fmt.Errorf("%s: %s", message, val)
	}
	return nil
}

func assertEqual(expected, actual, message string) error {
	if expected != actual {
		return fmt.Errorf("%s, expected: %s actual %s", message, expected, actual)
	}
	return nil
}

func assertContains(str, substr, message string) error {
	if !strings.Contains(str, substr) {
		return fmt.Errorf("%s, str: %s substr %s", message, str, substr)
	}
	return nil
}
