package inference_workloads

import (
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// ServingPortAuthFields — empty AuthorizationType means public (caller skips auth block + sidecar).
type ServingPortAuthFields struct {
	AuthorizationType string
	AuthorizedUsers   string
	AuthorizedGroups  string
}

func (a ServingPortAuthFields) IsPublic() bool {
	return a.AuthorizationType == ""
}

// ResolveServingPortAuth maps the InfW auth enum onto the WLR AuthorizationType.
func ResolveServingPortAuth(src *workloadsApi.ServingPort) (ServingPortAuthFields, *workloadsApi.Error) {
	if src == nil {
		return ServingPortAuthFields{}, nil
	}

	users := composer.NonEmptyStrings(src.AuthorizedUsers)
	groups := composer.NonEmptyStrings(src.AuthorizedGroups)

	if !src.AuthorizationType.IsSet() || src.AuthorizationType.Get() == nil {
		// Reject an allowlist without an authType — silently treating it as public would be unsafe.
		if len(users) > 0 || len(groups) > 0 {
			return ServingPortAuthFields{}, workloadsApi.NewError(
				http.StatusBadRequest,
				"servingPort.authorizationType must be provided when authorizedUsers or authorizedGroups are set",
			)
		}
		return ServingPortAuthFields{}, nil
	}

	authTypeStr := *src.AuthorizationType.Get()

	switch authTypeStr {
	case "public", "":
		return ServingPortAuthFields{}, nil

	case "authenticatedUsers":
		return ServingPortAuthFields{AuthorizationType: "authenticatedUsers"}, nil

	case "authorizedUsers":
		if len(users) == 0 {
			return ServingPortAuthFields{}, workloadsApi.NewError(
				http.StatusBadRequest,
				"servingPort.authorizedUsers must be non-empty when authorizationType=authorizedUsers",
			)
		}
		return ServingPortAuthFields{
			AuthorizationType: "authorizedUsers",
			AuthorizedUsers:   strings.Join(users, ","),
		}, nil

	case "authorizedGroups":
		if len(groups) == 0 {
			return ServingPortAuthFields{}, workloadsApi.NewError(
				http.StatusBadRequest,
				"servingPort.authorizedGroups must be non-empty when authorizationType=authorizedGroups",
			)
		}
		return ServingPortAuthFields{
			AuthorizationType: "authorizedGroups",
			AuthorizedGroups:  strings.Join(groups, ","),
		}, nil

	case "authorizedUsersOrGroups":

		switch {
		case len(users) > 0 && len(groups) > 0:
			return ServingPortAuthFields{}, workloadsApi.NewError(
				http.StatusBadRequest,
				"servingPort.authorizedUsersOrGroups: authorizedUsers and authorizedGroups are mutually exclusive",
			)
		case len(users) > 0:
			return ServingPortAuthFields{
				AuthorizationType: "authorizedUsers",
				AuthorizedUsers:   strings.Join(users, ","),
			}, nil
		case len(groups) > 0:
			return ServingPortAuthFields{
				AuthorizationType: "authorizedGroups",
				AuthorizedGroups:  strings.Join(groups, ","),
			}, nil
		default:
			return ServingPortAuthFields{}, workloadsApi.NewError(
				http.StatusBadRequest,
				"servingPort.authorizedUsersOrGroups: either authorizedUsers or authorizedGroups must be provided",
			)
		}

	default:
		return ServingPortAuthFields{}, workloadsApi.NewError(
			http.StatusBadRequest,
			"servingPort.authorizationType: unsupported value '"+string(authTypeStr)+"'",
		)
	}
}
