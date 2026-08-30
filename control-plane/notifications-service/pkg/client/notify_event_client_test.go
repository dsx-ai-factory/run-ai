package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetError_BodyIsNil(t *testing.T) {
	err := getError(&http.Response{StatusCode: http.StatusOK, Status: "OK"})
	assert.Equal(t, "failed: status: OK, error: ", err.Error())
}

func TestGetError_BodyNotEmpty(t *testing.T) {
	body := strings.NewReader("some error")
	err := getError(&http.Response{StatusCode: http.StatusOK, Status: "OK", Body: io.NopCloser(body)})
	assert.Equal(t, "failed: status: OK, error: some error", err.Error())
}
