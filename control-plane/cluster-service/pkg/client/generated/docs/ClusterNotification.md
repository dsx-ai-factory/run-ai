# ClusterNotification

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Operation** | [**NotificationOperation**](NotificationOperation.md) |  | 
**Id** | **string** |  | 
**Name** | Pointer to **string** |  | [optional] 
**SchedulingCapabilities** | Pointer to [**SchedulingCapabilitiesNullable**](SchedulingCapabilitiesNullable.md) |  | [optional] 

## Methods

### NewClusterNotification

`func NewClusterNotification(operation NotificationOperation, id string, ) *ClusterNotification`

NewClusterNotification instantiates a new ClusterNotification object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterNotificationWithDefaults

`func NewClusterNotificationWithDefaults() *ClusterNotification`

NewClusterNotificationWithDefaults instantiates a new ClusterNotification object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOperation

`func (o *ClusterNotification) GetOperation() NotificationOperation`

GetOperation returns the Operation field if non-nil, zero value otherwise.

### GetOperationOk

`func (o *ClusterNotification) GetOperationOk() (*NotificationOperation, bool)`

GetOperationOk returns a tuple with the Operation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperation

`func (o *ClusterNotification) SetOperation(v NotificationOperation)`

SetOperation sets Operation field to given value.


### GetId

`func (o *ClusterNotification) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ClusterNotification) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ClusterNotification) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *ClusterNotification) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ClusterNotification) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ClusterNotification) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ClusterNotification) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSchedulingCapabilities

`func (o *ClusterNotification) GetSchedulingCapabilities() SchedulingCapabilitiesNullable`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *ClusterNotification) GetSchedulingCapabilitiesOk() (*SchedulingCapabilitiesNullable, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *ClusterNotification) SetSchedulingCapabilities(v SchedulingCapabilitiesNullable)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.

### HasSchedulingCapabilities

`func (o *ClusterNotification) HasSchedulingCapabilities() bool`

HasSchedulingCapabilities returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


