package runaimcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discardLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

type noopIn struct{}

func (noopIn) Validate() error { return nil }

func TestWithRecover_CatchesPanic(t *testing.T) {
	h := func(ctx context.Context, _ noopIn) (string, error) {
		panic("boom")
	}
	wrapped := withRecover(h, discardLogger())
	out, err := wrapped(context.Background(), noopIn{})
	require.ErrorIs(t, err, ErrInternal)
	assert.Empty(t, out)
}

func TestWithRecover_DoesNotLeakPanicValue(t *testing.T) {
	h := func(ctx context.Context, _ noopIn) (string, error) {
		panic("super-secret-token-abc123")
	}
	wrapped := withRecover(h, discardLogger())
	_, err := wrapped(context.Background(), noopIn{})
	require.Error(t, err)
	assert.Equal(t, ErrInternal.Error(), err.Error())
}

func TestWithValidate_WrapsWithErrValidation(t *testing.T) {
	innerErr := errors.New("value is required")
	h := func(ctx context.Context, _ validatingIn) (string, error) {
		t.Fatal("handler must not run when validation fails")
		return "", nil
	}
	wrapped := withValidate(h)
	_, err := wrapped(context.Background(), validatingIn{value: ""})
	require.ErrorIs(t, err, ErrValidation)
	assert.Equal(t, innerErr.Error(), err.Error(), "validationError must surface inner message verbatim")
}

func TestWithClassifyErrors_PrefixesWithCategory(t *testing.T) {
	cases := []struct {
		name   string
		in     error
		wantIs error
		prefix string
	}{
		{"not_found", fmt.Errorf("rotate sa-1: %w", ErrNotFound), ErrNotFound, "not_found: "},
		{"auth_wins_over_upstream", fmt.Errorf("call x: %w: %w", ErrAuth, ErrUpstream), ErrAuth, "auth: "},
		{"validation", validationError{inner: errors.New("required")}, ErrValidation, "validation: "},
		{"fallback", errors.New("something weird"), nil, "error: "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := func(ctx context.Context, _ noopIn) (string, error) { return "", tc.in }
			wrapped := withClassifyErrors(h)
			_, err := wrapped(context.Background(), noopIn{})
			require.Error(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), tc.prefix), "err %q does not start with %q", err.Error(), tc.prefix)
			if tc.wantIs != nil {
				assert.ErrorIs(t, err, tc.wantIs)
			}
		})
	}
}

func TestWithClassifyErrors_ScrubsInternalDetails(t *testing.T) {
	leaky := fmt.Errorf("opened secret_file=/tmp/foo handle=0xdeadbeef: %w", ErrInternal)
	h := func(ctx context.Context, _ noopIn) (string, error) { return "", leaky }
	wrapped := withClassifyErrors(h)
	_, err := wrapped(context.Background(), noopIn{})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret_file")
	assert.NotContains(t, err.Error(), "0xdeadbeef")
	assert.ErrorIs(t, err, ErrInternal)
}

func TestWithRecover_PassesThroughNormalErrors(t *testing.T) {
	wantErr := errors.New("normal upstream error")
	h := func(ctx context.Context, _ noopIn) (string, error) {
		return "", wantErr
	}
	wrapped := withRecover(h, discardLogger())
	_, err := wrapped(context.Background(), noopIn{})
	assert.ErrorIs(t, err, wantErr)
}

type validatingIn struct {
	value string
}

func (v validatingIn) Validate() error {
	if v.value == "" {
		return errors.New("value is required")
	}
	return nil
}

func TestWithValidate_ShortCircuitsOnError(t *testing.T) {
	handlerRan := false
	h := func(ctx context.Context, _ validatingIn) (string, error) {
		handlerRan = true
		return "ok", nil
	}
	wrapped := withValidate(h)
	out, err := wrapped(context.Background(), validatingIn{value: ""})
	require.Error(t, err)
	assert.Equal(t, "value is required", err.Error())
	assert.Empty(t, out)
	assert.False(t, handlerRan, "handler must not run when validation fails")
}

