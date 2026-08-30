# PatchClusterConfigRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | Pointer to [**ClusterUserConfigDataPartialUpdate**](ClusterUserConfigDataPartialUpdate.md) |  | [optional] 

## Methods

### NewPatchClusterConfigRequest

`func NewPatchClusterConfigRequest() *PatchClusterConfigRequest`

NewPatchClusterConfigRequest instantiates a new PatchClusterConfigRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPatchClusterConfigRequestWithDefaults

`func NewPatchClusterConfigRequestWithDefaults() *PatchClusterConfigRequest`

NewPatchClusterConfigRequestWithDefaults instantiates a new PatchClusterConfigRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *PatchClusterConfigRequest) GetConfig() ClusterUserConfigDataPartialUpdate`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *PatchClusterConfigRequest) GetConfigOk() (*ClusterUserConfigDataPartialUpdate, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *PatchClusterConfigRequest) SetConfig(v ClusterUserConfigDataPartialUpdate)`

SetConfig sets Config field to given value.

### HasConfig

`func (o *PatchClusterConfigRequest) HasConfig() bool`

HasConfig returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


