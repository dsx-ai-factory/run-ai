package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLoginURL_MissingUserAgent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, err := OpenapiClient.LoginsAPI.GetLoginScreenURL(ctx).AuthenticationFlow("test").RedirectUrl("").Execute()
	require.NotNil(t, err)
	require.Equal(t, "userAgent is required and must be specified", err.Error())
}