func TestWithValidate_PassesValidatedInput(t *testing.T) {
	h := func(ctx context.Context, in validatingIn) (string, error) {
		return in.value, nil
	}
	wrapped := withValidate(h)
	out, err := wrapped(context.Background(), validatingIn{value: "hi"})
	require.NoError(t, err)
	assert.Equal(t, "hi", out)
}

func TestWithLogging_LogsCategoryFieldsOnSuccess(t *testing.T) {
	var buf strings.Builder
	log := zerolog.New(&buf)
	h := func(ctx context.Context, _ noopIn) (string, error) {
		return "hello", nil
	}
	wrapped := withLogging(h, log, "my_tool")
	ctx := withSessionID(context.Background(), "sess-123")
	_, err := wrapped(ctx, noopIn{})
	require.NoError(t, err)
	got := buf.String()
	for _, want := range []string{
		`"level":"info"`,
		`"tool":"my_tool"`,
		`"category":"ok"`,
		`"session_id":"sess-123"`,
		`"output_bytes":7`, // len(`"hello"`)
		`"duration":`,
	} {
		assert.Contains(t, got, want)
	}
}

func TestWithLogging_CategoryForNotFound(t *testing.T) {
	var buf strings.Builder
	log := zerolog.New(&buf)
	h := func(ctx context.Context, _ noopIn) (string, error) {
		return "", fmt.Errorf("rotate sa-1: %w", ErrNotFound)
	}
	wrapped := withLogging(h, log, "my_tool")
	_, err := wrapped(context.Background(), noopIn{})
	require.Error(t, err)
	got := buf.String()
	for _, want := range []string{
		`"level":"warn"`,
		`"category":"not_found"`,
	} {
		assert.Contains(t, got, want)
	}
	assert.NotContains(t, got, "output_bytes", "error log should not include output_bytes")
}

func TestWithLogging_CategoryForInternal(t *testing.T) {
	var buf strings.Builder
	log := zerolog.New(&buf)
	// withRecover sits between withLogging and the panicking handler.
	h := func(ctx context.Context, _ noopIn) (string, error) {
		panic("kaboom")
	}
	wrapped := withLogging(withRecover(h, discardLogger()), log, "my_tool")
	_, err := wrapped(context.Background(), noopIn{})
	require.ErrorIs(t, err, ErrInternal)
	got := buf.String()
	for _, want := range []string{
		`"level":"error"`,
		`"category":"internal"`,
	} {
		assert.Contains(t, got, want)
	}
}

func TestWithLogging_OmitsEmptySessionID(t *testing.T) {
	var buf strings.Builder
	log := zerolog.New(&buf)
	h := func(ctx context.Context, _ noopIn) (string, error) { return "ok", nil }
	wrapped := withLogging(h, log, "my_tool")
	_, err := wrapped(context.Background(), noopIn{})
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), "session_id", "session_id must be omitted when empty")
}

func TestWithLogging_HandlerSeesEnrichedLoggerInContext(t *testing.T) {
	var buf strings.Builder
	log := zerolog.New(&buf)
	h := func(ctx context.Context, _ noopIn) (string, error) {
		zerolog.Ctx(ctx).Warn().Str("workload_id", "wl-1").Msg("events fetch failed")
		return "ok", nil
	}
	wrapped := withLogging(h, log, "my_tool")
	_, err := wrapped(context.Background(), noopIn{})
	require.NoError(t, err)
	got := buf.String()
	for _, want := range []string{
		`"level":"warn"`,
		`"tool":"my_tool"`,
		`"workload_id":"wl-1"`,
		`"message":"events fetch failed"`,
	} {
		assert.Contains(t, got, want, "expected handler log line; got %q", got)
	}
}
