/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package advanced

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/service"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

var (
	decrypt            = encryption.NewEncrypterDecrypter([]byte("0123456789abcdefghwertol"))
	secretFieldHandler = encryption.NewSecretFieldHandler(decrypt, &route.BasicRouteConfiguration{})
	tenantId           = "tenant1"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetRouteProviders_OfTheRightListener(t *testing.T) {
	properties := map[string]interface{}{}
	properties["tenant_id"] = tenantId

	config := datatypes.JSON(`{
"route": {
      "concurrent_limit": 50,
      "cached_events": 1000
    },
    "listener_id": "listener1"
    }`)
	routes := []*models.RouteGlobalConfig{
		{
			TenantId:    tenantId,
			RouteId:     "email_k8s_notifications",
			RouteConfig: config,
		},
	}
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchRoutes", "listener1", tenantId).Return(routes, nil)

	routeProviders := provider.GetRouteProviders("listener1", &events.Event{
		Properties: properties,
	})
	require.Len(t, routeProviders, 1)
	require.Equal(t, "listener1", routeProviders[0].GetListenerId())
}

func TestGetRouteProviders_NoRoutesForUser(t *testing.T) {
	properties := map[string]interface{}{}
	properties["tenant_id"] = tenantId
	properties["user_id"] = "user1"

	config := datatypes.JSON(`{
"route": {
      "concurrent_limit": 50,
      "cached_events": 1000
    },
    "listener_id": "listener1"
    }`)
	routes := []*models.RouteGlobalConfig{
		{
			TenantId:    tenantId,
			RouteId:     "email_k8s_notifications",
			RouteConfig: config,
		},
	}
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchRoutes", "listener1", tenantId).Return(routes, nil)
	dao.On("FetchUserRoutes", "listener1", tenantId, "user1").Return([]*models.RouteGlobalConfig{}, nil)

	routeProviders := provider.GetRouteProviders("listener1", &events.Event{
		Properties: properties,
	})
	require.Equal(t, len(routeProviders), len(routes))
}

func TestGetRouteProviders_ErrorRoutesForUser(t *testing.T) {
	properties := map[string]interface{}{}
	properties["tenant_id"] = tenantId
	properties["user_id"] = "user1"
	config := datatypes.JSON(`{
"route": {
      "concurrent_limit": 50,
      "cached_events": 1000
    },
    "listener_id": "listener1"
    }`)
	routes := []*models.RouteGlobalConfig{
		{
			TenantId:    tenantId,
			RouteId:     "email_k8s_notifications",
			RouteConfig: config,
		},
	}
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchRoutes", "listener1", tenantId).Return(routes, nil)
	dao.On("FetchUserRoutes", "listener1", tenantId, "user1").Return(routes, errors.New("error"))

	routeProviders := provider.GetRouteProviders("listener1", &events.Event{
		Properties: properties,
	})
	require.Len(t, routeProviders, 0)
}

func TestGetRouteProviders_NoRoutesForTenant(t *testing.T) {
	properties := map[string]interface{}{}
	properties["tenant_id"] = "tenant2"
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchRoutes", "listener2", "tenant2").Return([]*models.RouteGlobalConfig{}, nil)

	routeProviders := provider.GetRouteProviders("listener2", &events.Event{
		Properties: properties,
	})
	require.Len(t, routeProviders, 0)
}

func TestGetRouteProviders_NoProperties(t *testing.T) {
	properties := map[string]interface{}{}
	properties["tenant_id"] = "tenant2"
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	routeProviders1 := provider.GetRouteProviders("listener1", &events.Event{
		Properties: nil,
	})
	require.Len(t, routeProviders1, 0)
	routeProviders2 := provider.GetRouteProviders("listener1", &events.Event{
		Properties: map[string]interface{}{},
	})
	require.Len(t, routeProviders2, 0)
}

func TestGetListeningProviders_ShouldReturnAllListeners(t *testing.T) {
	listeners := []*models.ListenersConfig{
		{
			ListenerId:     "listeners1",
			ListenerConfig: datatypes.JSON(`{"relevant_namespaces": ["runai-frontend.*"], "relevant_objects": [{"kind": "Pod","reasons": ["Evict", "Scheduled"] }]}`),
			ListenerType:   "kube_events",
		},
		{
			ListenerId:     "listeners2",
			ListenerConfig: datatypes.JSON(`{"port": 9090}`),
			ListenerType:   "prometheus_alerts",
		},
	}
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchListeners").Return(listeners, nil)

	fetchedListeners := provider.GetListeningProviders()
	require.Len(t, fetchedListeners, 2)
	require.Equal(t, fetchedListeners["listeners1"].GetListenerType(), "kube_events")
	require.Equal(t, fetchedListeners["listeners2"].GetListenerType(), "prometheus_alerts")
}

func TestGetListeningProviders_SetsConfigValues(t *testing.T) {
	listeners := []*models.ListenersConfig{
		{
			ListenerId:     "listeners1",
			ListenerConfig: datatypes.JSON(`{"relevant_namespaces": ["runai-frontend.*"], "relevant_objects": [{"kind": "Pod","reasons": ["Evict", "Scheduled"] }]}`),
			ListenerType:   "kube_events",
		},
	}
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchListeners").Return(listeners, nil)

	fetchedListeners := provider.GetListeningProviders()
	require.Len(t, fetchedListeners, 1)
	require.Equal(t, fetchedListeners["listeners1"].GetListenerType(), "kube_events")
	require.Contains(t, fetchedListeners["listeners1"].GetListeningConfig().KubeListener.RelevantNamespaces, "runai-frontend.*")
}

func TestGetServer(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)

	register := NewRoutesRegister(dao, secretFieldHandler, make(chan string, DefaultUpdateEventsChannelSize))

	router := chi.NewRouter()
	register.Register(router)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	dao.On("FetchRoutes", "", "tenant1").Return([]*models.RouteGlobalConfig{}, nil).Once()
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/tenants/tenant1/routes", server.URL), strings.NewReader(""))
	require.Nil(t, err)

	client := http.Client{}
	response, err := client.Do(req)
	require.Nil(t, err)
	require.NotEmpty(t, response)
	require.Equal(t, response.StatusCode, http.StatusOK)
}

