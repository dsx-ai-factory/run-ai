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
	"sync"
	"time"
)

const inspectionParserVersion = "questions-v2"

// DefaultCacheTTL applies only to exact-version HTTPS NGC repository selections.
const DefaultCacheTTL = 10 * time.Minute

type cacheEntry struct {
	data    []byte
	expires time.Time
}

// inspectionFlight tracks one acquisition and its result. Eligible NGC requests
// share it; each waiter can leave independently without cancelling the others.
type inspectionFlight struct {
	done    chan struct{}
	cancel  context.CancelFunc
	waiters int
	data    []byte
	err     error
}

// Inspector caches successful NGC repository metadata and bounds all expensive inspections.
// Callers must authorize the chart source and credential on EVERY request, including cache hits.
// Scope must include tenant, caller and authorization-relevant groups, derived from trusted context.
type Inspector struct {
	mu               sync.Mutex
	fetch            FetchFunc
	ttl              time.Duration
	now              func() time.Time
	salt             [32]byte
	entries          map[string]cacheEntry
	flights          map[string]*inspectionFlight
	active           int
	cachedBytes      int
	maxCacheBytes    int
	maxRecordBytes   int
	maxEntries       int
	uncachedSequence uint64
}

// NewInspector creates a metadata cache with at most four concurrent acquisitions.
// A nil fetcher uses FetchRemote; a nonpositive TTL uses DefaultCacheTTL.
func NewInspector(fetch FetchFunc, ttl time.Duration) *Inspector {
	if fetch == nil {
		fetch = FetchRemote
	}
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	inspector := &Inspector{fetch: fetch, ttl: ttl, now: time.Now, entries: make(map[string]cacheEntry), flights: make(map[string]*inspectionFlight), maxCacheBytes: 20 << 20, maxRecordBytes: 4 << 20, maxEntries: 128}
	if _, err := rand.Read(inspector.salt[:]); err != nil {
		panic(fmt.Sprintf("initialize chart inspection cache: %v", err))
	}
	return inspector
}

// Inspect returns normalized metadata using the configured fetcher. The service
// must authorize each request first and supply a scope derived from trusted identity.
func (i *Inspector) Inspect(ctx context.Context, selection Selection, scope, credential string) (*Result, error) {
	return i.InspectWithFetch(ctx, selection, scope, credential, i.fetch)
}

// InspectWithFetch runs a service-provided fetcher inside the acquisition limit,
// allowing uploaded-chart reads to share the same bounds as remote downloads.
// Authorization is the caller's responsibility, including on cache hits.
func (i *Inspector) InspectWithFetch(ctx context.Context, selection Selection, scope, credential string, fetch FetchFunc) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := selection.validate(); err != nil {
		return nil, err
	}
	if fetch == nil {
		return nil, acquisitionError(500, "chart_inspection_failed", "Chart inspection source is not configured.")
	}
	if scope == "" {
		return nil, acquisitionError(400, "invalid_inspection_scope", "Chart inspection requires an authenticated caller scope.")
	}
	cacheable := cacheableNGCRepository(selection)
	key := ""
	if cacheable {
		key = i.key(selection, scope, credential)
	}
	i.mu.Lock()
	if !cacheable {
		// Fresh sources must not reuse a cached result or another request's acquisition.
		i.uncachedSequence++
		key = fmt.Sprintf("uncached:%d", i.uncachedSequence)
	}
	if entry, ok := i.entries[key]; ok {
		if i.now().Before(entry.expires) {
			data := entry.data
			i.mu.Unlock()
			return decodeResult(data)
		}
		i.deleteEntry(key)
	}
	flight := i.flights[key]
	if flight == nil {
		if i.active >= 4 {
			i.mu.Unlock()
			return nil, acquisitionError(503, "chart_inspection_capacity", "Chart inspection capacity reached; retry shortly.")
		}
		// One disconnected waiter must not cancel work still needed by others.
		// The shared deadline and releaseWaiter bound its lifetime instead.
		workCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), downloadTimeout)
		flight = &inspectionFlight{done: make(chan struct{}), cancel: cancel}
		i.flights[key] = flight
		i.active++
		// Add the first waiter before starting work, so immediate completion can be cached.
		flight.waiters = 1
		go i.run(workCtx, key, flight, selection, credential, fetch, cacheable)
	} else {
		flight.waiters++
	}
	i.mu.Unlock()
	select {
	case <-ctx.Done():
		i.releaseWaiter(key, flight)
		return nil, ctx.Err()
	case <-flight.done:
		i.releaseWaiter(key, flight)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if flight.err != nil {
			return nil, flight.err
		}
		return decodeResult(flight.data)
	}
}

