# GetMappingFieldInfoReq

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadType** | Pointer to [**GVK**](GVK.md) | The Kubernetes group, version, and kind of the workload type whose mapping is queried. | [optional] 
**FieldsRef** | Pointer to [**[]BreakdownFieldRef**](BreakdownFieldRef.md) | The field references to retrieve the detailed breakdown for. | [optional] 

## Methods

### NewGetMappingFieldInfoReq

`func NewGetMappingFieldInfoReq() *GetMappingFieldInfoReq`

NewGetMappingFieldInfoReq instantiates a new GetMappingFieldInfoReq object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetMappingFieldInfoReqWithDefaults

`func NewGetMappingFieldInfoReqWithDefaults() *GetMappingFieldInfoReq`

NewGetMappingFieldInfoReqWithDefaults instantiates a new GetMappingFieldInfoReq object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadType

`func (o *GetMappingFieldInfoReq) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *GetMappingFieldInfoReq) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *GetMappingFieldInfoReq) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.

### HasWorkloadType

`func (o *GetMappingFieldInfoReq) HasWorkloadType() bool`

HasWorkloadType returns a boolean if a field has been set.

### GetFieldsRef

`func (o *GetMappingFieldInfoReq) GetFieldsRef() []BreakdownFieldRef`

GetFieldsRef returns the FieldsRef field if non-nil, zero value otherwise.

### GetFieldsRefOk

`func (o *GetMappingFieldInfoReq) GetFieldsRefOk() (*[]BreakdownFieldRef, bool)`

GetFieldsRefOk returns a tuple with the FieldsRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldsRef

`func (o *GetMappingFieldInfoReq) SetFieldsRef(v []BreakdownFieldRef)`

SetFieldsRef sets FieldsRef field to given value.

### HasFieldsRef

`func (o *GetMappingFieldInfoReq) HasFieldsRef() bool`

HasFieldsRef returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