func TestInsertSinkConfig_CreateSucceeds(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := &ConfigurationProvider{dao: dao}

	dao.On("UpsertSink", "sink1", sinks.EmailSinkType, mock.AnythingOfType("datatypes.JSON")).Return(nil).Once()

	insertSinkConfig("sink1", sinks.EmailSinkType, sinks.EmailSinkConfig{
		From: "a@b.com", LogoUrl: "http://x", FromDisplayName: "n", Subject: "s",
	}, provider)
}

func TestInsertSinkConfig_CreateNonDuplicateError(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := &ConfigurationProvider{dao: dao}

	dao.On("UpsertSink", "sink1", sinks.EmailSinkType, mock.AnythingOfType("datatypes.JSON")).
		Return(errors.New("db down")).Once()

	insertSinkConfig("sink1", sinks.EmailSinkType, sinks.EmailSinkConfig{
		From: "a@b.com", LogoUrl: "http://x", FromDisplayName: "n", Subject: "s",
	}, provider)
}

func TestInsertSinkConfig_CreateDaoErrorWrongCode(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := &ConfigurationProvider{dao: dao}

	dao.On("UpsertSink", "sink1", sinks.EmailSinkType, mock.AnythingOfType("datatypes.JSON")).
		Return(configprovider.NewError("not found", configprovider.ErrCodeNotFound)).Once()

	insertSinkConfig("sink1", sinks.EmailSinkType, sinks.EmailSinkConfig{
		From: "a@b.com", LogoUrl: "http://x", FromDisplayName: "n", Subject: "s",
	}, provider)
}

func TestInsertSinkConfig_ParseFails(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := &ConfigurationProvider{dao: dao}

	insertSinkConfig("sink1", sinks.EmailSinkType, make(chan int), provider)
}

