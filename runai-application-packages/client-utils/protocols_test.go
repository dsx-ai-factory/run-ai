package client_utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInternal(t *testing.T) {
	host := "abc.com"
	url := GetInternalUrl(HttpsProtocol + host)
	assert.True(t, strings.HasPrefix(url, HttpsProtocol))

	url = GetInternalUrl(HttpProtocol + host)
	assert.True(t, strings.HasPrefix(url, HttpProtocol))

	url = GetInternalUrl(host)
	assert.True(t, strings.HasPrefix(url, HttpProtocol))
}

func TestExternal(t *testing.T) {
	host := "abc.com"
	url := GetExternalUrl(HttpsProtocol + host)
	assert.True(t, strings.HasPrefix(url, HttpsProtocol))

	url = GetExternalUrl(HttpProtocol + host)
	assert.True(t, strings.HasPrefix(url, HttpProtocol))

	url = GetExternalUrl(host)
	assert.True(t, strings.HasPrefix(url, HttpsProtocol))
}
