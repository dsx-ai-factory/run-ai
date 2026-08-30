package composer

import (
	"fmt"
	"net"
	"net/http"
	url2 "net/url"
	"regexp"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	PortServiceType_SERVING_PORT = "ServingPort"
	CustomToolType               = "custom"
	TensorboardToolType          = "tensorboard"

	UnspecifiedAuthType = ""

	ConnectionAuthorizationType_PUBLIC                     = "public"
	ConnectionAuthorizationType_AUTHENTICATED_USERS        = "authenticatedUsers"
	ConnectionAuthorizationType_AUTHORIZED_USERS           = "authorizedUsers"
	ConnectionAuthorizationType_AUTHORIZED_GROUPS          = "authorizedGroups"
	ConnectionAuthorizationType_AUTHORIZED_USERS_OR_GROUPS = "authorizedUsersOrGroups"
)

var AuthTypesServingPortInference = datastructs.NewSet([]string{
	ConnectionAuthorizationType_PUBLIC,
	ConnectionAuthorizationType_AUTHENTICATED_USERS,
	ConnectionAuthorizationType_AUTHORIZED_USERS,
	ConnectionAuthorizationType_AUTHORIZED_GROUPS,
	ConnectionAuthorizationType_AUTHORIZED_USERS_OR_GROUPS, // this is the old way
}...)

var AuthTypesServingPortDinf = datastructs.NewSet([]string{
	ConnectionAuthorizationType_PUBLIC,
	ConnectionAuthorizationType_AUTHENTICATED_USERS,
	ConnectionAuthorizationType_AUTHORIZED_USERS,
	ConnectionAuthorizationType_AUTHORIZED_GROUPS,
}...)

var AuthTypesExposedUrls = datastructs.NewSet([]string{
	ConnectionAuthorizationType_AUTHENTICATED_USERS,
	ConnectionAuthorizationType_AUTHORIZED_USERS,
	ConnectionAuthorizationType_AUTHORIZED_GROUPS,
}...)

var ignoreAccessList = datastructs.NewSet([]string{
	ConnectionAuthorizationType_PUBLIC,
	ConnectionAuthorizationType_AUTHENTICATED_USERS,
}...)

func (fc *FieldsComposer) ComposePorts(prefix string, src []workloadsApi.Port,
	target any, setter FnMapToPortSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		// Mirror the ExposedUrlInfo pattern: the first setter argument carries
		// both the items and the fc-derived context (workload name, resource
		// prefix, cluster config) that the workloads-v2 networking setter needs to
		// build Service names/selectors. The cookie stays free for setter-specific
		// data such as a custom label selector.
		info := &PortsInfo{
			Items:          make(map[string]*workloadsApi.Port, len(src)),
			WorkloadName:   fc.WorkloadName,
			ResourcePrefix: fc.ResourcePrefix,
			Config:         fc.ClusterStorageConfig,
		}

		for _, entry := range src {
			urlItemId := fc.ItemizedNumbering.Compose(prefix, entry.Name.Get(), nil, false)

			if convutils.BoolIsTrue(entry.Exclude.Get()) {
				info.Items[urlItemId] = nil
				continue
			}

			info.Items[urlItemId] = &entry
		}

		return setter(info, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeServingPort(replicaType *workloadsApi.ReplicaType, src *workloadsApi.SupersetServingPort,
	target any, setter FnServingPortSetter, cookie any, portAuthInfo *PortAuthInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			if fc.WorkloadType == string(workloadsApi.WORKLOADTYPE_INFERENCE) {
				if !fc.Options.IgnoreMissingValues && IsMainNode(replicaType) {
					return workloadsApi.NewError(http.StatusBadRequest, "mandatory servingPort is not provided")
				}
			}
			return nil
		}

		itemId := fc.ItemizedNumbering.Compose(itemized.ServingPortPrefix, nil, nil, false)

		result := ServingPortInfo{
			ItemId:      itemId,
			PortInfo:    src,
			ServiceType: convutils.Ref(PortServiceType_SERVING_PORT),
			ReplicaType: replicaType,
		}

		switch fc.WorkloadType {
		case string(workloadsApi.WORKLOADTYPE_INFERENCE):
			if !convutils.HasValue(src.Container) && !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "servingPort: mandatory container is not provided")
			}
		case string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE):
			//
			//   distributed inference supports exposure of external url
			//
			result.ExposedUrl = src.ExposedUrl.Get()
			result.ExposedExternally = src.ExposeExternally.Get()

			if convutils.IsStringHasValue(src.ExposedUrl.Get()) {
				exposeExternally := convutils.Deref(src.ExposeExternally.Get(), true)
				if !exposeExternally {
					return workloadsApi.NewError(http.StatusBadRequest, "servingPort: exposedUrl should not be set when exposeExternally is false")
				}
			}

			result.PortInfo.Container.Set(src.Port.Get())
		}

		result.AuthInfo = PortAuthInfo{
			AuthorizationType: src.AuthorizationType.Get(),
			AuthorizedUsers:   src.AuthorizedUsers,
			AuthorizedGroups:  src.AuthorizedGroups,
		}

		authTypes := AuthTypesServingPortInference
		if fc.WorkloadType == string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE) {
			authTypes = AuthTypesServingPortDinf
		}

		authErr := ValidatePortAuth(&result.AuthInfo, authTypes, ConnectionAuthorizationType_PUBLIC,
			"servingPort", fc.Options.IgnoreMissingValues)
		if authErr != nil {
			return authErr
		}

		if fc.Options.NormalizeServingPortAuthType {
			// the cluster CRD does not support the new values authorizedGroups and authorizedUsers
			// so now that we know that only one of them is given we just need to set the
			// type to users-and-groups, the cluster will know what to do with it
			if convutils.IsStringHasValue(result.AuthInfo.AuthorizationType) {
				authType := *result.AuthInfo.AuthorizationType
				if authType == ConnectionAuthorizationType_AUTHORIZED_GROUPS || authType == ConnectionAuthorizationType_AUTHORIZED_USERS {
					authType = ConnectionAuthorizationType_AUTHORIZED_USERS_OR_GROUPS
				}

				result.AuthInfo.AuthorizationType = convutils.Ref(authType)
			}
		}

		err := setter(&result, target, cookie)
		if err != nil {
			return err
		}

		*portAuthInfo = result.AuthInfo

		return nil
	}
}

