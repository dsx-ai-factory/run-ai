# PVCListInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**ScopeType** | **string** | the scope type of the resource. | 
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 
**Name** | **string** | name of the pvc. | 
**Origin** | [**NullableOrigin**](Origin.md) |  | 

## Methods

### NewPVCListInfo

`func NewPVCListInfo(clusterId string, scopeType string, scopeId string, name string, origin NullableOrigin, ) *PVCListInfo`

NewPVCListInfo instantiates a new PVCListInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCListInfoWithDefaults

`func NewPVCListInfoWithDefaults() *PVCListInfo`

NewPVCListInfoWithDefaults instantiates a new PVCListInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *PVCListInfo) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PVCListInfo) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PVCListInfo) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetScopeType

`func (o *PVCListInfo) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *PVCListInfo) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *PVCListInfo) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *PVCListInfo) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *PVCListInfo) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *PVCListInfo) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetName

`func (o *PVCListInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PVCListInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PVCListInfo) SetName(v string)`

SetName sets Name field to given value.


### GetOrigin

`func (o *PVCListInfo) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *PVCListInfo) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *PVCListInfo) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.


### SetOriginNil

`func (o *PVCListInfo) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *PVCListInfo) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