func TestGetSinkProvider_Email_NoEvent(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{
    "user": "team@run.ai",
    "password": "12345asd!",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "recipients": ["user@example.com"]
  }`),
	}}, nil)

	c, err := provider.GetSinkProvider(nil, "some_id")
	require.NoError(t, err)
	config, err := c.GetSinks("some_id")
	require.NoError(t, err)
	require.NotNil(t, config["some_id"].Email)
	require.Equal(t, false, config["some_id"].Email.DirectNotifications)
	require.Equal(t, "nvidia-runai@nvidia.com", config["some_id"].Email.From)
	require.Equal(t, "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}", config["some_id"].Email.Subject)
	require.Equal(t, "generic_notification.tpl", config["some_id"].Email.TemplateFile)
	require.Equal(t, "sinks_templates/email", config["some_id"].Email.TemplatesFolder)
}

func TestGetSinkProvider_Email_WithEventNoProperties(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}
	event := &events.Event{}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{
    "user": "team@run.ai",
    "password": "12345asd!",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "recipients": ["user@example.com"]
  }`),
	}}, nil)

	c, err := provider.GetSinkProvider(event, "some_id")
	require.NoError(t, err)
	config, err := c.GetSinks("some_id")
	require.NoError(t, err)
	require.NotNil(t, config["some_id"].Email)
	require.Equal(t, false, config["some_id"].Email.DirectNotifications)
	require.Equal(t, "nvidia-runai@nvidia.com", config["some_id"].Email.From)
	require.Equal(t, "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}", config["some_id"].Email.Subject)
	require.Equal(t, "generic_notification.tpl", config["some_id"].Email.TemplateFile)
	require.Equal(t, "sinks_templates/email", config["some_id"].Email.TemplatesFolder)
}

func TestGetSinkProvider_Email_WithEventNoCategory(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}
	event := &events.Event{
		Properties: map[string]interface{}{
			"type": "type",
		},
	}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{
    "user": "team@run.ai",
    "password": "12345asd!",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "recipients": ["user@example.com"]
  }`),
	}}, nil)

	c, err := provider.GetSinkProvider(event, "some_id")
	require.NoError(t, err)
	config, err := c.GetSinks("some_id")
	require.NoError(t, err)
	require.NotNil(t, config["some_id"].Email)
	require.Equal(t, false, config["some_id"].Email.DirectNotifications)
	require.Equal(t, "nvidia-runai@nvidia.com", config["some_id"].Email.From)
	require.Equal(t, "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}", config["some_id"].Email.Subject)
	require.Equal(t, "generic_notification.tpl", config["some_id"].Email.TemplateFile)
	require.Equal(t, "sinks_templates/email", config["some_id"].Email.TemplatesFolder)
}

func TestGetSinkProvider_Email_WithEventNoType(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}
	event := &events.Event{
		Properties: map[string]interface{}{
			"category": "category",
			"type":     "",
		},
	}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{
    "user": "team@run.ai",
    "password": "12345asd!",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "recipients": ["user@example.com"]
  }`),
	}}, nil)

	c, err := provider.GetSinkProvider(event, "some_id")
	require.NoError(t, err)
	config, err := c.GetSinks("some_id")
	require.NoError(t, err)
	require.NotNil(t, config["some_id"].Email)
	require.Equal(t, false, config["some_id"].Email.DirectNotifications)
	require.Equal(t, "nvidia-runai@nvidia.com", config["some_id"].Email.From)
	require.Equal(t, "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}", config["some_id"].Email.Subject)
	require.Equal(t, "generic_notification.tpl", config["some_id"].Email.TemplateFile)
	require.Equal(t, "sinks_templates/email", config["some_id"].Email.TemplatesFolder)
}

func TestGetSinkProvider_Email_WithEvent_WithEmptyNotifications(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}
	event := &events.Event{
		Properties: map[string]interface{}{
			"category": "category",
			"type":     "type",
		},
	}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{
    "user": "team@run.ai",
    "password": "12345asd!",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "recipients": ["user@example.com"]
  }`),
	}}, nil)
	dao.On("FetchNotifications", "", "category", "type").Return([]*models.Notification{}, nil)

	c, err := provider.GetSinkProvider(event, "some_id")
	require.NoError(t, err)
	config, err := c.GetSinks("some_id")
	require.NoError(t, err)
	require.NotNil(t, config["some_id"].Email)
	require.Equal(t, false, config["some_id"].Email.DirectNotifications)
	require.Equal(t, "nvidia-runai@nvidia.com", config["some_id"].Email.From)
	require.Equal(t, "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}", config["some_id"].Email.Subject)
	require.Equal(t, "generic_notification.tpl", config["some_id"].Email.TemplateFile)
	require.Equal(t, "sinks_templates/email", config["some_id"].Email.TemplatesFolder)
}

func TestGetSinkProvider_Email_WithEvent_WithErrorNotification(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}
	event := &events.Event{
		Properties: map[string]interface{}{
			"category": "category",
			"type":     "type",
		},
	}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{
    "user": "team@run.ai",
    "password": "12345asd!",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "recipients": ["user@example.com"]
  }`),
	}}, nil)
	dao.On("FetchNotifications", "", "category", "type").Return([]*models.Notification{
		{
			Template: "template",
		},
	}, errors.New("error"))

	c, err := provider.GetSinkProvider(event, "some_id")
	require.NoError(t, err)
	config, err := c.GetSinks("some_id")
	require.NoError(t, err)
	require.NotNil(t, config["some_id"].Email)
	require.Equal(t, false, config["some_id"].Email.DirectNotifications)
	require.Equal(t, "nvidia-runai@nvidia.com", config["some_id"].Email.From)
	require.Equal(t, "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}", config["some_id"].Email.Subject)
	require.Equal(t, "generic_notification.tpl", config["some_id"].Email.TemplateFile)
	require.Equal(t, "sinks_templates/email", config["some_id"].Email.TemplatesFolder)
}

