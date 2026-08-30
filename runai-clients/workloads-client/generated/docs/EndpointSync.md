# EndpointSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Host** | Pointer to **string** | The hostname used to access the endpoint (for example, chat.example.com). This value is typically defined by the exposing Kubernetes resource (such as an Ingress). | [optional] 
**Port** | Pointer to **int32** | The network port exposed by the endpoint. This is the port used when connecting to the workload. | [optional] 
**Path** | Pointer to **string** | The URL path associated with the endpoint (for example, /v1/completions). Combined with host, it forms the full HTTP route to the workload. | [optional] 
**TlsEnabled** | Pointer to **NullableBool** | Indicates whether TLS is enabled for the endpoint. When true, the endpoint expects encrypted connections (typically HTTPS). | [optional] 
**Protocol** | Pointer to **string** | The network protocol used for communication with the endpoint (for example, HTTP or HTTPS). | [optional] 
**Url** | Pointer to **string** | The fully qualified URL through which the workload is accessible. This value represents the complete access address derived from the protocol, host, and path (for example, https://chat.example.com/v1/completions). | [optional] 
**SourceRef** | Pointer to [**EndpointSyncAllOfSourceRef**](EndpointSyncAllOfSourceRef.md) |  | [optional] 

## Methods

### NewEndpointSync

`func NewEndpointSync() *EndpointSync`

NewEndpointSync instantiates a new EndpointSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEndpointSyncWithDefaults

`func NewEndpointSyncWithDefaults() *EndpointSync`

NewEndpointSyncWithDefaults instantiates a new EndpointSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHost

`func (o *EndpointSync) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *EndpointSync) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *EndpointSync) SetHost(v string)`

SetHost sets Host field to given value.

### HasHost

`func (o *EndpointSync) HasHost() bool`

HasHost returns a boolean if a field has been set.

### GetPort

`func (o *EndpointSync) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *EndpointSync) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *EndpointSync) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *EndpointSync) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetPath

`func (o *EndpointSync) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *EndpointSync) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *EndpointSync) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *EndpointSync) HasPath() bool`

HasPath returns a boolean if a field has been set.

### GetTlsEnabled

`func (o *EndpointSync) GetTlsEnabled() bool`

GetTlsEnabled returns the TlsEnabled field if non-nil, zero value otherwise.

### GetTlsEnabledOk

`func (o *EndpointSync) GetTlsEnabledOk() (*bool, bool)`

GetTlsEnabledOk returns a tuple with the TlsEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTlsEnabled

`func (o *EndpointSync) SetTlsEnabled(v bool)`

SetTlsEnabled sets TlsEnabled field to given value.

### HasTlsEnabled

`func (o *EndpointSync) HasTlsEnabled() bool`

HasTlsEnabled returns a boolean if a field has been set.

### SetTlsEnabledNil

`func (o *EndpointSync) SetTlsEnabledNil(b bool)`

 SetTlsEnabledNil sets the value for TlsEnabled to be an explicit nil

### UnsetTlsEnabled
`func (o *EndpointSync) UnsetTlsEnabled()`

UnsetTlsEnabled ensures that no value is present for TlsEnabled, not even an explicit nil
### GetProtocol

`func (o *EndpointSync) GetProtocol() string`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *EndpointSync) GetProtocolOk() (*string, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *EndpointSync) SetProtocol(v string)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *EndpointSync) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### GetUrl

`func (o *EndpointSync) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *EndpointSync) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *EndpointSync) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *EndpointSync) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetSourceRef

`func (o *EndpointSync) GetSourceRef() EndpointSyncAllOfSourceRef`

GetSourceRef returns the SourceRef field if non-nil, zero value otherwise.

### GetSourceRefOk

`func (o *EndpointSync) GetSourceRefOk() (*EndpointSyncAllOfSourceRef, bool)`

GetSourceRefOk returns a tuple with the SourceRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceRef

`func (o *EndpointSync) SetSourceRef(v EndpointSyncAllOfSourceRef)`

SetSourceRef sets SourceRef field to given value.

### HasSourceRef

`func (o *EndpointSync) HasSourceRef() bool`

HasSourceRef returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


