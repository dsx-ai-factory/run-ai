# ClusterReportedResourceRequirements

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Limits** | Pointer to **map[string]string** | Resource limits keyed by Kubernetes resource name. | [optional] 
**Requests** | Pointer to **map[string]string** | Resource requests keyed by Kubernetes resource name. | [optional] 

## Methods

### NewClusterReportedResourceRequirements

`func NewClusterReportedResourceRequirements() *ClusterReportedResourceRequirements`

NewClusterReportedResourceRequirements instantiates a new ClusterReportedResourceRequirements object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedResourceRequirementsWithDefaults

`func NewClusterReportedResourceRequirementsWithDefaults() *ClusterReportedResourceRequirements`

NewClusterReportedResourceRequirementsWithDefaults instantiates a new ClusterReportedResourceRequirements object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLimits

`func (o *ClusterReportedResourceRequirements) GetLimits() map[string]string`

GetLimits returns the Limits field if non-nil, zero value otherwise.

### GetLimitsOk

`func (o *ClusterReportedResourceRequirements) GetLimitsOk() (*map[string]string, bool)`

GetLimitsOk returns a tuple with the Limits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimits

`func (o *ClusterReportedResourceRequirements) SetLimits(v map[string]string)`

SetLimits sets Limits field to given value.

### HasLimits

`func (o *ClusterReportedResourceRequirements) HasLimits() bool`

HasLimits returns a boolean if a field has been set.

### GetRequests

`func (o *ClusterReportedResourceRequirements) GetRequests() map[string]string`

GetRequests returns the Requests field if non-nil, zero value otherwise.

### GetRequestsOk

`func (o *ClusterReportedResourceRequirements) GetRequestsOk() (*map[string]string, bool)`

GetRequestsOk returns a tuple with the Requests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequests

`func (o *ClusterReportedResourceRequirements) SetRequests(v map[string]string)`

SetRequests sets Requests field to given value.

### HasRequests

`func (o *ClusterReportedResourceRequirements) HasRequests() bool`

HasRequests returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