func TestGetSinkProvider_Email_WithEvent_WithNotification(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}
	event := &events.Event{
		Properties: map[string]interface{}{
			"category": "category",
			"type":     "type",
		},
	}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{
    "user": "team@run.ai",
    "password": "12345asd!",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "recipients": ["user@example.com"]
  }`),
	}}, nil)
	dao.On("FetchNotifications", "", "category", "type").Return([]*models.Notification{
		{
			Template: "template",
		},
	}, nil)

	c, err := provider.GetSinkProvider(event, "some_id")
	require.NoError(t, err)
	config, err := c.GetSinks("some_id")
	require.NoError(t, err)
	require.NotNil(t, config["some_id"].Email)
	require.Equal(t, false, config["some_id"].Email.DirectNotifications)
	require.Equal(t, "nvidia-runai@nvidia.com", config["some_id"].Email.From)
	require.Equal(t, "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}", config["some_id"].Email.Subject)
	require.Equal(t, "template.tpl", config["some_id"].Email.TemplateFile)
	require.Equal(t, "sinks_templates/email", config["some_id"].Email.TemplatesFolder)
}

func TestUpdateServer(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	register := NewRoutesRegister(dao, secretFieldHandler, make(chan string, DefaultUpdateEventsChannelSize))

	router := chi.NewRouter()
	register.Register(router)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	configStr := `{
"route": {
      "concurrent_limit": 50,
      "cached_events": 1000
    },
    "listener_id": "listener1"
    }`
	routeStr := `{"tenant_id":"tenant1","route_config":` + configStr + `}`

	dao.On("UpdateRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	dao.On("FetchRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.RouteGlobalConfig{RouteConfig: datatypes.JSON("{}")}, nil).Once()
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/tenants/tenant1/routes/email_k8s_notifications", server.URL), strings.NewReader(routeStr))
	require.Nil(t, err)

	client := http.Client{}
	response, err := client.Do(req)
	require.NoError(t, err)
	require.NotEmpty(t, response)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func TestMergeRoutes_UserOverridesNotify(t *testing.T) {
	tenantRoute := models.RouteGlobalConfig{
		RouteConfig: datatypes.JSON(`{
      "notify_ids": {
        "email": {
		  "templates_folder": "sinks_templates/email",
		  "template_file": "tenant_email.html",
          "from": "tenant@mail.com",
          "user": "tenant_user",
          "password": "tenant_password",
          "smtp_host": "smtp.mail.com"
        }
	}
}
`),
	}
	userRoute1 := models.RouteGlobalConfig{
		RouteConfig: datatypes.JSON(`{
      "notify_ids": {
        "email": {
		  "templates_folder": "sinks_templates/email",
		  "template_file": "user_email.html",
          "from": "user@mail.com",
          "user": "user_user",
          "password": "user_password"
        }
    }
}`),
	}
	userRoute2 := models.RouteGlobalConfig{
		RouteConfig: datatypes.JSON(`{
      "notify_ids": {
        "email": {
		  "templates_folder": "sinks_templates/email",
          "template_file": "user2_email.html",
          "from": "user2@mail.com",
          "user": "user2_user",
          "password": "user2_password"
        }
    }
}`),
	}
	routes := []*models.RouteGlobalConfig{&userRoute1, &userRoute2}
	PatchRoutes(routes, &tenantRoute)

	require.Equal(
		t,
		datatypes.JSON(`{"notify_ids":{"email":{"from":"user@mail.com","password":"user_password","smtp_host":"smtp.mail.com","template_file":"user_email.html","templates_folder":"sinks_templates/email","user":"user_user"}}}`).String(),
		routes[0].RouteConfig.String(),
	)
	require.Equal(
		t,
		datatypes.JSON(`{"notify_ids":{"email":{"from":"user2@mail.com","password":"user2_password","smtp_host":"smtp.mail.com","template_file":"user2_email.html","templates_folder":"sinks_templates/email","user":"user2_user"}}}`).String(),
		routes[1].RouteConfig.String(),
	)
}

func TestMergeRoutes_UserOverridesEnrich(t *testing.T) {
	tenantRoute := models.RouteGlobalConfig{
		RouteConfig: datatypes.JSON(`{
      "enrich": {
        "kube_metadata": {
          "labels": "label"
        },
        "static": {
          "environment": "test"
        }
      }
}
`),
	}
	userRoute1 := models.RouteGlobalConfig{
		RouteConfig: datatypes.JSON(`{
      "enrich": {
        "kube_metadata": {
          "labels": "label_user1"
        },
        "static": {
          "environment": "dev"
        }
      }
}`),
	}
	userRoute2 := models.RouteGlobalConfig{
		RouteConfig: datatypes.JSON(`{
      "enrich": {
        "kube_metadata": {
          "labels": "label_user2"
        },
        "static": {
          "environment": "dev2"
        }
      }
}`),
	}
	routes := []*models.RouteGlobalConfig{&userRoute1, &userRoute2}
	PatchRoutes(routes, &tenantRoute)

	require.Equal(t, routes[0].RouteConfig.String(), datatypes.JSON(`{"enrich":{"kube_metadata":{"labels":"label_user1"},"static":{"environment":"dev"}}}`).String())
	require.Equal(t, routes[1].RouteConfig.String(), datatypes.JSON(`{"enrich":{"kube_metadata":{"labels":"label_user2"},"static":{"environment":"dev2"}}}`).String())
}

func TestGetRouteProviders_MergeRoutes(t *testing.T) {
	config := datatypes.JSON(`{
    "listener_id": "http-runai",
    "notify_ids": "mail",
    "filtering": {
      "policy": "drop"
    }
}`)
	userConfig := datatypes.JSON(`{
    "listener_id": "listener-from-user-route",
    "filtering": {
        "policy": "all",
        "filters":[
        {
          "type": "matcher",
          "matchers": [
            {
              "match_type": "==",
              "key": "tenant_id",
              "value": "tenant2"
            },
            {
              "match_type": "==",
              "key": "user_id",
              "value": "user@example.com"
            }
          ]
        }
      ]}

  }`)
	route := &models.RouteGlobalConfig{
		TenantId:    "tenant1",
		RouteId:     "email_k8s_notifications",
		RouteConfig: config,
	}
	userRoutes := []*models.RouteGlobalConfig{
		{
			TenantId:    "tenant1",
			RouteId:     "email_k8s_notifications",
			RouteConfig: userConfig,
		},
	}
	PatchRoutes(userRoutes, route)

	require.Len(t, userRoutes, 1)
	p := map[string]interface{}{}
	_ = json.Unmarshal(userRoutes[0].RouteConfig, &p)
	require.Equal(t, p["listener_id"], "listener-from-user-route")
	require.Equal(t, p["filtering"].(map[string]interface{})["policy"], "all")
	require.Len(t, p["filtering"].(map[string]interface{})["filters"].([]interface{})[0].(map[string]interface{})["matchers"], 2)
}

func TestGetRouteProviders_MergeRoutesTenantRouteWithFilters(t *testing.T) {
	config := datatypes.JSON(`{
    "listener_id": "http-runai",
    "notify_ids": "mail",
    "filtering": {
      "policy": "all",
       "filters":[
        {
          "type": "matcher",
          "matchers": [
            {
              "match_type": "==",
              "key": "tenant_id",
              "value": "tenant2"
            }
          ]
        }
    }
}`)
	userConfig := datatypes.JSON(`{
    "listener_id": "listener-from-user-route",
    "filtering": {
        "policy": "any",
        "filters":[
        {
          "type": "matcher",
          "matchers": [
            {
              "match_type": "==",
              "key": "user_id",
              "value": "user@example.com"
            }
          ]
        }
      ]}

  }`)
	route := &models.RouteGlobalConfig{
		TenantId:    "tenant1",
		RouteId:     "email_k8s_notifications",
		RouteConfig: config,
	}
	userRoutes := []*models.RouteGlobalConfig{{
		TenantId:    "tenant1",
		RouteId:     "email_k8s_notifications",
		RouteConfig: userConfig,
	}}
	PatchRoutes(userRoutes, route)

	require.Len(t, userRoutes, 1)
	p := map[string]interface{}{}
	_ = json.Unmarshal(userRoutes[0].RouteConfig, &p)
	require.Equal(t, p["listener_id"], "listener-from-user-route")
	require.Equal(t, p["filtering"].(map[string]interface{})["policy"], "any")
	require.Len(t, p["filtering"].(map[string]interface{})["filters"].([]interface{})[0].(map[string]interface{})["matchers"], 1)
	require.Equal(t, p["filtering"].(map[string]interface{})["filters"].([]interface{})[0].(map[string]interface{})["matchers"].([]interface{})[0].(map[string]interface{})["value"], "user@example.com")
}

func TestGetRouteProviders_MergeRoutesTenantRouteUserFilterNonExistent(t *testing.T) {
	config := datatypes.JSON(`{
    "listener_id": "http-runai",
    "notify_ids": "mail",
    "filtering": {
        "policy": "any",
        "filters":[
        {
          "type": "matcher",
          "matchers": [
            {
              "match_type": "==",
              "key": "user_id",
              "value": "user@example.com"
            }
          ]
        }
      ]}
}`)
	userConfig := datatypes.JSON(`{
    "listener_id": "listener-from-user-route",
    "filtering": {
	"filters": null
	}

  }`)
	route := &models.RouteGlobalConfig{
		TenantId:    "tenant1",
		RouteId:     "email_k8s_notifications",
		RouteConfig: config,
	}
	userRoutes := []*models.RouteGlobalConfig{{
		TenantId:    "tenant1",
		RouteId:     "email_k8s_notifications",
		RouteConfig: userConfig,
	}}
	PatchRoutes(userRoutes, route)

	require.Len(t, userRoutes, 1)
	p := map[string]interface{}{}
	_ = json.Unmarshal(userRoutes[0].RouteConfig, &p)
	fmt.Printf("p: %+v\n", p)
	require.Equal(t, p["listener_id"], "listener-from-user-route")
	require.Equal(t, p["filtering"].(map[string]interface{})["policy"], "any")
	require.Len(t, p["filtering"].(map[string]interface{})["filters"].([]interface{})[0].(map[string]interface{})["matchers"], 1)
	require.Equal(t, p["filtering"].(map[string]interface{})["filters"].([]interface{})[0].(map[string]interface{})["matchers"].([]interface{})[0].(map[string]interface{})["value"], "user@example.com")
}

func TestPatchJSON(t *testing.T) {
	target := map[string]any{
		"b": map[string]any{"c": "c"},
	}
	patch := map[string]any{
		"a": "a",
		"b": map[string]any{
			"c": "c",
			"d": "d",
		},
	}

	expected := map[string]any{
		"a": "a",
		"b": map[string]any{
			"c": "c",
			"d": "d",
		},
	}

	_, err := patchJSON(target, patch)
	require.NoError(t, err)
	require.Equal(t, expected, target)
}

func TestEmailSinkConfigToJson(t *testing.T) {
	sinker, err := sinks.NewEmailSinkConfig(viper.New(), "id")
	require.NoError(t, err)
	res, err := toJson(*sinker)
	require.NoError(t, err)
	require.Equal(t, `{"auth_type":"auth_login","direct_notifications":false,"from":"nvidia-runai@nvidia.com","from_display_name":"NVIDIA Run:ai","logo_url":"https://s3.amazonaws.com/www.run.ai/OS_NVIDIA_Runai_black.svg","password":"","recipients":[],"smtp_host":"","smtp_port":587,"smtp_tls_enabled":true,"subject":"NVIDIA Run:ai notification - {{ .category }} {{ .reason }}","template_file":"generic_notification.tpl","templates_folder":"sinks_templates/email","use_attachments":false,"user":""}`, string(res)) //nolint:lll // Ignoring long line warning for mock
}

const testCreateListenersYAML = `listeners:
  listener1:
    type: kube_events
    relevant_namespaces:
      - runai-.*
`

func testPostgresProviderWithListeners(t *testing.T, yaml string, dao *mocks.ConfigProviderDao) *ConfigurationProvider {
	t.Helper()
	f := config.CreateTempConfigFile(yaml)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	cfg := service.NewConfig(f.Name())
	return &ConfigurationProvider{Config: *cfg, dao: dao}
}

func TestCreateListeners_CreateSucceeds(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	dao.On("CreateListener", "listener1", "kube_events", mock.AnythingOfType("datatypes.JSON")).Return(nil).Once()
	createListeners(testPostgresProviderWithListeners(t, testCreateListenersYAML, dao))
}

func TestCreateListeners_DuplicateThenUpdateSucceeds(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	dao.On("CreateListener", "listener1", "kube_events", mock.AnythingOfType("datatypes.JSON")).
		Return(configprovider.NewError("listener with same id already exists", configprovider.ErrCodeDuplicate)).Once()
	dao.On("UpdateListener", "listener1", "kube_events", mock.AnythingOfType("datatypes.JSON")).Return(nil).Once()
	createListeners(testPostgresProviderWithListeners(t, testCreateListenersYAML, dao))
}

func TestCreateListeners_DuplicateThenUpdateFails(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	dao.On("CreateListener", "listener1", "kube_events", mock.AnythingOfType("datatypes.JSON")).
		Return(configprovider.NewError("listener with same id already exists", configprovider.ErrCodeDuplicate)).Once()
	dao.On("UpdateListener", "listener1", "kube_events", mock.AnythingOfType("datatypes.JSON")).
		Return(errors.New("update failed")).Once()
	createListeners(testPostgresProviderWithListeners(t, testCreateListenersYAML, dao))
}

func TestCreateListeners_CreateNonDuplicateError(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	dao.On("CreateListener", "listener1", "kube_events", mock.AnythingOfType("datatypes.JSON")).
		Return(errors.New("db unavailable")).Once()
	createListeners(testPostgresProviderWithListeners(t, testCreateListenersYAML, dao))
}

func TestCreateListeners_CreateDaoErrorWrongCode(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	dao.On("CreateListener", "listener1", "kube_events", mock.AnythingOfType("datatypes.JSON")).
		Return(configprovider.NewError("not found", configprovider.ErrCodeNotFound)).Once()
	createListeners(testPostgresProviderWithListeners(t, testCreateListenersYAML, dao))
}

func TestCreateListeners_TwoListenersBothCreate(t *testing.T) {
	yaml := `listeners:
  listener1:
    type: kube_events
    relevant_namespaces: [a]
  listener2:
    type: kube_events
    relevant_namespaces: [b]
`
	dao := mocks.NewConfigProviderDao(t)
	dao.On("CreateListener", mock.AnythingOfType("string"), "kube_events", mock.AnythingOfType("datatypes.JSON")).
		Return(nil).Times(2)
	createListeners(testPostgresProviderWithListeners(t, yaml, dao))
}

func TestCreateSinks_PostgresSink(t *testing.T) {
	f := config.CreateTempConfigFile(`sinks:
  pg_store:
    type: postgres
    host: pg.example
    port: "5432"
    database: notifications
    schema: svc
    table: events
    user: app
    password: secret
    ssl_mode: require
`)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	dao := mocks.NewConfigProviderDao(t)
	cfg := service.NewConfig(f.Name())
	provider := &ConfigurationProvider{Config: *cfg, dao: dao}

	expected := datatypes.JSON(`{"connection_max_idle_time_minutes":5,"connection_max_lifetime_minutes":30,"database":"notifications","host":"pg.example","max_idle_connections":5,"max_open_connections":25,"password":"secret","port":"5432","schema":"svc","ssl_mode":"require","table":"events","user":"app"}`)
	dao.On("UpsertSink", "pg_store", sinks.PostgresSinkType, expected).Return(nil).Once()

	createSinks(provider)
}

func TestCreateSinks_HistorySink(t *testing.T) {
	f := config.CreateTempConfigFile(`history:
  enabled: true
sinks:
  my_history_sink:
    type: history
`)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	dao := mocks.NewConfigProviderDao(t)
	cfg := service.NewConfig(f.Name())
	provider := &ConfigurationProvider{Config: *cfg, dao: dao}

	dao.On("UpsertSink", "my_history_sink", sinks.HistorySinkType, datatypes.JSON(`{"history_enabled":true}`)).Return(nil).Once()

	createSinks(provider)
}

func TestCreateSinks_HistorySinkDisabled(t *testing.T) {
	f := config.CreateTempConfigFile(`sinks:
  my_history_sink:
    type: history
`)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	dao := mocks.NewConfigProviderDao(t)
	cfg := service.NewConfig(f.Name())
	provider := &ConfigurationProvider{Config: *cfg, dao: dao}

	dao.On("UpsertSink", "my_history_sink", sinks.HistorySinkType, datatypes.JSON(`{"history_enabled":false}`)).Return(nil).Once()

	createSinks(provider)
}

func TestGetSinkProvider_OnlyHistoryRequested(t *testing.T) {
	f := config.CreateTempConfigFile(`app:
  enable_history: true
`)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{Config: *service.NewConfig(f.Name()), dao: dao}

	dao.On("FetchSinks", sinks.HistorySinkType).Return([]*models.SinkConfig{{
		SinkId: sinks.HistorySinkType, Type: sinks.HistorySinkType, Config: datatypes.JSON(`{}`),
	}}, nil).Once()

	sinkProvider, err := provider.GetSinkProvider(nil, sinks.HistorySinkType)
	require.NoError(t, err)

	sinkMap, err := sinkProvider.GetSinks()
	require.NoError(t, err)

	entry, ok := sinkMap[sinks.HistorySinkType]
	require.True(t, ok)
	require.NotNil(t, entry)
	require.NotNil(t, entry.History)
}

func TestGetSinkProvider_WithHistoryFromDaoCall(t *testing.T) {
	f := config.CreateTempConfigFile(`app:
  enable_history: true
`)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{Config: *service.NewConfig(f.Name()), dao: dao}

	dao.On("FetchSinks", "some_id", sinks.HistorySinkType).Return([]*models.SinkConfig{
		{SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{}`)},
		{SinkId: sinks.HistorySinkType, Type: sinks.HistorySinkType, Config: datatypes.JSON(`{}`)},
	}, nil).Once()

	sinkProvider, err := provider.GetSinkProvider(nil, "some_id", sinks.HistorySinkType)
	require.NoError(t, err)

	sinksMap, err := sinkProvider.GetSinks()
	require.NoError(t, err)

	entry, ok := sinksMap[sinks.HistorySinkType]
	require.True(t, ok)
	require.NotNil(t, entry)
	require.NotNil(t, entry.History)
	require.Contains(t, sinksMap, "some_id")
}

