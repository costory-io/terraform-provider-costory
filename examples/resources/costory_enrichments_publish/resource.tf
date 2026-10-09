resource "costory_enrichment" "all" {
  for_each = fileset("${path.module}/enrichments", "*.yaml")

  name = trimsuffix(each.value, ".yaml")
  file = "${path.module}/enrichments/${each.value}"
}

# Publishes every draft in one transaction after the enrichments are written,
# and removes the enrichments whose file was deleted.
resource "costory_enrichments_publish" "this" {
  enrichments = { for enrichment in costory_enrichment.all : enrichment.name => enrichment.content_sha256 }
}
