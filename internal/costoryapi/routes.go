package costoryapi

import (
	"net/http"
	"net/url"
)

const (
	routeServiceAccount            = "/terraform/"
	routeBillingDatasourceBase     = "/terraform/billingDatasources"
	routeBillingDatasourceValidate = "/terraform/billingDatasources/validate"
	routeMetricsDatasourceBase     = "/terraform/metricsDatasources"
	routeMetricsDatasourceValidate = "/terraform/metricsDatasources/validate"
)

type requestTransport string

const (
	requestTransportNone        requestTransport = "none"
	requestTransportJSONBody    requestTransport = "json_body"
	requestTransportRouteParams requestTransport = "route_params"
)

type noRequest struct{}
type noResponse struct{}

type billingDatasourceByIDRouteParams struct {
	ID string
}

type metricsDatasourceByIDRouteParams struct {
	ID string
}

type endpointContract[Req any, Resp any] struct {
	Method           string
	Path             string
	RequestTransport requestTransport
}

type endpointWithRouteParamsContract[Params any, Req any, Resp any] struct {
	Method               string
	Path                 func(params Params) string
	ParamsTransport      requestTransport
	RequestBodyTransport requestTransport
}

var endpointGetServiceAccount = endpointContract[noRequest, ServiceAccountResponse]{
	Method:           http.MethodGet,
	Path:             routeServiceAccount,
	RequestTransport: requestTransportNone,
}

var endpointValidateGCPBillingDatasource = endpointContract[gcpBillingDatasourceAPIRequest, noResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceValidate,
	RequestTransport: requestTransportJSONBody,
}

var endpointValidateAWSBillingDatasource = endpointContract[awsBillingDatasourceAPIRequest, noResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceValidate,
	RequestTransport: requestTransportJSONBody,
}

var endpointValidateCursorBillingDatasource = endpointContract[externalBillingDatasourceAPIRequest, noResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceValidate,
	RequestTransport: requestTransportJSONBody,
}

var endpointValidateAnthropicBillingDatasource = endpointContract[externalBillingDatasourceAPIRequest, noResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceValidate,
	RequestTransport: requestTransportJSONBody,
}

var endpointValidateAzureBillingDatasource = endpointContract[azureBillingDatasourceAPIRequest, noResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceValidate,
	RequestTransport: requestTransportJSONBody,
}

var endpointCreateGCPBillingDatasource = endpointContract[gcpBillingDatasourceAPIRequest, gcpBillingDatasourceAPIResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceBase,
	RequestTransport: requestTransportJSONBody,
}

var endpointCreateAWSBillingDatasource = endpointContract[awsBillingDatasourceAPIRequest, awsBillingDatasourceAPIResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceBase,
	RequestTransport: requestTransportJSONBody,
}

var endpointCreateCursorBillingDatasource = endpointContract[externalBillingDatasourceAPIRequest, externalBillingDatasourceAPIResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceBase,
	RequestTransport: requestTransportJSONBody,
}

var endpointCreateAnthropicBillingDatasource = endpointContract[externalBillingDatasourceAPIRequest, externalBillingDatasourceAPIResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceBase,
	RequestTransport: requestTransportJSONBody,
}

var endpointCreateAzureBillingDatasource = endpointContract[azureBillingDatasourceAPIRequest, azureBillingDatasourceAPIResponse]{
	Method:           http.MethodPost,
	Path:             routeBillingDatasourceBase,
	RequestTransport: requestTransportJSONBody,
}

var endpointGetGCPBillingDatasourceByID = endpointWithRouteParamsContract[billingDatasourceByIDRouteParams, noRequest, gcpBillingDatasourceAPIResponse]{
	Method:               http.MethodGet,
	Path:                 routeBillingDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

var endpointGetAWSBillingDatasourceByID = endpointWithRouteParamsContract[billingDatasourceByIDRouteParams, noRequest, awsBillingDatasourceAPIResponse]{
	Method:               http.MethodGet,
	Path:                 routeBillingDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

var endpointGetCursorBillingDatasourceByID = endpointWithRouteParamsContract[billingDatasourceByIDRouteParams, noRequest, externalBillingDatasourceAPIResponse]{
	Method:               http.MethodGet,
	Path:                 routeBillingDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

var endpointGetAnthropicBillingDatasourceByID = endpointWithRouteParamsContract[billingDatasourceByIDRouteParams, noRequest, externalBillingDatasourceAPIResponse]{
	Method:               http.MethodGet,
	Path:                 routeBillingDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

var endpointGetAzureBillingDatasourceByID = endpointWithRouteParamsContract[billingDatasourceByIDRouteParams, noRequest, azureBillingDatasourceAPIResponse]{
	Method:               http.MethodGet,
	Path:                 routeBillingDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

var endpointDeleteBillingDatasourceByID = endpointWithRouteParamsContract[billingDatasourceByIDRouteParams, noRequest, noResponse]{
	Method:               http.MethodDelete,
	Path:                 routeBillingDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

var endpointValidateMetricsDatasource = endpointContract[metricsDatasourceAPIRequest, metricsDatasourceValidateAPIResponse]{
	Method:           http.MethodPost,
	Path:             routeMetricsDatasourceValidate,
	RequestTransport: requestTransportJSONBody,
}

var endpointCreateMetricsDatasource = endpointContract[metricsDatasourceAPIRequest, metricsDatasourceAPIResponse]{
	Method:           http.MethodPost,
	Path:             routeMetricsDatasourceBase,
	RequestTransport: requestTransportJSONBody,
}

var endpointGetMetricsDatasourceByID = endpointWithRouteParamsContract[metricsDatasourceByIDRouteParams, noRequest, metricsDatasourceAPIResponse]{
	Method:               http.MethodGet,
	Path:                 routeMetricsDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

var endpointPatchMetricsDatasourceByID = endpointWithRouteParamsContract[metricsDatasourceByIDRouteParams, metricsDatasourcePatchAPIRequest, noResponse]{
	Method:               http.MethodPatch,
	Path:                 routeMetricsDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportJSONBody,
}

var endpointDeleteMetricsDatasourceByID = endpointWithRouteParamsContract[metricsDatasourceByIDRouteParams, noRequest, noResponse]{
	Method:               http.MethodDelete,
	Path:                 routeMetricsDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

func routeBillingDatasourceByID(id string) string {
	return routeBillingDatasourceBase + "/" + url.PathEscape(id)
}

func routeBillingDatasourceByIDFromParams(params billingDatasourceByIDRouteParams) string {
	return routeBillingDatasourceByID(params.ID)
}

func routeMetricsDatasourceByID(id string) string {
	return routeMetricsDatasourceBase + "/" + url.PathEscape(id)
}

func routeMetricsDatasourceByIDFromParams(params metricsDatasourceByIDRouteParams) string {
	return routeMetricsDatasourceByID(params.ID)
}
