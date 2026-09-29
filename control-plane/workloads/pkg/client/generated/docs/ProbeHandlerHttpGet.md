# ProbeHandlerHttpGet

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Path to access on the HTTP server, defaults to /. | [optional] 
**Port** | Pointer to **NullableInt32** | Number of the port to access on the container. | [optional] 
**Host** | Pointer to **NullableString** | Host name to connect to, defaults to the pod IP. | [optional] 
**Scheme** | Pointer to [**NullableProbeHandlerScheme**](ProbeHandlerScheme.md) |  | [optional] 

## Methods

### NewProbeHandlerHttpGet

`func NewProbeHandlerHttpGet() *ProbeHandlerHttpGet`

NewProbeHandlerHttpGet instantiates a new ProbeHandlerHttpGet object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProbeHandlerHttpGetWithDefaults

`func NewProbeHandlerHttpGetWithDefaults() *ProbeHandlerHttpGet`

NewProbeHandlerHttpGetWithDefaults instantiates a new ProbeHandlerHttpGet object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *ProbeHandlerHttpGet) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *ProbeHandlerHttpGet) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *ProbeHandlerHttpGet) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *ProbeHandlerHttpGet) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *ProbeHandlerHttpGet) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *ProbeHandlerHttpGet) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetPort

`func (o *ProbeHandlerHttpGet) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *ProbeHandlerHttpGet) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *ProbeHandlerHttpGet) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *ProbeHandlerHttpGet) HasPort() bool`

HasPort returns a boolean if a field has been set.

### SetPortNil

`func (o *ProbeHandlerHttpGet) SetPortNil(b bool)`

 SetPortNil sets the value for Port to be an explicit nil

### UnsetPort
`func (o *ProbeHandlerHttpGet) UnsetPort()`

UnsetPort ensures that no value is present for Port, not even an explicit nil
### GetHost

`func (o *ProbeHandlerHttpGet) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *ProbeHandlerHttpGet) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *ProbeHandlerHttpGet) SetHost(v string)`

SetHost sets Host field to given value.

### HasHost

`func (o *ProbeHandlerHttpGet) HasHost() bool`

HasHost returns a boolean if a field has been set.

### SetHostNil

`func (o *ProbeHandlerHttpGet) SetHostNil(b bool)`

 SetHostNil sets the value for Host to be an explicit nil

### UnsetHost
`func (o *ProbeHandlerHttpGet) UnsetHost()`

UnsetHost ensures that no value is present for Host, not even an explicit nil
### GetScheme

`func (o *ProbeHandlerHttpGet) GetScheme() ProbeHandlerScheme`

GetScheme returns the Scheme field if non-nil, zero value otherwise.

### GetSchemeOk

`func (o *ProbeHandlerHttpGet) GetSchemeOk() (*ProbeHandlerScheme, bool)`

GetSchemeOk returns a tuple with the Scheme field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScheme

`func (o *ProbeHandlerHttpGet) SetScheme(v ProbeHandlerScheme)`

SetScheme sets Scheme field to given value.

### HasScheme

`func (o *ProbeHandlerHttpGet) HasScheme() bool`

HasScheme returns a boolean if a field has been set.

### SetSchemeNil

`func (o *ProbeHandlerHttpGet) SetSchemeNil(b bool)`

 SetSchemeNil sets the value for Scheme to be an explicit nil

### UnsetScheme
`func (o *ProbeHandlerHttpGet) UnsetScheme()`

UnsetScheme ensures that no value is present for Scheme, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


