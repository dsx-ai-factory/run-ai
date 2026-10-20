/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package chart_inspection

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// Inspection flow:
// - Callers authorize every request, including cache hits.
// - CP cache stores parsed metadata for exact-version HTTPS NGC charts (default: 10 minutes).
//   Archive bytes are never cached. Custom URLs and uploads always fetch fresh data.
// - Cache misses start a local worker that fetches and parses the chart.
//   Matching NGC requests share that worker and use the same key as cached results.
// - Shared credentials reuse metadata within a tenant for the same asset and key value.
//   Personal credentials retain caller isolation; credential-free requests share within a tenant.
// - At most four workers run at once; excess new work returns 503.
//   A worker is cancelled when its last caller leaves.
// - Cache failures never fail inspection.

const (
	inspectionParserVersion  = "questions-v2"
	cacheSaltBytes           = sha256.Size
	defaultMaxCacheSizeMB    = 20
	defaultMaxRecordSizeKB   = 4096
	cacheAccessTimeout       = time.Second
	maxConcurrentInspections = 4
)

// DefaultCacheTTL is the NGC metadata lifetime.
const DefaultCacheTTL = 10 * time.Minute

// Raw JSON prevents the CP cache from converting integer defaults to float64.
type chartInspectionMetadata struct {
	Data json.RawMessage `json:"data"`
}

// inspectionFlight holds one worker and its result. Matching NGC callers share it.
type inspectionFlight struct {
	done    chan struct{}
	cancel  context.CancelFunc
	waiters int
	data    []byte
	err     *api.Error
}

// Inspector caches NGC metadata and limits concurrent inspections.
// Callers must authorize the source and credential before every request.
// Scope must include the trusted tenant and asset, plus caller identity for personal credentials.
type Inspector struct {
	mu               sync.Mutex
	fetch            FetchFunc
	cacheReady       bool
	salt             [cacheSaltBytes]byte
	flights          map[string]*inspectionFlight
	active           int
	uncachedSequence uint64
}

// NewInspector registers the CP metadata cache.
// A nil fetcher uses FetchRemote; a nonpositive TTL uses DefaultCacheTTL.
func NewInspector(fetch FetchFunc, ttl time.Duration) *Inspector {
	if fetch == nil {
		fetch = FetchRemote
	}
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	cacheErr := cache.RegisterStore[chartInspectionMetadata](
		options.Ttl(ttl),
		options.MaxStoreSizeMB(defaultMaxCacheSizeMB),
		options.MaxRecordSizeKB(defaultMaxRecordSizeKB),
	)
	if cacheErr != nil {
		logging.Warnf("Failed to register chart inspection cache; inspections will run uncached")
	}
	inspector := &Inspector{
		fetch:      fetch,
		cacheReady: cacheErr == nil,
		flights:    make(map[string]*inspectionFlight),
	}
	if _, err := rand.Read(inspector.salt[:]); err != nil {
		panic(fmt.Sprintf("initialize chart inspection cache: %v", err))
	}
	return inspector
}

// Inspect fetches and parses chart metadata, or returns a cached NGC result.
// The service must authorize the request and supply a trusted tenant/credential scope.
func (i *Inspector) Inspect(ctx context.Context, selection Selection, scope, credential string) (*api.AiApplicationChartInspectionResponse, *api.Error) {
	return i.InspectWithFetch(ctx, selection, scope, credential, i.fetch)
}

// InspectWithFetch applies the same concurrency limit to a custom fetcher, such as upload reads.
// Callers must authorize each request, including cache hits.
func (i *Inspector) InspectWithFetch(ctx context.Context, selection Selection, scope, credential string, fetch FetchFunc) (*api.AiApplicationChartInspectionResponse, *api.Error) {
	if err := ctx.Err(); err != nil {
		return nil, inspectionError(err)
	}
	if err := selection.validate(); err != nil {
		return nil, err
	}
	if fetch == nil {
		return nil, api.NewError(http.StatusInternalServerError, "chart_inspection_failed: Chart inspection source is not configured.", nil)
	}
	if scope == "" {
		return nil, api.NewValidationError("invalid_inspection_scope: Chart inspection requires an authenticated caller scope.")
	}
	cacheable := cacheableNGCRepository(selection)
	key := ""
	if cacheable {
		key = i.key(selection, scope, credential)
		result := i.cachedMetadata(ctx, key)
		if err := ctx.Err(); err != nil {
			return nil, inspectionError(err)
		}
		if result != nil {
			return result, nil
		}
	}
	i.mu.Lock()
	if !cacheable {
		// Custom URLs and uploads must not share work between requests.
		i.uncachedSequence++
		key = fmt.Sprintf("uncached:%d", i.uncachedSequence)
	}
	flight := i.flights[key]
	if flight == nil {
		if i.active >= maxConcurrentInspections {
			i.mu.Unlock()
			return nil, api.NewError(http.StatusServiceUnavailable, "chart_inspection_capacity: Chart inspection capacity reached; retry shortly.", nil)
		}
		// One caller leaving must not cancel work needed by others.
		// Stop on timeout or when the last caller leaves.
		workCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), downloadTimeout)
		flight = &inspectionFlight{done: make(chan struct{}), cancel: cancel}
		i.flights[key] = flight
		i.active++
		// Register the caller before starting a worker that could finish immediately.
		flight.waiters = 1
		go i.run(workCtx, key, flight, selection, credential, fetch, cacheable)
	} else {
		flight.waiters++
	}
	i.mu.Unlock()
	select {
	case <-ctx.Done():
		i.releaseWaiter(key, flight)
		return nil, inspectionError(ctx.Err())
	case <-flight.done:
		i.releaseWaiter(key, flight)
		if err := ctx.Err(); err != nil {
			return nil, inspectionError(err)
		}
		if flight.err != nil {
			return nil, flight.err
		}
		return decodeResult(flight.data)
	}
}

