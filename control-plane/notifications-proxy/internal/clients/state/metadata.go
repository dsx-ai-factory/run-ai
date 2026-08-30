package state

import (
	"maps"
	"strconv"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"k8s.io/utils/ptr"
)

const (
	EmailEnabled       = "email_enabled"
	SlackEnabled       = "slack_enabled"
	SlackWorkspaceName = "slack_workspace_name"
)

func MergeMetadata(currentMetadata, metadataToUpdate map[string]string) map[string]string {
	if currentMetadata == nil {
		currentMetadata = map[string]string{}
	}
	maps.Copy(currentMetadata, metadataToUpdate)
	return currentMetadata
}

func getRouteMetadata(route *types.BasicRouteConfig) map[string]string {
	if route == nil {
		return nil
	}
	return route.Metadata
}

func FromState(states []notification_state.Stater) map[string]string {
	m := map[string]string{}
	if states == nil {
		return m
	}

	for _, s := range states {
		switch v := s.(type) {
		case *EmailState:
			if v.Enabled != nil {
				m[EmailEnabled] = strconv.FormatBool(*v.Enabled)
			}
		case *SlackState:
			if v.Enabled == nil {
				continue
			}
			m[SlackEnabled] = strconv.FormatBool(*v.Enabled)
			if v.WorkspaceName != nil {
				m[SlackWorkspaceName] = *v.WorkspaceName
			}
		}
	}

	return m
}

func ToState(metadata map[string]string) []notification_state.Stater {
	return []notification_state.Stater{
		ToEmailState(metadata),
		ToSlackState(metadata),
	}
}

func ToEmailState(metadata map[string]string) notification_state.Stater {
	return &EmailState{
		// if key in metadata doesn't exist, then pointer is nil
		Enabled: getOptionalBool(metadata, EmailEnabled),
	}
}

func ToSlackState(metadata map[string]string) notification_state.Stater {
	return &SlackState{
		// if key in metadata doesn't exist, then pointer is nil
		Enabled:       getOptionalBool(metadata, SlackEnabled),
		WorkspaceName: getOptionalString(metadata, SlackWorkspaceName),
	}
}

func ToDetailedState(metadata map[string]string, slackAppConfigured bool) []notification_state.Stater {
	return []notification_state.Stater{
		ToEmailState(metadata),
		ToDetailedSlackState(metadata, slackAppConfigured),
	}
}

func ToDetailedSlackState(metadata map[string]string, slackAppConfigured bool) notification_state.Stater {
	return &SlackState{
		Enabled:                  getOptionalBool(metadata, SlackEnabled),
		WorkspaceName:            getOptionalString(metadata, SlackWorkspaceName),
		AppCredentialsConfigured: ptr.To(slackAppConfigured),
	}
}

func getOptionalBool(m map[string]string, key string) *bool {
	if m == nil {
		return nil
	}

	v, ok := m[key]
	if !ok {
		return nil
	}
	b, _ := strconv.ParseBool(v)
	return &b
}

func getOptionalString(m map[string]string, key string) *string {
	if m == nil {
		return nil
	}

	v, ok := m[key]
	if !ok {
		return nil
	}
	return &v
}
