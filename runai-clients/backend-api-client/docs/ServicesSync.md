# ServicesSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Namespace** | **string** | The namespace we are syncing objects in | 
**Services** | [**[]Service**](Service.md) |  | 

## Methods

### NewServicesSync

`func NewServicesSync(namespace string, services []Service, ) *ServicesSync`

NewServicesSync instantiates a new ServicesSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServicesSyncWithDefaults

`func NewServicesSyncWithDefaults() *ServicesSync`

NewServicesSyncWithDefaults instantiates a new ServicesSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNamespace

`func (o *ServicesSync) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ServicesSync) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ServicesSync) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetServices

`func (o *ServicesSync) GetServices() []Service`

GetServices returns the Services field if non-nil, zero value otherwise.

### GetServicesOk

`func (o *ServicesSync) GetServicesOk() (*[]Service, bool)`

GetServicesOk returns a tuple with the Services field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServices

`func (o *ServicesSync) SetServices(v []Service)`

SetServices sets Services field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


