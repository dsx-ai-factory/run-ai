package grafana

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const baseUrl = "https://grafana.com/api"

var ErrInstanceNotFound = errors.New("instance not found")

type LabsClient struct {
	c *resty.Client
}

func NewLabsClient(token string) *LabsClient {
	c := resty.New().
		SetBaseURL(baseUrl).
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json")
	return &LabsClient{c: c}
}

func (glc *LabsClient) CreateInstance(ctx context.Context, datasourceId string) (*Instance, error) {
	var result Instance
	resp, err := glc.c.R().
		SetContext(ctx).
		SetBody(CreateInstanceBody{
			Name: fmt.Sprintf("%s.grafana.net", datasourceId),
			Slug: datasourceId,
			Url:  fmt.Sprintf("https://%s.grafana.net", datasourceId),
		}).
		SetResult(&result).
		Post("/instances")
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create instance, status: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return &result, nil
}

func (glc *LabsClient) CreateAccessPolicy(ctx context.Context, policyScope string, instance *Instance) (*Policy, error) {
	policyName := generatePolicyName(instance, policyScope)

	var result Policy
	resp, err := glc.c.R().
		SetContext(ctx).
		SetQueryParam("region", instance.Region).
		SetBody(CreateAccessPolicyBody{
			Name:        policyName,
			DisplayName: policyName,
			Scopes:      []string{policyScope},
			Realms: []AccessPolicyRealm{
				{
					Type:       "stack",
					Identifier: fmt.Sprintf("%d", instance.Id),
				},
			},
		}).
		SetResult(&result).
		Post("/v1/accesspolicies")
	if err != nil {
		return nil, fmt.Errorf("failed to create access policy: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create access policy, status: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return &result, nil
}

func (glc *LabsClient) CreateToken(ctx context.Context, region string, policy *Policy) (*Token, error) {
	tokenName := generateTokenName(policy)

	var result Token
	resp, err := glc.c.R().
		SetContext(ctx).
		SetQueryParam("region", region).
		SetBody(CreateTokenBody{
			AccessPolicyId: policy.Id,
			Name:           tokenName,
			DisplayName:    tokenName,
		}).
		SetResult(&result).
		Post("/v1/tokens")
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create token, status: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return &result, nil
}

func (glc *LabsClient) DeleteInstance(ctx context.Context, datasourceId string) error {
	resp, err := glc.c.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/instances/%s", datasourceId))
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return ErrInstanceNotFound
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete instance, status: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (glc *LabsClient) DeleteAccessPolicy(ctx context.Context, accessPolicyId string, region string) error {
	resp, err := glc.c.R().
		SetContext(ctx).
		SetQueryParam("region", region).
		Delete(fmt.Sprintf("/v1/accesspolicies/%s", accessPolicyId))
	if err != nil {
		return fmt.Errorf("failed to delete access policy: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete access policy, status: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (glc *LabsClient) DeleteToken(ctx context.Context, tokenId string, region string) error {
	resp, err := glc.c.R().
		SetContext(ctx).
		SetQueryParam("region", region).
		Delete(fmt.Sprintf("/v1/tokens/%s", tokenId))
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete token, status: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func generateTokenName(policy *Policy) string {
	timestamp := fmt.Sprintf("%d", time.Now().UTC().UnixMilli())
	return fmt.Sprintf("%s-token-%s", policy.Name, timestamp)
}

func generatePolicyName(instance *Instance, policyScope string) string {
	scopeDisplayName := strings.ReplaceAll(policyScope, ":", "-")
	return fmt.Sprintf("%s-%s", instance.Slug, scopeDisplayName)
}