func (fc *FieldsComposer) ComposeServingPortAuthType(src *string, authInfo *PortAuthInfo,
	target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if fc.WorkloadType != string(workloadsApi.WORKLOADTYPE_INFERENCE) {
			return nil
		}

		// the cluster CRD does not support the new values authorizedGroups and authorizedUsers
		// so now that we know that only one of them is given we just need to set the
		// type to users-and-groups, the cluster will know what to do with it
		if !convutils.IsStringHasValue(src) {
			return nil
		}

		authType := *authInfo.AuthorizationType
		if authType == ConnectionAuthorizationType_AUTHORIZED_GROUPS || authType == ConnectionAuthorizationType_AUTHORIZED_USERS {
			authType = ConnectionAuthorizationType_AUTHORIZED_USERS_OR_GROUPS
		}

		return setter(&authType, target, cookie)
	}
}

func ValidatePortAuth(
	portAuth *PortAuthInfo,
	supportedTypes datastructs.Set[string],
	defaultAuthType string,
	fldName string,
	ignoreMissingValues bool,
) *workloadsApi.Error {
	//
	//   determine authorization type. note that for non-inference, authorization type may be
	//   unspecified, and in this case - its value is determined by existence of authorizedUsers
	//   or authorizedGroups in the spec
	//
	var authorizationType string
	if convutils.IsStringHasValue(portAuth.AuthorizationType) {
		authorizationType = *portAuth.AuthorizationType
	} else if defaultAuthType != UnspecifiedAuthType {
		authorizationType = defaultAuthType
	}

	if authorizationType != UnspecifiedAuthType {
		if !supportedTypes.Has(authorizationType) {
			return workloadsApi.NewError(http.StatusForbidden, fmt.Sprintf("%s: unsupported authorizationType '%s'", fldName, authorizationType))
		}
	}

	//
	//   format the authorization type for error messages
	//
	var when string
	if authorizationType != UnspecifiedAuthType {
		when = fmt.Sprintf("when authorizationType is %s", authorizationType)
	} else {
		when = "when authorizationType is not provided"
	}

	//
	//   check which access list is specified
	//
	authorizedUsersIsSet := len(portAuth.AuthorizedUsers) > 0
	authorizedGroupsIsSet := len(portAuth.AuthorizedGroups) > 0

	// Validate empty authorized users/groups arrays
	if portAuth.AuthorizedUsers != nil && len(portAuth.AuthorizedUsers) == 0 {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: authorizedUsers cannot be empty", fldName))
	}
	if portAuth.AuthorizedGroups != nil && len(portAuth.AuthorizedGroups) == 0 {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: authorizedGroups cannot be empty", fldName))
	}

	switch authorizationType {
	case ConnectionAuthorizationType_AUTHORIZED_USERS_OR_GROUPS:
		// for type=sers-or-groups, we expect either list of users or list of groups to be provided, but not both
		if authorizedUsersIsSet && authorizedGroupsIsSet {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: authorizedUsers and authorizedGroups are mutually exclusive %s", fldName, when))
		}

		if !ignoreMissingValues {
			if !authorizedUsersIsSet && !authorizedGroupsIsSet {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: authorizedUsers or authorizedGroups should be set %s", fldName, when))
			}
		}

	case ConnectionAuthorizationType_AUTHORIZED_GROUPS:
		// for type=groups, we expect list of groups to be provided. list of users can also be provided, but will be ignored
		if !ignoreMissingValues {
			if !authorizedGroupsIsSet {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: authorizedGroups must be set %s", fldName, when))
			}
		}

	case ConnectionAuthorizationType_AUTHORIZED_USERS:
		// for type=users, we expect list of users to be provided. list of groups can also be provided, but will be ignored
		if !ignoreMissingValues {
			if !authorizedUsersIsSet {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: authorizedUsers must be set %s", fldName, when))
			}
		}

	case ConnectionAuthorizationType_AUTHENTICATED_USERS, ConnectionAuthorizationType_PUBLIC:
		// for public or authenticated users, we do not have any requirements from the workload, it can include list of users and/or groups
		// but we ignore them

	case UnspecifiedAuthType:
		// if not specified, we derive it from the existence of authorizedGroups and authorizedUsers.
		// if none specified, we treat it as "public" port
		if authorizedUsersIsSet && authorizedGroupsIsSet {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: authorizedUsers and authorizedGroups are mutually exclusive %s", fldName, when))
		}

		if !ignoreMissingValues {
			if authorizedGroupsIsSet {
				authorizationType = ConnectionAuthorizationType_AUTHORIZED_GROUPS
			} else if authorizedUsersIsSet {
				authorizationType = ConnectionAuthorizationType_AUTHORIZED_USERS
			} else {
				authorizationType = ConnectionAuthorizationType_PUBLIC
			}
		}
	}

	//
	//  fix authInfo to have the final values for the port authorization
	//
	if authorizationType != UnspecifiedAuthType {
		portAuth.AuthorizationType = &authorizationType
	}

	if ignoreAccessList.Has(authorizationType) {
		portAuth.AuthorizedGroups = nil
		portAuth.AuthorizedUsers = nil
	}

	if authorizationType == ConnectionAuthorizationType_AUTHORIZED_USERS {
		portAuth.AuthorizedGroups = nil
	}

	if authorizationType == ConnectionAuthorizationType_AUTHORIZED_GROUPS {
		portAuth.AuthorizedUsers = nil
	}

	return nil
}

