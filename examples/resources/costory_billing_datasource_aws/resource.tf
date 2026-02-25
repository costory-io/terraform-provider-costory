provider "costory" {
  token = var.costory_token
}

resource "costory_billing_datasource_aws" "main" {
  name        = "AWS CUR"
  bucket_name = "my-cur-bucket"
  role_arn    = "arn:aws:iam::123456789012:role/costory-billing-read"
  prefix      = "cur/"
}