// Stop abandoned work when its last waiter leaves. run retains the active slot
// until the fetcher exits, so cancellation cannot bypass the concurrency limit.
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

// run completes an acquisition for its waiters and releases its capacity slot.
// Only successful metadata still needed by a live waiter is eligible for caching.
func (i *Inspector) run(ctx context.Context, key string, flight *inspectionFlight, selection Selection, credential string, fetch FetchFunc, cacheable bool) {
	defer flight.cancel()
	data, err := fetch(ctx, selection, credential)
	var result *Result
	if err == nil {
		result, err = ParseArchive(data)
		if err != nil {
			err = fmt.Errorf("inspect chart archive: %w", err)
		}
	}
	if err == nil && result == nil {
		err = errors.New("chart inspection returned no result")
	}
	if err == nil && selection.Repo != "" && (result.ChartName != selection.ChartName || result.ChartVersion != selection.ChartVersion) {
		err = acquisitionError(422, "chart_identity_mismatch", "Downloaded chart identity does not match the selected name and version.")
	}
	var encoded []byte
	if err == nil {
		encoded, err = json.Marshal(result)
		if err != nil {
			err = errors.New("chart inspection metadata could not be encoded")
		}
	}
	i.mu.Lock()
	flight.data, flight.err = encoded, err
	if cacheable && err == nil && flight.waiters > 0 && ctx.Err() == nil {
		i.put(key, encoded)
	}
	if i.flights[key] == flight {
		delete(i.flights, key)
	}
	i.active--
	close(flight.done)
	i.mu.Unlock()
}

func cacheableNGCRepository(selection Selection) bool {
	if selection.Repo == "" || selection.ChartName == "" || selection.ChartVersion == "" {
		return false
	}
	u, err := validateSourceURL(selection.Repo)
	return err == nil && u.Scheme == httpsScheme && ngcHost(u.Hostname())
}

// Decode a fresh object per caller so response mutation cannot alter the cache.
// UseNumber preserves large integer defaults across the JSON round trip.
func decodeResult(data []byte) (*Result, error) {
	var result Result
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode chart inspection metadata: %w", err)
	}
	return &result, nil
}

// Include caller scope and a salted credential fingerprint so different access
// contexts cannot share metadata and raw key material is not retained in cache keys.
func (i *Inspector) key(selection Selection, scope, credential string) string {
	fingerprint := hmac.New(sha256.New, i.salt[:])
	_, _ = fingerprint.Write([]byte(credential))
	data, _ := json.Marshal([]any{scope, selection, hex.EncodeToString(fingerprint.Sum(nil)), inspectionParserVersion})
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func (i *Inspector) deleteEntry(key string) {
	entry := i.entries[key]
	i.cachedBytes -= len(entry.data)
	delete(i.entries, key)
}

// Called with the cache mutex held. Eviction uses expiration order; records contain immutable JSON only.
func (i *Inspector) put(key string, data []byte) {
	if len(data) > i.maxRecordBytes || len(data) > i.maxCacheBytes || i.maxEntries < 1 {
		return
	}
	now := i.now()
	for key, entry := range i.entries {
		if !now.Before(entry.expires) {
			i.deleteEntry(key)
		}
	}
	if _, ok := i.entries[key]; ok {
		i.deleteEntry(key)
	}
	for len(i.entries) >= i.maxEntries || i.cachedBytes+len(data) > i.maxCacheBytes {
		oldestKey := ""
		var oldest time.Time
		for candidate, entry := range i.entries {
			if oldestKey == "" || entry.expires.Before(oldest) {
				oldestKey, oldest = candidate, entry.expires
			}
		}
		if oldestKey == "" {
			return
		}
		i.deleteEntry(oldestKey)
	}
	i.entries[key] = cacheEntry{data: data, expires: now.Add(i.ttl)}
	i.cachedBytes += len(data)
}
