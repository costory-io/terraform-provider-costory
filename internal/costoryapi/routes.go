package costoryapi

import (
	"net/http"
	"net/url"
)

const (
	routeServiceAccount            = "/terraform/"
	routeBillingDatasourceBase     = "/terraform/billingDatasources"
	routeBillingDatasourceValidate = "/terraform/billingDatasources/validate"
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

var endpointDeleteBillingDatasourceByID = endpointWithRouteParamsContract[billingDatasourceByIDRouteParams, noRequest, noResponse]{
	Method:               http.MethodDelete,
	Path:                 routeBillingDatasourceByIDFromParams,
	ParamsTransport:      requestTransportRouteParams,
	RequestBodyTransport: requestTransportNone,
}

func routeBillingDatasourceByID(id string) string {
	return routeBillingDatasourceBase + "/" + url.PathEscape(id)
}

func routeBillingDatasourceByIDFromParams(params billingDatasourceByIDRouteParams) string {
	return routeBillingDatasourceByID(params.ID)
}
