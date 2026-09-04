/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// logHook is a hook for zerolog that stores log events.
type logHook struct {
	logEvents []zerolog.Event
}

// Run logs the event.
func (logHook *logHook) Run(logEvent *zerolog.Event, level zerolog.Level, message string) {
	logHook.logEvents = append(logHook.logEvents, *logEvent)
}

func TestHttpMiddleware_Info_DoesntPrintLogs(t *testing.T) {
	config := &logging.Config{LogLevel: "info"}
	_, logHook := hookLogger(config)

	handler := HttpMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("Hello, world!"))
	}))

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("User-Agent", "Go-http-client/1.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Empty(t, logHook.logEvents)
}

func TestHttpMiddleware_Debug_PrintsLogs(t *testing.T) {
	config := &logging.Config{LogLevel: "debug"}
	_, logHook := hookLogger(config)

	handler := HttpMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("Hello, world!"))
	}))

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("User-Agent", "Go-http-client/1.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert logs
	buf := getEventBuf(logHook.logEvents[0])

	assert.Len(t, logHook.logEvents, 1)
	assert.Contains(t, buf, `"method":"GET"`)
	assert.Contains(t, buf, `"path":"/foo"`)
	assert.Contains(t, buf, `"user_agent":"Go-http-client/1.1"`)
	assert.Contains(t, buf, `"status_code":206`)
}

func hookLogger(config *logging.Config) (logging.Logger, *logHook) {
	testLogger := logging.Init(config).(*logging.ConsoleLogger)
	logHook := &logHook{}
	testLogger.Logger = testLogger.Logger.Hook(logHook)
	return testLogger, logHook
}

func getEventBuf(e zerolog.Event) string {
	reflectedE := reflect.ValueOf(e)
	buf := reflectedE.FieldByName("buf")
	return string(buf.Bytes())
}