// Cancel work when the last caller leaves. Keep its slot until run exits
// so a slow cancellation cannot bypass the concurrency limit.
func (i *Inspector) releaseWaiter(key string, flight *inspectionFlight) {
	i.mu.Lock()
	defer i.mu.Unlock()
	flight.waiters--
	if flight.waiters == 0 {
		flight.cancel()
		if i.flights[key] == flight {
			delete(i.flights, key)
		}
	}
}

// run fetches, parses and delivers one result, then releases its worker slot.
// Cache successful NGC results only while a caller still needs them.
func (i *Inspector) run(ctx context.Context, key string, flight *inspectionFlight, selection Selection, credential string, fetch FetchFunc, cacheable bool) {
	defer flight.cancel()
	var result *api.AiApplicationChartInspectionResponse
	if cacheable {
		// Another worker may have filled the cache since the first lookup.
		result = i.cachedMetadata(ctx, key)
	}
	fromCache := result != nil
	var err *api.Error
	if !fromCache {
		data, fetchErr := fetch(ctx, selection, credential)
		err = fetchErr
		if err == nil {
			result, err = ParseArchive(data)
		}
	}
	if err == nil && result == nil {
		err = api.NewError(http.StatusInternalServerError, "chart inspection returned no result", nil)
	}
	if err == nil && selection.Repo != "" && (result.ChartName != selection.ChartName || result.ChartVersion != selection.ChartVersion) {
		err = api.NewValidationError("chart_identity_mismatch: Downloaded chart identity does not match the selected name and version.")
	}
	var encoded []byte
	if err == nil {
		var marshalErr error
		encoded, marshalErr = json.Marshal(result)
		if marshalErr != nil {
			err = api.NewError(http.StatusInternalServerError, "chart inspection metadata could not be encoded", nil)
		}
	}
	// Write outside the lock so other requests can proceed or cancel.
	// The last caller leaving cancels this write; write failures still return the result.
	if cacheable && !fromCache && err == nil && ctx.Err() == nil && i.cacheReady {
		cacheCtx, cancel := context.WithTimeout(ctx, cacheAccessTimeout)
		if cacheErr := cache.Set(cacheCtx, key, chartInspectionMetadata{Data: encoded}); cacheErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to cache chart inspection metadata; returning uncached result")
		}
		cancel()
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	flight.data, flight.err = encoded, err
	if i.flights[key] == flight {
		delete(i.flights, key)
	}
	i.active--
	close(flight.done)
}

// Preserve API errors, map cancellation to 408, and hide unexpected fetcher errors.
func inspectionError(err error) *api.Error {
	if err == nil {
		return nil
	}
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		return apiErr
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return api.NewError(http.StatusRequestTimeout, "chart inspection request cancelled or expired", nil)
	}
	return api.NewError(http.StatusInternalServerError, "chart inspection failed unexpectedly", nil)
}

func cacheableNGCRepository(selection Selection) bool {
	if selection.Repo == "" || selection.ChartName == "" || selection.ChartVersion == "" {
		return false
	}
	u, err := validateSourceURL(selection.Repo)
	return err == nil && u.Scheme == httpsScheme && ngcHost(u.Hostname())
}

// Give each caller its own result. UseNumber preserves large integer defaults.
func decodeResult(data []byte) (*api.AiApplicationChartInspectionResponse, *api.Error) {
	var result api.AiApplicationChartInspectionResponse
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "could not decode cached chart inspection metadata", nil)
	}
	return &result, nil
}

// Keys isolate chart, tenant/credential scope, actual key material and parser version.
// HMAC keeps raw credentials out of keys; the salt is unique to this Inspector.
func (i *Inspector) key(selection Selection, scope, credential string) string {
	fingerprint := hmac.New(sha256.New, i.salt[:])
	_, _ = fingerprint.Write([]byte(credential))
	data, _ := json.Marshal([]any{scope, selection, hex.EncodeToString(fingerprint.Sum(nil)), inspectionParserVersion})
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// Treat unavailable or invalid cache entries as misses.
func (i *Inspector) cachedMetadata(ctx context.Context, key string) *api.AiApplicationChartInspectionResponse {
	if !i.cacheReady {
		return nil
	}
	cacheCtx, cancel := context.WithTimeout(ctx, cacheAccessTimeout)
	defer cancel()
	entry, err := cache.Get[chartInspectionMetadata](cacheCtx, key)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to read chart inspection cache; inspecting source instead")
		return nil
	}
	if entry == nil {
		return nil
	}
	result, apiErr := decodeResult(entry.Data)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Invalid chart inspection cache entry; inspecting source instead")
		return nil
	}
	return result
}
