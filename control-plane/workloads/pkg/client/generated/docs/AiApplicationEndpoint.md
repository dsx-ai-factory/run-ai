# AiApplicationEndpoint

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Host** | Pointer to **string** | The hostname used to access the endpoint (for example, chat.example.com). This value is typically defined by the exposing Kubernetes resource (such as an Ingress). | [optional] 
**Port** | Pointer to **int32** | The network port exposed by the endpoint. This is the port used when connecting to the workload. | [optional] 
**Path** | Pointer to **string** | The URL path associated with the endpoint (for example, /v1/completions). Combined with host, it forms the full HTTP route to the workload. | [optional] 
**TlsEnabled** | Pointer to **NullableBool** | Indicates whether TLS is enabled for the endpoint. When true, the endpoint expects encrypted connections (typically HTTPS). | [optional] 
**Protocol** | Pointer to **string** | The network protocol used for communication with the endpoint (for example, HTTP or HTTPS). | [optional] 
**Url** | Pointer to **string** | The fully qualified URL through which the workload is accessible. This value represents the complete access address derived from the protocol, host, and path (for example, https://chat.example.com/v1/completions). | [optional] 
**WorkloadId** | Pointer to **string** | The id of the workload associated with this endpoint. | [optional] 

## Methods

### NewAiApplicationEndpoint

`func NewAiApplicationEndpoint() *AiApplicationEndpoint`

NewAiApplicationEndpoint instantiates a new AiApplicationEndpoint object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAiApplicationEndpointWithDefaults

`func NewAiApplicationEndpointWithDefaults() *AiApplicationEndpoint`

NewAiApplicationEndpointWithDefaults instantiates a new AiApplicationEndpoint object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHost

`func (o *AiApplicationEndpoint) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *AiApplicationEndpoint) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *AiApplicationEndpoint) SetHost(v string)`

SetHost sets Host field to given value.

### HasHost

`func (o *AiApplicationEndpoint) HasHost() bool`

HasHost returns a boolean if a field has been set.

### GetPort

`func (o *AiApplicationEndpoint) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *AiApplicationEndpoint) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *AiApplicationEndpoint) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *AiApplicationEndpoint) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetPath

`func (o *AiApplicationEndpoint) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *AiApplicationEndpoint) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *AiApplicationEndpoint) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *AiApplicationEndpoint) HasPath() bool`

HasPath returns a boolean if a field has been set.

### GetTlsEnabled

`func (o *AiApplicationEndpoint) GetTlsEnabled() bool`

GetTlsEnabled returns the TlsEnabled field if non-nil, zero value otherwise.

### GetTlsEnabledOk

`func (o *AiApplicationEndpoint) GetTlsEnabledOk() (*bool, bool)`

GetTlsEnabledOk returns a tuple with the TlsEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTlsEnabled

`func (o *AiApplicationEndpoint) SetTlsEnabled(v bool)`

SetTlsEnabled sets TlsEnabled field to given value.

### HasTlsEnabled

`func (o *AiApplicationEndpoint) HasTlsEnabled() bool`

HasTlsEnabled returns a boolean if a field has been set.

### SetTlsEnabledNil

`func (o *AiApplicationEndpoint) SetTlsEnabledNil(b bool)`

 SetTlsEnabledNil sets the value for TlsEnabled to be an explicit nil

### UnsetTlsEnabled
`func (o *AiApplicationEndpoint) UnsetTlsEnabled()`

UnsetTlsEnabled ensures that no value is present for TlsEnabled, not even an explicit nil
### GetProtocol

`func (o *AiApplicationEndpoint) GetProtocol() string`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *AiApplicationEndpoint) GetProtocolOk() (*string, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *AiApplicationEndpoint) SetProtocol(v string)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *AiApplicationEndpoint) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### GetUrl

`func (o *AiApplicationEndpoint) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *AiApplicationEndpoint) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *AiApplicationEndpoint) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *AiApplicationEndpoint) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetWorkloadId

`func (o *AiApplicationEndpoint) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *AiApplicationEndpoint) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *AiApplicationEndpoint) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.

### HasWorkloadId

`func (o *AiApplicationEndpoint) HasWorkloadId() bool`

HasWorkloadId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


