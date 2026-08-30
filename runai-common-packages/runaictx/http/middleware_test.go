package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
)

const (
	tenantIdValue           = "tenant-id"
	subjectValue            = "subject"
	groupsValue1            = "group1"
	groupsValue2            = "group2"
	supplementalGroupsValue = "supplementalGroups"
	gid                     = "gid in test"
	uid                     = "uid in test"
	clientIdValue           = "clientId"
)

func TestContextMiddleware_AllValuesExists(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set(tenantIdHeader, tenantIdValue)
	incomeReq.Header.Set(subjectIdHeader, subjectValue)
	incomeReq.Header.Set(clientIdHeader, clientIdValue)
	incomeReq.Header.Set(groupsHeader, fmt.Sprintf("%s,%s", groupsValue1, groupsValue2))
	incomeReq.Header.Set(supplementalGroupsHeader, supplementalGroupsValue)
	incomeReq.Header.Set(acceptHeader, "text/json")

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.Equal(t, "text/json", ctx.Value(runaictx.CfAccept))
		assert.Equal(t, tenantIdValue, ctx.Value(runaictx.CfTenantId))
		assert.Equal(t, subjectValue, ctx.Value(runaictx.CfSubjectID))
		assert.Equal(t, []string{groupsValue1, groupsValue2}, ctx.Value(runaictx.CfGroups).([]string))
		assert.Equal(t, supplementalGroupsValue, ctx.Value(runaictx.CfSupplementalGroups))
		assert.Equal(t, clientIdValue, ctx.Value(runaictx.CfClientId))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}

func TestContextMiddleware_MissingTenantId(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set(subjectIdHeader, subjectValue)
	incomeReq.Header.Add(groupsHeader, groupsValue1)
	incomeReq.Header.Add(groupsHeader, groupsValue2)
	incomeReq.Header.Set(supplementalGroupsHeader, supplementalGroupsValue)
	incomeReq.Header.Set(clientIdHeader, clientIdValue)

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.Nil(t, ctx.Value(runaictx.CfTenantId))
		assert.Equal(t, subjectValue, ctx.Value(runaictx.CfSubjectID))
		assert.Equal(t, []string{groupsValue1, groupsValue2}, ctx.Value(runaictx.CfGroups).([]string))
		assert.Equal(t, supplementalGroupsValue, ctx.Value(runaictx.CfSupplementalGroups))
		assert.Equal(t, clientIdValue, ctx.Value(runaictx.CfClientId))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}

func TestContextMiddleware_ResponseWithTraceId(t *testing.T) {
	traceId := "abc123traceId"
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set(traceIdHeader, traceId)

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, traceId, r.Context().Value(runaictx.CfTraceId))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
	assert.Equal(t, traceId, recorder.Header().Get(traceIdHeader))
}

func TestContextMiddleware_WithAcceptType(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set("ACCEPT", "text/csv")

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.Equal(t, "text/csv", ctx.Value(runaictx.CfAccept))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}

func TestContextMiddleware_MissingSubject(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set(tenantIdHeader, tenantIdValue)
	incomeReq.Header.Add(groupsHeader, groupsValue1)
	incomeReq.Header.Add(groupsHeader, groupsValue2)
	incomeReq.Header.Set(supplementalGroupsHeader, supplementalGroupsValue)
	incomeReq.Header.Set(clientIdHeader, clientIdValue)

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.Equal(t, tenantIdValue, ctx.Value(runaictx.CfTenantId))
		assert.Nil(t, ctx.Value(runaictx.CfSubjectID))
		assert.Equal(t, []string{groupsValue1, groupsValue2}, ctx.Value(runaictx.CfGroups).([]string))
		assert.Equal(t, supplementalGroupsValue, ctx.Value(runaictx.CfSupplementalGroups))
		assert.Equal(t, clientIdValue, ctx.Value(runaictx.CfClientId))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}

func TestContextMiddleware_MissingGroups(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set(tenantIdHeader, tenantIdValue)
	incomeReq.Header.Set(subjectIdHeader, subjectValue)
	incomeReq.Header.Set(supplementalGroupsHeader, supplementalGroupsValue)

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.Equal(t, tenantIdValue, ctx.Value(runaictx.CfTenantId))
		assert.Equal(t, subjectValue, ctx.Value(runaictx.CfSubjectID))
		assert.Nil(t, ctx.Value(runaictx.CfGroups))
		assert.Equal(t, supplementalGroupsValue, ctx.Value(runaictx.CfSupplementalGroups))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}

func TestContextMiddleware_MissingSupplemental(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set(tenantIdHeader, tenantIdValue)
	incomeReq.Header.Set(subjectIdHeader, subjectValue)
	incomeReq.Header.Add(groupsHeader, groupsValue1)
	incomeReq.Header.Add(groupsHeader, groupsValue2)

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.Equal(t, tenantIdValue, ctx.Value(runaictx.CfTenantId))
		assert.Equal(t, subjectValue, ctx.Value(runaictx.CfSubjectID))
		assert.Equal(t, []string{groupsValue1, groupsValue2}, ctx.Value(runaictx.CfGroups).([]string))
		assert.Nil(t, ctx.Value(runaictx.CfSupplementalGroups))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}

func TestContextMiddleware_MissingClientId(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	incomeReq.Header.Set(tenantIdHeader, tenantIdValue)
	incomeReq.Header.Set(subjectIdHeader, subjectValue)
	incomeReq.Header.Set(supplementalGroupsHeader, supplementalGroupsValue)
	incomeReq.Header.Set(gIdHeader, gid)
	incomeReq.Header.Set(uIdHeader, uid)

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.Equal(t, tenantIdValue, ctx.Value(runaictx.CfTenantId))
		assert.Equal(t, subjectValue, ctx.Value(runaictx.CfSubjectID))
		assert.Equal(t, supplementalGroupsValue, ctx.Value(runaictx.CfSupplementalGroups))
		assert.Equal(t, gid, ctx.Value(runaictx.CfGId))
		assert.Equal(t, uid, ctx.Value(runaictx.CfUId))
		assert.Nil(t, ctx.Value(runaictx.CfClientId))
	})

	handler := ContextMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}
