# GetMappingMinimalInfoReq

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadType** | [**GVK**](GVK.md) | The Kubernetes group, version, and kind of the workload type whose minimal mapping information is requested. | 

## Methods

### NewGetMappingMinimalInfoReq

`func NewGetMappingMinimalInfoReq(workloadType GVK, ) *GetMappingMinimalInfoReq`

NewGetMappingMinimalInfoReq instantiates a new GetMappingMinimalInfoReq object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetMappingMinimalInfoReqWithDefaults

`func NewGetMappingMinimalInfoReqWithDefaults() *GetMappingMinimalInfoReq`

NewGetMappingMinimalInfoReqWithDefaults instantiates a new GetMappingMinimalInfoReq object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadType

`func (o *GetMappingMinimalInfoReq) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *GetMappingMinimalInfoReq) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *GetMappingMinimalInfoReq) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


