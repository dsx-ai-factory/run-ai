/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composite

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

const (
	Parallel = iota
	Serial
)

type CompositeBuilder struct {
	builders []builders.Interface
	mode     int
}

func NewCompositeBuilder(builders []builders.Interface, mode int) CompositeBuilder {
	return CompositeBuilder{
		builders: builders,
		mode:     mode,
	}
}

func (h CompositeBuilder) Construct(ctx context.Context, tenant *builders.TenantInfo, builderResponse map[string]interface{}) (err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Constructing tenant %s", tenant.Name)
	switch h.mode {
	case Parallel:
		return h.constructParallel(ctx, tenant, builderResponse)
	default:
		return h.constructSerial(ctx, tenant, builderResponse)
	}
}

func (h CompositeBuilder) Destruct(ctx context.Context, tenant *builders.TenantInfo) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Destructing tenant %s", tenant.Name)
	var wg sync.WaitGroup
	for _, builder := range h.builders {
		builder := builder
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := builder.Destruct(ctx, tenant); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Destructing step %s failed for tenant %s. err: %v", builder.Name(), tenant.Name, err)
			}
		}()
	}
	wg.Wait()
	return nil
}

type builderResult struct {
	response map[string]interface{}
	err      error
}

func (h CompositeBuilder) constructParallel(ctx context.Context, tenant *builders.TenantInfo, builderResponse map[string]interface{}) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Constructing tenant %s", tenant.Name)

	results := make([]builderResult, len(h.builders))
	var wg sync.WaitGroup

	for i, b := range h.builders {
		i, b := i, b
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = h.constructInternalBuilder(ctx, b, tenant)
		}()
	}
	wg.Wait()

	var errs []string
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err.Error())
		} else {
			maps.Copy(builderResponse, r.response)
		}
	}
	if len(errs) > 0 {
		combined := strings.Join(errs, "; ")
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("one or more builders failed for tenant %s: %s", tenant.Name, combined)
		return fmt.Errorf("%s", combined)
	}
	return nil
}

func (h CompositeBuilder) constructInternalBuilder(ctx context.Context, builder builders.Interface, tenant *builders.TenantInfo) builderResult {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Constructing %s for tenant %s", builder.Name(), tenant.Name)
	response := make(map[string]interface{})
	if err := builder.Construct(ctx, tenant, response); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error constructing %s for tenant %s: %v", builder.Name(), tenant.Name, err)
		return builderResult{err: fmt.Errorf("%s: %s", builder.Name(), err.Error())}
	}
	return builderResult{response: response}
}

func (h CompositeBuilder) constructSerial(ctx context.Context, tenant *builders.TenantInfo, builderResponse map[string]interface{}) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Constructing tenant %s", tenant.Name)
	for _, builder := range h.builders {
		r := h.constructInternalBuilder(ctx, builder, tenant)
		if r.err != nil {
			return r.err
		}
		maps.Copy(builderResponse, r.response)
	}
	return nil
}
