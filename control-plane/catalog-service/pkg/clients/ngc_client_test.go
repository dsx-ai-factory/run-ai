package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendNgcImagesRequestSearchesImageIdentifierAndPaginates(t *testing.T) {
	var (
		mu      sync.Mutex
		queries []map[string]interface{}
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query map[string]interface{}
		if err := json.Unmarshal([]byte(r.URL.Query().Get("q")), &query); err != nil {
			t.Errorf("decode query: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		queries = append(queries, query)
		mu.Unlock()

		page := int(query["page"].(float64))
		resourceID := "nvidia/pytorch"
		if page == 1 {
			resourceID = "nvidia/pytorch-ltsb2"
		}
		_, err := fmt.Fprintf(w, `{
			"resultPageTotal": 2,
			"results": [{"resources": [{"resourceId": %q}, {"resourceId": %q}]}]
		}`, resourceID, resourceID)
		if err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client := &ngcClient{httpClient: server.Client()}
	images, err := client.sendNgcImagesRequest(context.Background(), server.URL, "", "nvidia/pytorch", true)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"nvidia/pytorch", "nvidia/pytorch-ltsb2"}, images)
	require.Len(t, queries, 2)
	assert.Equal(t, float64(ngcImagesPageSize), queries[0]["pageSize"])
	assert.NotContains(t, queries[0], "query")

	filters := queries[0]["filters"].([]interface{})
	require.Len(t, filters, 3)
	assert.Equal(t, "isPublic", filters[0].(map[string]interface{})["field"])
	assert.Equal(t, "guestAccess", filters[1].(map[string]interface{})["field"])
	assert.Equal(t, "resourceId", filters[2].(map[string]interface{})["field"])
	assert.Equal(t, "nvidia/pytorch*", filters[2].(map[string]interface{})["value"])
	assert.Equal(t, float64(0), queries[0]["page"])
	assert.Equal(t, float64(1), queries[1]["page"])
}

func TestSendNgcImagesRequestForwardsTextSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query map[string]interface{}
		if err := json.Unmarshal([]byte(r.URL.Query().Get("q")), &query); err != nil {
			t.Errorf("decode query: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		assert.Equal(t, "pytorch", query["query"])
		assert.NotContains(t, query, "filters")
		_, err := fmt.Fprint(w, `{
			"resultPageTotal": 1,
			"results": [{"resources": [{"resourceId": "nvidia/pytorch"}]}]
		}`)
		if err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client := &ngcClient{httpClient: server.Client()}
	images, err := client.sendNgcImagesRequest(context.Background(), server.URL, "api-key", " pytorch ", false)

	require.NoError(t, err)
	assert.Equal(t, []string{"nvidia/pytorch"}, images)
}

func TestBuildImagesCacheKeyIncludesRegistryUrlAndSearch(t *testing.T) {
	assert.Equal(t, "ngc_images_registry-url_", buildImagesCacheKey("registry-url", ""))
	assert.Equal(t, "ngc_images_registry-url_nvidia/pytorch", buildImagesCacheKey("registry-url", " nvidia/pytorch "))
}
