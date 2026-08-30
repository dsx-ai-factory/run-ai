package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type OauthAccessResponse struct {
	Ok          bool                    `json:"ok"`
	AccessToken string                  `json:"access_token,omitempty"`
	Error       string                  `json:"error,omitempty"`
	Team        OauthAccessResponseTeam `json:"team,omitempty"`
}

type OauthAccessResponseTeam struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type DefaultOauthAccess struct {
	response OauthAccessResponse
}

func (oa *DefaultOauthAccess) SaveResponse(r *http.Request) {
	fmt.Println("Slack mock got new DefaultOauthAccess.SaveResponse request")
	res, err := parseBody[OauthAccessResponse](r)
	if err != nil {
		fmt.Printf("Slack mock ERROR!!! got new DefaultOauthAccess.SaveResponse failed to parse: %+v \n", err)
		return
	}
	fmt.Printf("Slack mock got new DefaultOauthAccess.SaveResponse response: %+v \n", res)
	oa.response = *res
}

func (oa *DefaultOauthAccess) Respond(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Slack mock got new DefaultOauthAccess.Respond request")

	err := errors.Join(assertNotEmpty(r.URL.Query().Get("client_id"), "ClientId empty"))
	err = errors.Join(err, assertEqual(slackAuthCode, r.URL.Query().Get("code"), "ClientId does not match"))
	err = errors.Join(err, assertContains(r.URL.Query().Get("redirect_uri"), "/api/v1/notification-channels/slack/auth-app", "redirect_uri not match validation"))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res, err := json.Marshal(oa.response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
