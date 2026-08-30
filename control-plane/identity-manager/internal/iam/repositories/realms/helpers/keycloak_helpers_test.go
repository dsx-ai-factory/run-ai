package realms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/stretchr/testify/assert"
)

func TestGetUserProfileMetadata_HappyFlow(t *testing.T) {
	httpReq := getHttpRequest()
	defer httpReq.Close()

	res, _ := GetUserProfileMetadata(context.Background(), "runai", "token", httpReq.URL)
	assert.NotNil(t, res)
}

func TestUpdateUserProfileMetadata_HappyFlow(t *testing.T) {
	httpReq := getHttpRequest()
	defer httpReq.Close()

	err := UpdateUserProfileMetadata(context.Background(), "runai", "token", httpReq.URL, models.UserProfileMetadata{})
	assert.Nil(t, err)
}

func getHttpRequest() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			res.WriteHeader(http.StatusOK)
			res.Write(getUserProfileMetadataResponse())
		}
		if req.Method == http.MethodPut {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte{})
		}
	}))
}

func getUserProfileMetadataResponse() []byte {
	externalTokenReponse := models.UserProfileMetadata{
		Attributes: nil,
		Groups:     nil,
	}
	jsonResponse, _ := json.Marshal(externalTokenReponse)
	return jsonResponse
}
