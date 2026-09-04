/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"time"

	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	mcpRequestsTotal   = "mcp_requests_total"
	mcpRequestDuration = "mcp_request_duration_seconds"
	mcpToolCallsTotal  = "mcp_tool_calls_total"

	methodToolsCall = "tools/call"
)

var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

func Init(metricsPort int) error {
	_, err := runaimetrics.Init(&runaimetrics.Config{
		ExposePort:  metricsPort,
		ServiceName: "runai-mcp-server",
	})
	if err != nil {
		return err
	}
	runaimetrics.CounterVec(mcpRequestsTotal).Register("Total MCP protocol requests received", []string{"method"})
	runaimetrics.HistogramVec(mcpRequestDuration).Register("Duration of MCP protocol request handling in seconds", []string{"method"}, durationBuckets)
	runaimetrics.CounterVec(mcpToolCallsTotal).Register("Total MCP tool calls", []string{"tool_name", "status"})
	return nil
}

func MCPMiddleware(knownTools map[string]bool) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			result, err := next(ctx, method, req)
			elapsed := time.Since(start)

			runaimetrics.CounterVec(mcpRequestsTotal).Get().WithLabels(map[string]string{"method": normalizeLabel(method)}).Inc()
			runaimetrics.HistogramVec(mcpRequestDuration).Get().WithLabels(map[string]string{"method": normalizeLabel(method)}).Observe(elapsed.Seconds())

			if method == methodToolsCall {
				status := "ok"
				if err != nil {
					status = "error"
				}
				runaimetrics.CounterVec(mcpToolCallsTotal).Get().WithLabels(map[string]string{
					"tool_name": toolLabel(req.GetParams(), knownTools),
					"status":    status,
				}).Inc()
			}

			return result, err
		}
	}
}
