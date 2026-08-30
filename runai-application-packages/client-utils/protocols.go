package client_utils

import "strings"

const (
	HttpProtocol  = "http://"
	HttpsProtocol = "https://"
)

func GetInternalUrl(baseURL string) string {
	if !strings.HasPrefix(baseURL, HttpProtocol) && !strings.HasPrefix(baseURL, HttpsProtocol) {
		baseURL = HttpProtocol + baseURL
	}
	return baseURL
}

func GetExternalUrl(baseURL string) string {
	if !strings.HasPrefix(baseURL, HttpProtocol) && !strings.HasPrefix(baseURL, HttpsProtocol) {
		baseURL = HttpsProtocol + baseURL
	}
	return baseURL
}
