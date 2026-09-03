# WorkloadV2UpdateConfigurationPvcResourcesRequests

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Storage** | **string** | Requested storage size as a Kubernetes quantity string (e.g. &#x60;1Gi&#x60;, &#x60;500Mi&#x60;, &#x60;10Ti&#x60;).  | 

## Methods

### NewWorkloadV2UpdateConfigurationPvcResourcesRequests

`func NewWorkloadV2UpdateConfigurationPvcResourcesRequests(storage string, ) *WorkloadV2UpdateConfigurationPvcResourcesRequests`

NewWorkloadV2UpdateConfigurationPvcResourcesRequests instantiates a new WorkloadV2UpdateConfigurationPvcResourcesRequests object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2UpdateConfigurationPvcResourcesRequestsWithDefaults

`func NewWorkloadV2UpdateConfigurationPvcResourcesRequestsWithDefaults() *WorkloadV2UpdateConfigurationPvcResourcesRequests`

NewWorkloadV2UpdateConfigurationPvcResourcesRequestsWithDefaults instantiates a new WorkloadV2UpdateConfigurationPvcResourcesRequests object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStorage

`func (o *WorkloadV2UpdateConfigurationPvcResourcesRequests) GetStorage() string`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkloadV2UpdateConfigurationPvcResourcesRequests) GetStorageOk() (*string, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkloadV2UpdateConfigurationPvcResourcesRequests) SetStorage(v string)`

SetStorage sets Storage field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


