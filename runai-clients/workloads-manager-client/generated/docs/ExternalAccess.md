# ExternalAccess

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ingress** | Pointer to **map[string]interface{}** | Full Kubernetes Ingress resource (networking.k8s.io/v1). Used when cluster uses Ingress. | [optional] 
**Route** | Pointer to **map[string]interface{}** | Full OpenShift Route resource (route.openshift.io/v1). Used when cluster uses Route. | [optional] 
**HttpRoute** | Pointer to **map[string]interface{}** | Full Gateway API HTTPRoute resource (gateway.networking.k8s.io/v1). Used when cluster uses Gateway API. | [optional] 

## Methods

### NewExternalAccess

`func NewExternalAccess() *ExternalAccess`

NewExternalAccess instantiates a new ExternalAccess object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExternalAccessWithDefaults

`func NewExternalAccessWithDefaults() *ExternalAccess`

NewExternalAccessWithDefaults instantiates a new ExternalAccess object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIngress

`func (o *ExternalAccess) GetIngress() map[string]interface{}`

GetIngress returns the Ingress field if non-nil, zero value otherwise.

### GetIngressOk

`func (o *ExternalAccess) GetIngressOk() (*map[string]interface{}, bool)`

GetIngressOk returns a tuple with the Ingress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIngress

`func (o *ExternalAccess) SetIngress(v map[string]interface{})`

SetIngress sets Ingress field to given value.

### HasIngress

`func (o *ExternalAccess) HasIngress() bool`

HasIngress returns a boolean if a field has been set.

### SetIngressNil

`func (o *ExternalAccess) SetIngressNil(b bool)`

 SetIngressNil sets the value for Ingress to be an explicit nil

### UnsetIngress
`func (o *ExternalAccess) UnsetIngress()`

UnsetIngress ensures that no value is present for Ingress, not even an explicit nil
### GetRoute

`func (o *ExternalAccess) GetRoute() map[string]interface{}`

GetRoute returns the Route field if non-nil, zero value otherwise.

### GetRouteOk

`func (o *ExternalAccess) GetRouteOk() (*map[string]interface{}, bool)`

GetRouteOk returns a tuple with the Route field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoute

`func (o *ExternalAccess) SetRoute(v map[string]interface{})`

SetRoute sets Route field to given value.

### HasRoute

`func (o *ExternalAccess) HasRoute() bool`

HasRoute returns a boolean if a field has been set.

### SetRouteNil

`func (o *ExternalAccess) SetRouteNil(b bool)`

 SetRouteNil sets the value for Route to be an explicit nil

### UnsetRoute
`func (o *ExternalAccess) UnsetRoute()`

UnsetRoute ensures that no value is present for Route, not even an explicit nil
### GetHttpRoute

`func (o *ExternalAccess) GetHttpRoute() map[string]interface{}`

GetHttpRoute returns the HttpRoute field if non-nil, zero value otherwise.

### GetHttpRouteOk

`func (o *ExternalAccess) GetHttpRouteOk() (*map[string]interface{}, bool)`

GetHttpRouteOk returns a tuple with the HttpRoute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHttpRoute

`func (o *ExternalAccess) SetHttpRoute(v map[string]interface{})`

SetHttpRoute sets HttpRoute field to given value.

### HasHttpRoute

`func (o *ExternalAccess) HasHttpRoute() bool`

HasHttpRoute returns a boolean if a field has been set.

### SetHttpRouteNil

`func (o *ExternalAccess) SetHttpRouteNil(b bool)`

 SetHttpRouteNil sets the value for HttpRoute to be an explicit nil

### UnsetHttpRoute
`func (o *ExternalAccess) UnsetHttpRoute()`

UnsetHttpRoute ensures that no value is present for HttpRoute, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


