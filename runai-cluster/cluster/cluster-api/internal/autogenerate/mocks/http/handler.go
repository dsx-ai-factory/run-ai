package http

import "net/http"

type MockUrlHandler struct {
	GetCalled  bool
	PostCalled bool
}

func (m *MockUrlHandler) Get(w http.ResponseWriter, r *http.Request) {
	m.GetCalled = true

	w.WriteHeader(http.StatusOK)
}

func (m *MockUrlHandler) Post(w http.ResponseWriter, r *http.Request) {
	m.PostCalled = true

	w.WriteHeader(http.StatusOK)
}
