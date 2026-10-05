/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	catalogConfig "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// ngcClient implements the NGCClient interface
type ngcClient struct {
	httpClient *http.Client
}

// NewNGCClient creates a new NGC client instance
func NewNGCClient() *ngcClient {
	ngcTimeout := time.Duration(config.Get[*catalogConfig.NGC]().RequestTimeoutSec) * time.Second

	return &ngcClient{
		httpClient: &http.Client{Timeout: ngcTimeout},
	}
}

const (
	ngcImagesCacheKey     = "ngc_images"
	ngcTagsCacheKeyPrefix = "ngc_image_tags"
	ngcImagesPageSize     = 500
)

// NGC API types
// These types are from the NGC API.
type catalogResponse struct {
	ResultTotal     int32                  `json:"resultTotal,omitempty"`
	ResultPageTotal int32                  `json:"resultPageTotal,omitempty"`
	Params          map[string]interface{} `json:"params,omitempty"`
	Results         []catalogResultsInner  `json:"results,omitempty"`
}
type catalogResultsInner struct {
	TotalCount int32             `json:"totalCount,omitempty"`
	Resources  []catalogResource `json:"resources,omitempty"`
}
type catalogResource struct {
	OrgName      string `json:"orgName,omitempty"`
	ResourceId   string `json:"resourceId,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Name         string `json:"name,omitempty"`
	ResourceType string `json:"resourceType,omitempty"`
	Description  string `json:"description,omitempty"`
	IsPublic     bool   `json:"isPublic,omitempty"`
}
type imageTagsResponse struct {
	Images []imageTag `json:"images,omitempty"`
}
type imageTag struct {
	Tag string `json:"tag,omitempty"`
}

// Cache types
type tagsCache struct {
	TagsByImage map[string]api.ImageTags
}
type imagesCache struct {
	Images []string
}

// GetImageTags retrieves image tags from NGC API with caching
func (c *ngcClient) GetImageTags(ctx context.Context, registryUrl string, imageName string, apiKey string) (api.ImageTags, error) {
	cacheKey := fmt.Sprintf("%s_%s_%s", ngcTagsCacheKeyPrefix, registryUrl, imageName)
	cachedResult, err := cache.Get[tagsCache](ctx, cacheKey)
	if err == nil && cachedResult != nil {
		return cachedResult.TagsByImage[imageName], nil
	}

	tags, err := c.sendNgcTagsRequest(ctx, imageName, apiKey)
	if err != nil {
		return api.ImageTags{}, err
	}

	tagsResponse := api.ImageTags{
		Name: imageName,
		Tags: tags,
	}

	err = cache.Set[tagsCache](ctx, cacheKey, tagsCache{
		TagsByImage: map[string]api.ImageTags{
			imageName: tagsResponse,
		},
	})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set NGC tags cache: %v", err)
	}

	return tagsResponse, nil
}

// GetImages retrieves matching NGC images with caching
func (c *ngcClient) GetImages(ctx context.Context, registryUrl string, apiKey string, search string, isNgcCatalog bool) ([]string, error) {
	cacheKey := buildImagesCacheKey(registryUrl, search)
	cachedResult, err := cache.Get[imagesCache](ctx, cacheKey)
	if err == nil && cachedResult != nil {
		return cachedResult.Images, nil
	}

	images, err := c.sendNgcImagesRequest(ctx, registryUrl, apiKey, search, isNgcCatalog)
	if err != nil {
		return []string{}, err
	}

	err = cache.Set[imagesCache](ctx, cacheKey, imagesCache{
		Images: images,
	})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set NGC images cache: %v", err)
	}

	return images, nil
}

func buildImagesCacheKey(registryUrl string, search string) string {
	return fmt.Sprintf("%s_%s_%s", ngcImagesCacheKey, registryUrl, strings.TrimSpace(search))
}

func (c *ngcClient) sendNgcImagesRequest(ctx context.Context, registryUrl string, apiKey string, search string, isNgcCatalog bool) ([]string, error) {
	query := map[string]interface{}{
		"pageSize": ngcImagesPageSize,
		"page":     0,
	}

	filters := make([]map[string]interface{}, 0)

	// We filter the images by isPublic and guestAccess only for NGC catalog, otherwise it's user private registry, no need to filter.
	if isNgcCatalog {
		filters = append(filters, map[string]interface{}{
			"field":    "isPublic",
			"value":    "true",
			"operator": "EQ",
		})

		// If no api key provided, we query only the public and guest access images.
		if apiKey == "" {
			filters = append(filters, map[string]interface{}{
				"field":    "guestAccess",
				"value":    "true",
				"operator": "EQ",
			})
		}
	}

	search = strings.TrimSpace(search)
	if search != "" {
		// NGC's free-text query parser does not match image identifiers containing
		// slashes (for example, "nvidia/pytorch"). Use its resourceId prefix
		// filter for image identifiers and free-text search for all other input.
		if strings.Contains(search, "/") {
			resourceID := strings.TrimPrefix(search, constants.NgcImagePrefix)
			if !strings.HasSuffix(resourceID, "*") {
				resourceID += "*"
			}
			filters = append(filters, map[string]interface{}{
				"field": "resourceId",
				"value": resourceID,
			})
		} else {
			query["query"] = search
		}
	}

	if len(filters) > 0 {
		query["filters"] = filters
	}

	imageSet := datastructs.NewSet[string]()
	for page := 0; ; page++ {
		query["page"] = page
		queryBytes, err := json.Marshal(query)
		if err != nil {
			return nil, fmt.Errorf("failed to encode NGC images query: %v", err)
		}
		ngcFullUrl := fmt.Sprintf("%s?q=%s", registryUrl, url.QueryEscape(string(queryBytes)))

		body, _, err := c.ngcGetRequest(ctx, ngcFullUrl, apiKey)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NGC API request failed. Error: %v", err)
			return []string{}, err
		}

		var ngcResp catalogResponse
		if err := json.Unmarshal(body, &ngcResp); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NGC API response parsing failed. Error: %v", err)
			return []string{}, fmt.Errorf("failed to parse NGC API response: %v", err)
		}
		for _, result := range ngcResp.Results {
			for _, resource := range result.Resources {
				imageSet.Add(resource.ResourceId)
			}
		}

		if page+1 >= int(ngcResp.ResultPageTotal) {
			break
		}
	}

	images := imageSet.Keys()
	return images, nil
}

func (c *ngcClient) sendNgcTagsRequest(ctx context.Context, imageName string, apiKey string) ([]string, error) {
	var ngcFullUrl string
	var err error

	// Use authenticated URL if API key is provided, otherwise use guest URL
	if apiKey != "" {
		ngcFullUrl, err = c.buildNgcAuthenticatedTagsUrl(imageName)
	} else {
		ngcFullUrl, err = c.buildNgcGuestTagsUrl(imageName)
	}

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to build NGC tags URL: %v", err)
		return []string{}, err
	}

	body, _, err := c.ngcGetRequest(ctx, ngcFullUrl, apiKey)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NGC API request failed. Error: %v", err)
		return []string{}, err
	}

	tags, err := c.parseTagsFromResponse(body)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse NGC tags response: %v", err)
		return []string{}, err
	}

	return tags, nil
}

func (c *ngcClient) buildNgcGuestTagsUrl(imageName string) (string, error) {
	orgName, teamName, name, err := c.parseNgcImageName(imageName)
	if err != nil {
		return "", fmt.Errorf("failed to parse NGC image name: %v", err)
	}

	ngcFullUrl := constants.NgcGuestTagsBaseURL

	// Build URL with query parameters for guest access
	if orgName != "" {
		ngcFullUrl = fmt.Sprintf("%s/%s", ngcFullUrl, orgName)
	}
	if teamName != "" {
		ngcFullUrl = fmt.Sprintf("%s/%s", ngcFullUrl, teamName)
	}

	ngcFullUrl = fmt.Sprintf("%s/%s/images", ngcFullUrl, name)

	return ngcFullUrl, nil
}

func (c *ngcClient) buildNgcAuthenticatedTagsUrl(imageName string) (string, error) {
	orgName, teamName, name, err := c.parseNgcImageName(imageName)
	if err != nil {
		return "", fmt.Errorf("failed to parse NGC image name: %v", err)
	}

	ngcFullUrl := fmt.Sprintf("%s/%s", constants.NgcAuthenticatedTagsBaseURL, orgName)
	if teamName != "" {
		ngcFullUrl = fmt.Sprintf("%s/team/%s", ngcFullUrl, teamName)
	}

	ngcFullUrl = fmt.Sprintf("%s/repos/%s/images", ngcFullUrl, name)

	return ngcFullUrl, nil
}

func (c *ngcClient) ngcGetRequest(ctx context.Context, url string, apiKey string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create NGC API request")
	}

	setHeaders(req, apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("failed to call NGC API: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to read NGC API response")
	}

	if resp.StatusCode != http.StatusOK {
		return body, resp.StatusCode, fmt.Errorf("NGC API returned status %d: %s", resp.StatusCode, body)
	}

	return body, http.StatusOK, nil
}

func (c *ngcClient) parseNgcImageName(input string) (string, string, string, error) {
	parts := strings.Split(input, "/")

	if len(parts) == 3 {
		orgName, teamName, name := parts[0], parts[1], parts[2]
		return orgName, teamName, name, nil
	} else if len(parts) == 2 {
		orgName, name := parts[0], parts[1]
		return orgName, "", name, nil
	}

	return "", "", "", fmt.Errorf("invalid NGC image name: %s", input)
}

func (c *ngcClient) parseTagsFromResponse(body []byte) ([]string, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("NGC API returned empty response")
	}

	var ngcResp imageTagsResponse
	if err := json.Unmarshal(body, &ngcResp); err != nil {
		return nil, fmt.Errorf("failed to parse NGC API response: %v", err)
	}

	if ngcResp.Images == nil {
		return nil, fmt.Errorf("NGC API response is empty")
	}

	tags := make([]string, len(ngcResp.Images))
	for i, image := range ngcResp.Images {
		tags[i] = image.Tag
	}

	return tags, nil
}

func setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", constants.UserAgentName)
	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}
}
