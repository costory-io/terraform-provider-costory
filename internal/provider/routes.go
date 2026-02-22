package provider

import "net/url"

const (
	routeServiceAccount            = "/api/v1/terraform/context"
	routeBillingDatasourceBase     = "/billing-datasources/terraform"
	routeBillingDatasourceValidate = "/billing-datasources/terraform/validate"
)

func routeBillingDatasourceByID(id string) string {
	return routeBillingDatasourceBase + "/" + url.PathEscape(id)
}
