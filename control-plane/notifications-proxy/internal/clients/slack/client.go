package slack

import (
	"fmt"
	"net/url"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
	"github.com/run-ai/runai/runai-common-packages/config"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

const (
	slackAuthEndpoint      = "/oauth/v2/authorize"
	createManifestEndpoint = "/api/apps.manifest.create"
	oauthAccessEndpoint    = "/api/oauth.v2.access"
	//nolint:lll
	manifestFormat = `
{
  "display_information": {
    "name": "Nvidia Run.ai",
    "description": "The official NVIDIA Run:ai app for Slack keeps your team informed by sending real-time notifications",
    "background_color": "#000000",
	"long_description": "The official NVIDIA Run:ai app for Slack keeps your team informed by sending real-time notifications whenever a workload’s status changes from creating and pending to running, completed, and beyond. Each user can personalize which status change alerts they receive in their user settings, so you get the information that matters most to your workflow. With this app, your team can:\r\n• Get notifications for every workload status change\r\n• Fine-tune notification for workload status changes such as creating, pending, running, and completed, choosing only the updates relevant to you\r\n• Receive timely warnings if a workload is about to timeout"
  },
  "features": {
    "bot_user": {
      "display_name": "Nvidia Run.ai",
      "always_online": false
    }
  },
  "oauth_config": {
	"redirect_urls": [
	  "%s"
	],
    "scopes": {
      "bot": [
		"chat:write",
        "chat:write.public",
        "im:write",
        "users:read",
        "users:read.email",
        "channels:read"
	  ]
    }
  },
  "settings": {
    "socket_mode_enabled": false,
    "org_deploy_enabled": false,
	"token_rotation_enabled": false
  }
}`
	scopesForQuery = `chat:write,chat:write.public,im:write,users:read,users:read.email,channels:read`
)

type ExchangeResult struct {
	AccessToken string
	TeamName    string
}

type Client struct {
	httpClient        *resty.Client
	createManifestUrl string
	oauthAccessUrl    string
	slackAuthURL      string
	redirectUrl       string
}

type appsManifestCreateRequest struct {
	Manifest string `json:"manifest"`
}

func NewSlackClient(baseUrl string) *Client {
	conf := config.Get[*slack.SlackConfig]()
	return &Client{
		httpClient:        resty.New().SetTimeout(30 * time.Second),
		createManifestUrl: conf.BaseUrl + createManifestEndpoint,
		oauthAccessUrl:    conf.BaseUrl + oauthAccessEndpoint,
		slackAuthURL:      conf.BaseUrl + slackAuthEndpoint,
		redirectUrl:       getRedirectUrl(baseUrl),
	}
}

func (sc *Client) CreateApp(accessToken string) (*slack.SlackCredentials, error) {
	manifest := fmt.Sprintf(manifestFormat, sc.redirectUrl)
	return sc.sendCreateAppRequest(accessToken, manifest)
}

func (sc *Client) ExchangeCode(code, clientId, clientSecret string) (*slack.ExchangeCodeResult, error) {
	var res oauthAccessResponse
	r, err := sc.httpClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     clientId,
			"client_secret": clientSecret,
			"code":          code,
			"redirect_uri":  sc.redirectUrl,
		}).
		SetResult(&res).
		Post(sc.oauthAccessUrl)
	if err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, fmt.Errorf("failed to exchange slack code: %s", string(r.Body()))
	}

	return &slack.ExchangeCodeResult{
		AccessToken: res.AccessToken,
		TeamName:    res.Team.Name,
	}, nil
}

func (sc *Client) GetAuthUrl(state, clientId string) string {
	params := url.Values{}
	params.Add("client_id", clientId)
	params.Add("scope", scopesForQuery)
	params.Add("redirect_uri", sc.redirectUrl)
	params.Add("state", state)

	fullURL := fmt.Sprintf("%s?%s", sc.slackAuthURL, params.Encode())
	return fullURL
}

func (sc *Client) sendCreateAppRequest(accessToken string, manifest string) (*slack.SlackCredentials, error) {
	createRequest := &appsManifestCreateRequest{
		Manifest: manifest,
	}

	var createResponse CreateAppResponse
	r, err := sc.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		SetBody(createRequest).
		SetResult(&createResponse).
		Post(sc.createManifestUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Slack app from manifest")
	}

	if !createResponse.Ok {
		return nil, fmt.Errorf("failed to create Slack app from manifest: %s", string(r.Body()))
	}

	return &slack.SlackCredentials{
		AppId:        createResponse.AppId,
		ClientId:     createResponse.Credentials.ClientId,
		ClientSecret: createResponse.Credentials.ClientSecret,
	}, nil
}

func getRedirectUrl(baseUrl string) string {
	return fmt.Sprintf("%s/api/v1/notification-channels/slack/auth-app", baseUrl)
}
