# ProbeHandler

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**HttpGet** | Pointer to [**ProbeHandlerHttpGet**](ProbeHandlerHttpGet.md) |  | [optional] 

## Methods

### NewProbeHandler

`func NewProbeHandler() *ProbeHandler`

NewProbeHandler instantiates a new ProbeHandler object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProbeHandlerWithDefaults

`func NewProbeHandlerWithDefaults() *ProbeHandler`

NewProbeHandlerWithDefaults instantiates a new ProbeHandler object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHttpGet

`func (o *ProbeHandler) GetHttpGet() ProbeHandlerHttpGet`

GetHttpGet returns the HttpGet field if non-nil, zero value otherwise.

### GetHttpGetOk

`func (o *ProbeHandler) GetHttpGetOk() (*ProbeHandlerHttpGet, bool)`

GetHttpGetOk returns a tuple with the HttpGet field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHttpGet

`func (o *ProbeHandler) SetHttpGet(v ProbeHandlerHttpGet)`

SetHttpGet sets HttpGet field to given value.

### HasHttpGet

`func (o *ProbeHandler) HasHttpGet() bool`

HasHttpGet returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


