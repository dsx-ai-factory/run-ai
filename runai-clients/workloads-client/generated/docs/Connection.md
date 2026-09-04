# Connection

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**ToolType** | **string** |  | 
**ConnectionType** | **string** |  | 
**Url** | Pointer to **string** |  | [optional] 
**AuthorizationType** | Pointer to **string** | Specifies who can access the connection URL: - &#x60;authenticatedUsers&#x60;: Any authenticated user or service account can access the URL; the authorizedUsers and authorizedGroups fields are ignored. - &#x60;authorizedUsers&#x60;: Only users listed in the authorizedUsers field are allowed to access the URL; the authorizedGroups field is ignored. - &#x60;authorizedGroups&#x60;: Only members of user groups listed in the authorizedGroups field are allowed to access the URL; the authorizedUsers field is ignored. If not specified, authorization is determined by whether authorizedUsers or authorizedGroups is present. If both fields are set, this results in an error. If neither is set, any authenticatedUser can access.  | [optional] 
**AuthorizedUsers** | Pointer to **[]string** | List of users or service accounts that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**AuthorizedGroups** | Pointer to **[]string** |  | [optional] 
**ContainerPort** | Pointer to **int32** | The port that the container running the workload exposes. | [optional] 
**CorrelationName** | Pointer to **string** | A given name for the url/port, used for correlating cluster and control-plane information. | [optional] 
**ExposedUrlSubDomain** | Pointer to **NullableString** | for cluster supporting sub-domain, stores the sub-domain of the exposedUrl. e.g, myproject-myworkload-url-1., t when syncing from the cluster, this subdomain is patched woth the cluster domain to compose the complete connection path | [optional] 
**ExposedUrlPath** | Pointer to **NullableString** | for cluster not supporting sub-domain, stores the path of the exposedUrl. e.g. /myproject/myworkload/url-1 when syncing from the cluster, this path is patched with the cluster domain to compose the complete connection path | [optional] 
**PortResourceName** | Pointer to **NullableString** | stores the assigned name to the port resource in the cluster. e.g. workspace-nodeport-0 when syncing from the cluster, this name is used to match the connection with it&#39;s connection url. | [optional] 
**ServiceResourceName** | Pointer to **NullableString** | stores the assigned resource name to the service holding this port in the cluster. e.g. myworkload-service-0 when syncing from the cluster, this name, along with portResourceName, is used to match the connection with it&#39;s connection url | [optional] 

## Methods

### NewConnection

`func NewConnection(name string, toolType string, connectionType string, ) *Connection`

NewConnection instantiates a new Connection object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectionWithDefaults

`func NewConnectionWithDefaults() *Connection`

NewConnectionWithDefaults instantiates a new Connection object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Connection) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Connection) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Connection) SetName(v string)`

SetName sets Name field to given value.


### GetToolType

`func (o *Connection) GetToolType() string`

GetToolType returns the ToolType field if non-nil, zero value otherwise.

### GetToolTypeOk

`func (o *Connection) GetToolTypeOk() (*string, bool)`

GetToolTypeOk returns a tuple with the ToolType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolType

`func (o *Connection) SetToolType(v string)`

SetToolType sets ToolType field to given value.


### GetConnectionType

`func (o *Connection) GetConnectionType() string`

GetConnectionType returns the ConnectionType field if non-nil, zero value otherwise.

### GetConnectionTypeOk

`func (o *Connection) GetConnectionTypeOk() (*string, bool)`

GetConnectionTypeOk returns a tuple with the ConnectionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionType

`func (o *Connection) SetConnectionType(v string)`

SetConnectionType sets ConnectionType field to given value.


### GetUrl

`func (o *Connection) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *Connection) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *Connection) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *Connection) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetAuthorizationType

`func (o *Connection) GetAuthorizationType() string`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *Connection) GetAuthorizationTypeOk() (*string, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *Connection) SetAuthorizationType(v string)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *Connection) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### GetAuthorizedUsers

`func (o *Connection) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *Connection) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *Connection) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *Connection) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### GetAuthorizedGroups

`func (o *Connection) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *Connection) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *Connection) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *Connection) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### GetContainerPort

`func (o *Connection) GetContainerPort() int32`

GetContainerPort returns the ContainerPort field if non-nil, zero value otherwise.

### GetContainerPortOk

`func (o *Connection) GetContainerPortOk() (*int32, bool)`

GetContainerPortOk returns a tuple with the ContainerPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainerPort

`func (o *Connection) SetContainerPort(v int32)`

SetContainerPort sets ContainerPort field to given value.

### HasContainerPort

`func (o *Connection) HasContainerPort() bool`

HasContainerPort returns a boolean if a field has been set.

### GetCorrelationName

`func (o *Connection) GetCorrelationName() string`