// ComposeExposedUrls composes the exposed-URL items into ExposedUrlInfo and
// invokes the supplied setter. forceSuffix is the per-call control over the
// "first auto URL is suffix-less" behavior: when true (e.g. distributed
// non-primary role) every URL host carries its nick suffix, avoiding a
// collision with the primary role's no-suffix host.
func (fc *FieldsComposer) ComposeExposedUrls(prefix string, src []workloadsApi.ExposedUrl,
	target any, setter FnExposedUrlSetter, cookie any, jupyterContainerPort *int32, infoOut **ExposedUrlInfo,
	forceSuffix bool,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		cfg := fc.ClusterStorageConfig
		if cfg == nil {
			cfg = DefaultClusterStorageConfig()
		}

		info := &ExposedUrlInfo{
			Items:             make(map[string]*ExposedUrlInfoItem, len(src)),
			WorkloadName:      fc.WorkloadName,
			ProjectName:       fc.ProjectName,
			Config:            cfg,
			ResourcePrefix:    fc.ResourcePrefix,
			ForceSuffixInUrls: forceSuffix,
		}

		for idx, entry := range src {
			urlItemId := fc.ItemizedNumbering.Compose(prefix, entry.Name.Get(), nil, false)

			// note that we always consume a port number to maintain URL index alignment,
			// so url-1 always gets port 4181 even if url-0 is excluded.
			currentOAuthPort := int(fc.OauthPortNumber) + idx

			if convutils.BoolIsTrue(entry.Exclude.Get()) {
				info.Items[urlItemId] = nil
				continue
			}

			//
			//   validate authorization by users or groups
			//
			authInfo := PortAuthInfo{
				AuthorizedUsers:  entry.AuthorizedUsers,
				AuthorizedGroups: entry.AuthorizedGroups,
			}

			if entry.AuthorizationType.Get() != nil {
				authTypeStr := string(*entry.AuthorizationType.Get())
				authInfo.AuthorizationType = convutils.Ref(authTypeStr)
			}

			err := ValidatePortAuth(&authInfo, AuthTypesExposedUrls, UnspecifiedAuthType, "exposedUrls", fc.Options.IgnoreMissingValues)
			if err != nil {
				return err
			}

			err = ValidateUrlSyntax(entry.Url.Get(), true)
			if err != nil {
				return err
			}

			toolType := convutils.Deref(entry.ToolType.Get(), "")
			if toolType == JupyterToolType {
				*jupyterContainerPort = convutils.Deref(entry.Container.Get(), DefaultJupyterPort)
			}
			// The cluster reserves the no-suffix auto URL for the jupyter/tensorboard
			// tool itself, so any exposed url on such a workload must carry a nick suffix.
			if toolType == JupyterToolType || toolType == TensorboardToolType {
				info.IsJupyterOrTensorboard = true
			}

			item := &ExposedUrlInfoItem{
				ExposedUrl:    &entry,
				Nick:          urlItemId,
				ContainerPort: entry.Container.Get(),
				Url:           convutils.Deref(entry.Url.Get(), ""),
			}

			item.OAuthPort = convutils.Ref(int32(currentOAuthPort))

			if len(authInfo.AuthorizedUsers) > 0 {
				item.AuthorizedUsers = convutils.Ref(strings.Join(authInfo.AuthorizedUsers, ","))
			}
			if len(authInfo.AuthorizedGroups) > 0 {
				item.AuthorizedGroups = convutils.Ref(strings.Join(authInfo.AuthorizedGroups, ","))
			}

			info.Items[urlItemId] = item
		}

		if infoOut != nil {
			*infoOut = info
		}

		return setter(info, target, cookie)
	}
}

