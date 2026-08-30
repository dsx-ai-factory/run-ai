# ConnectivityFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ports** | Pointer to [**[]Port**](Port.md) | Set of container ports that the workload exposes. | [optional] 
**ExposedUrls** | Pointer to [**[]ExposedUrl**](ExposedUrl.md) | Set of container ports that the workload exposes via URLs. | [optional] 
**RelatedUrls** | Pointer to [**[]RelatedUrl**](RelatedUrl.md) | Set of URLs that are related to the workload. | [optional] 

## Methods

### NewConnectivityFields

`func NewConnectivityFields() *ConnectivityFields`

NewConnectivityFields instantiates a new ConnectivityFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectivityFieldsWithDefaults

`func NewConnectivityFieldsWithDefaults() *ConnectivityFields`

NewConnectivityFieldsWithDefaults instantiates a new ConnectivityFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPorts

`func (o *ConnectivityFields) GetPorts() []Port`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *ConnectivityFields) GetPortsOk() (*[]Port, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *ConnectivityFields) SetPorts(v []Port)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *ConnectivityFields) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *ConnectivityFields) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *ConnectivityFields) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *ConnectivityFields) GetExposedUrls() []ExposedUrl`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *ConnectivityFields) GetExposedUrlsOk() (*[]ExposedUrl, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *ConnectivityFields) SetExposedUrls(v []ExposedUrl)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *ConnectivityFields) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *ConnectivityFields) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *ConnectivityFields) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *ConnectivityFields) GetRelatedUrls() []RelatedUrl`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *ConnectivityFields) GetRelatedUrlsOk() (*[]RelatedUrl, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *ConnectivityFields) SetRelatedUrls(v []RelatedUrl)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *ConnectivityFields) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *ConnectivityFields) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *ConnectivityFields) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