func TestGetSinkProvider_HistoryDisabled_CarriesDisabledFlag(t *testing.T) {
	f := config.CreateTempConfigFile(`app:
  enable_history: false
`)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{Config: *service.NewConfig(f.Name()), dao: dao}

	dao.On("FetchSinks", sinks.HistorySinkType).Return([]*models.SinkConfig{{
		SinkId: sinks.HistorySinkType, Type: sinks.HistorySinkType, Config: datatypes.JSON(`{}`),
	}}, nil).Once()

	sinkProvider, err := provider.GetSinkProvider(nil, sinks.HistorySinkType)
	require.NoError(t, err)

	sinkMap, err := sinkProvider.GetSinks()
	require.NoError(t, err)

	entry := sinkMap[sinks.HistorySinkType]
	require.NotNil(t, entry)
	require.NotNil(t, entry.History)
	require.False(t, entry.History.HistoryEnabled)
}

func TestGetSinkProvider_NoHistoryRequested(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchSinks", "some_id").Return([]*models.SinkConfig{{
		SinkId: "some_id", Type: "email", Config: datatypes.JSON(`{}`),
	}}, nil).Once()

	sinkProvider, err := provider.GetSinkProvider(nil, "some_id")
	require.NoError(t, err)

	sinksMap, err := sinkProvider.GetSinks()
	require.NoError(t, err)

	require.NotContains(t, sinksMap, sinks.HistorySinkType)
	require.Contains(t, sinksMap, "some_id")
}

func TestGetSinkProvider_UnknownIdRequested(t *testing.T) {
	dao := mocks.NewConfigProviderDao(t)
	provider := ConfigurationProvider{dao: dao}

	dao.On("FetchSinks", "does_not_exist").Return([]*models.SinkConfig{}, nil).Once()

	sinkProvider, err := provider.GetSinkProvider(nil, "does_not_exist")
	require.NoError(t, err)

	sinksMap, err := sinkProvider.GetSinks()
	require.NoError(t, err)

	require.NotContains(t, sinksMap, sinks.HistorySinkType)
	require.NotContains(t, sinksMap, "does_not_exist")
}