GetCorrelationName returns the CorrelationName field if non-nil, zero value otherwise.

### GetCorrelationNameOk

`func (o *Connection) GetCorrelationNameOk() (*string, bool)`

GetCorrelationNameOk returns a tuple with the CorrelationName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCorrelationName

`func (o *Connection) SetCorrelationName(v string)`

SetCorrelationName sets CorrelationName field to given value.

### HasCorrelationName

`func (o *Connection) HasCorrelationName() bool`

HasCorrelationName returns a boolean if a field has been set.

### GetExposedUrlSubDomain

`func (o *Connection) GetExposedUrlSubDomain() string`

GetExposedUrlSubDomain returns the ExposedUrlSubDomain field if non-nil, zero value otherwise.

### GetExposedUrlSubDomainOk

`func (o *Connection) GetExposedUrlSubDomainOk() (*string, bool)`

GetExposedUrlSubDomainOk returns a tuple with the ExposedUrlSubDomain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrlSubDomain

`func (o *Connection) SetExposedUrlSubDomain(v string)`

SetExposedUrlSubDomain sets ExposedUrlSubDomain field to given value.

### HasExposedUrlSubDomain

`func (o *Connection) HasExposedUrlSubDomain() bool`

HasExposedUrlSubDomain returns a boolean if a field has been set.

### SetExposedUrlSubDomainNil

`func (o *Connection) SetExposedUrlSubDomainNil(b bool)`

 SetExposedUrlSubDomainNil sets the value for ExposedUrlSubDomain to be an explicit nil

### UnsetExposedUrlSubDomain
`func (o *Connection) UnsetExposedUrlSubDomain()`

UnsetExposedUrlSubDomain ensures that no value is present for ExposedUrlSubDomain, not even an explicit nil
### GetExposedUrlPath

`func (o *Connection) GetExposedUrlPath() string`

GetExposedUrlPath returns the ExposedUrlPath field if non-nil, zero value otherwise.

### GetExposedUrlPathOk

`func (o *Connection) GetExposedUrlPathOk() (*string, bool)`

GetExposedUrlPathOk returns a tuple with the ExposedUrlPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrlPath

`func (o *Connection) SetExposedUrlPath(v string)`

SetExposedUrlPath sets ExposedUrlPath field to given value.

### HasExposedUrlPath

`func (o *Connection) HasExposedUrlPath() bool`

HasExposedUrlPath returns a boolean if a field has been set.

### SetExposedUrlPathNil

`func (o *Connection) SetExposedUrlPathNil(b bool)`

 SetExposedUrlPathNil sets the value for ExposedUrlPath to be an explicit nil

### UnsetExposedUrlPath
`func (o *Connection) UnsetExposedUrlPath()`

UnsetExposedUrlPath ensures that no value is present for ExposedUrlPath, not even an explicit nil
### GetPortResourceName

`func (o *Connection) GetPortResourceName() string`

GetPortResourceName returns the PortResourceName field if non-nil, zero value otherwise.

### GetPortResourceNameOk

`func (o *Connection) GetPortResourceNameOk() (*string, bool)`

GetPortResourceNameOk returns a tuple with the PortResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPortResourceName

`func (o *Connection) SetPortResourceName(v string)`

SetPortResourceName sets PortResourceName field to given value.

### HasPortResourceName

`func (o *Connection) HasPortResourceName() bool`

HasPortResourceName returns a boolean if a field has been set.

### SetPortResourceNameNil

`func (o *Connection) SetPortResourceNameNil(b bool)`

 SetPortResourceNameNil sets the value for PortResourceName to be an explicit nil

### UnsetPortResourceName
`func (o *Connection) UnsetPortResourceName()`

UnsetPortResourceName ensures that no value is present for PortResourceName, not even an explicit nil
### GetServiceResourceName

`func (o *Connection) GetServiceResourceName() string`

GetServiceResourceName returns the ServiceResourceName field if non-nil, zero value otherwise.

### GetServiceResourceNameOk

`func (o *Connection) GetServiceResourceNameOk() (*string, bool)`

GetServiceResourceNameOk returns a tuple with the ServiceResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceResourceName

`func (o *Connection) SetServiceResourceName(v string)`

SetServiceResourceName sets ServiceResourceName field to given value.

### HasServiceResourceName

`func (o *Connection) HasServiceResourceName() bool`

HasServiceResourceName returns a boolean if a field has been set.

### SetServiceResourceNameNil

`func (o *Connection) SetServiceResourceNameNil(b bool)`

 SetServiceResourceNameNil sets the value for ServiceResourceName to be an explicit nil

### UnsetServiceResourceName
`func (o *Connection) UnsetServiceResourceName()`

UnsetServiceResourceName ensures that no value is present for ServiceResourceName, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


