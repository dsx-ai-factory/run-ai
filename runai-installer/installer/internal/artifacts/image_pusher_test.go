package artifacts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckPushPermissionHonorsContext(t *testing.T) {
	started := make(chan struct{})
	var once sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/blobs/uploads/") {
			once.Do(func() { close(started) })
			<-r.Context().Done()
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	host := strings.TrimPrefix(server.URL, "http://")
	pusher, err := NewImagePusher(Credential{Prefix: host, Insecure: true})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = pusher.CheckPushPermission(ctx, host+"/runai/test:latest")

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	select {
	case <-started:
	default:
		t.Fatal("push permission request did not start")
	}
}

func TestPushRejectsEmptyExpectedDigest(t *testing.T) {
	pusher, err := NewImagePusher(Credential{})
	require.NoError(t, err)

	err = pusher.Push(context.Background(), t.TempDir(), "registry.example.com/runai/image:latest", "")

	require.EqualError(t, err, "image registry.example.com/runai/image:latest has no expected digest")
}
