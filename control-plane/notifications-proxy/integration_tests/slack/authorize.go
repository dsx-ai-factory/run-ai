package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type AuthorizeResponse struct {
	Code        string `json:"code"`
	RedirectUri string `json:"redirect_uri"`
}

type DefaultAuthorize struct {
	response AuthorizeResponse
}

func (a *DefaultAuthorize) SaveResponse(r *http.Request) {
	fmt.Println("Slack mock got new DefaultAuthorize.SaveResponse request")
	res, err := parseBody[AuthorizeResponse](r)
	if err != nil {
		return
	}
	a.response = *res
}

func (a *DefaultAuthorize) Respond(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Slack-mock got new DefaultAuthorize.Respond request %s\n", r.Form.Encode())
	err := errors.Join(assertNotEmpty(r.URL.Query().Get("state"), "state is empty"))
	err = errors.Join(err, assertNotEmpty(r.URL.Query().Get("scope"), "scope is empty"))
	err = errors.Join(err, assertNotEmpty(r.URL.Query().Get("client_id"), "ClientId is empty"))
	err = errors.Join(err, assertContains(r.URL.Query().Get("redirect_uri"), "/api/v1/notification-channels/slack/auth-app", "redirect_uri not match validation"))
	if err != nil {
		log.Printf("Slack-mock bad request: %s %+v", err.Error(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params := url.Values{}
	params.Add("code", a.response.Code)
	params.Add("state", r.URL.Query().Get("state"))
	uri := fmt.Sprintf("%s?%s", a.response.RedirectUri, params.Encode())
	log.Printf("Slack-mock redirect to uri: %s", uri)
	http.Redirect(w, r, uri, http.StatusFound)
}