// ApplyExposedUrls re-apply the exposed urls setting that was collected by ComposeExposeUrls into
// ExposedUrlInfo, sending the information to another setter for re-apply in other parts of the spec
// this way we collect the information only once and use it with multiple setters.
func (fc *FieldsComposer) ApplyExposedUrls(info **ExposedUrlInfo,
	target any, setter FnExposedUrlSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if info == nil || *info == nil {
			return nil
		}

		return setter(*info, target, cookie)
	}
}

// absoluteUrlRegex is a regex pattern to verify if the URL is absolute url
// example
//
//	https://www.google.com           --> isn't absolute url
//	http://www.google.com/           -->  is absolute url (note the trailing /)
//	https://www.google.com/some/path --> is absolute url (as it has a Path after the domain)
var absoluteUrlRegex = regexp.MustCompile(`^(^[a-zA-Z]+?://[^/]+/.*)$`)

func ValidateUrlSyntax(url *string, absolute bool) *workloadsApi.Error {
	if url == nil {
		return nil
	}

	parsedUrl, err := url2.ParseRequestURI(*url)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("malformed url: %v", err))
	}

	if absolute {
		// Not using parsedUrl.IsAbs as it verifies use of scheme only, absolute path must have a path or a trailing / after the domain
		if !absoluteUrlRegex.MatchString(*url) {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("Invalid url <%s> must be an absolute path", *url))
		}
	}

	address := net.ParseIP(parsedUrl.Host)
	if address == nil {
		if !strings.Contains(parsedUrl.Host, ".") {
			return workloadsApi.NewError(http.StatusBadRequest, "malformed url host")
		}
	}

	return nil
}
