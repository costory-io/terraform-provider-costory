# COS-2804 — Enrichment filepath between CI and local

Status: Implemented following user review

Created with `yarn spec:new cos-2804-enrichment-filepath` before implementation. Review chose the compatibility path: keep `file`, do not break existing HCL.

**Linear:** [COS-2804](https://linear.app/costory/issue/COS-2804/terraform-issue-with-enrichments-filepath-between-ci-and-local)

Make `costory_enrichment` portable across local and CI by storing YAML content instead of a machine-specific file path. Existing `file = ...` configs keep working.

Related files:

- `internal/provider/enrichment/resource.go` — `file` / `content`, plan, Read, state upgrade
- `examples/resources/costory_enrichment/resource.tf`
- `examples/resources/costory_enrichments_publish/resource.tf`
- `docs/resources/enrichment.md` / `docs/resources/enrichments_publish.md` (generated)
- `templates/resources/enrichment.md.tmpl`

---

## 1. Problem

`costory_enrichment` takes a required `file` string and stores that string in Terraform state. The provider then reads it with `os.ReadFile` on every plan, apply, and refresh:

```go
document, readErr := readDocument(state.File.ValueString())
```

The path is whatever the operator wrote. Docs and examples use:

```hcl
file = "${path.module}/enrichments/${each.value}"
```

`path.module` is not stable across machines:

| How Terraform is invoked | Typical `path.module` |
| --- | --- |
| `terraform apply` from the module directory | `.` |
| `terraform apply -chdir=...`, Terraform Cloud, Atlantis, most CI | absolute checkout path |

Remote state then holds a laptop path such as `/Users/tanguy/infra/enrichments/namespace.yaml`. CI has `/home/runner/work/infra/enrichments/namespace.yaml`.

That shows up as two failures:

1. **Refresh error.** `Read` opens the path from **state**, not from current config. On CI the laptop path does not exist, so plan dies with `Unable to read enrichment file`.
2. **Perpetual diff.** When both paths exist, Terraform still plans an update because the `file` strings differ, even when the YAML is identical.

---

## 2. Goal

Same git checkout, same YAML, no path in state: `terraform plan` is empty on CI after a local apply, and the other way around. Refresh never opens a path from another machine.

Existing setups that pass `file` keep applying without an HCL change on Terraform 1.11+. On Terraform 1.10 or earlier, switch to `content = file(...)` before upgrading the provider.

---

## 3. Change (compatibility path)

Keep `file`. Add `content`. Store only the YAML in state. `file` is write-only (Terraform ≥ 1.11): present in config, never in state or plan diffs. Write-only attributes require Terraform 1.11; older Terraform must use `content = file(...)`.

```hcl
# Existing configs — unchanged, still valid
resource "costory_enrichment" "all" {
  for_each = fileset("${path.module}/enrichments", "*.yaml")
  name     = trimsuffix(each.value, ".yaml")
  file     = "${path.module}/enrichments/${each.value}"
}

# Preferred for new configs
resource "costory_enrichment" "all" {
  for_each = fileset("${path.module}/enrichments", "*.yaml")
  name     = trimsuffix(each.value, ".yaml")
  content  = file("${path.module}/enrichments/${each.value}")
}
```

Exactly one of `file` or `content` is set.

### Schema

| Attribute | Required | In state | Role |
| --- | --- | --- | --- |
| `name` | yes | yes | Unchanged. |
| `file` | optional, write-only | no | Path. Plan reads it into `content`. Existing HCL keeps working. |
| `content` | optional, computed | yes | YAML body. Set in config, or filled from `file`. |
| `content_sha256` | computed | yes | Hash of the YAML bytes, empty when live API document ≠ parsed content. |
| `id` | computed | yes | Unchanged. |

Do not mark `content` sensitive. Do not warn/deprecate `file` (noise for existing setups).

### Provider behavior

- Validate: exactly one of `file` or `content`.
- ModifyPlan: if `content` is empty and `file` is set, `os.ReadFile` into `content`; hash into `content_sha256`.
- Create / Update: `parseDocument` from planned `content`. No path in apply.
- Read: parse `state.Content`, compare to `GetEnrichment`. Never `os.ReadFile`. If content is empty (upgrade on a machine that lacked the old path), leave `content_sha256` as stored.
- `file` is on the Go model so Plan/State Get match the schema, but it is always set null. The path is read from Config.

### State upgrade (schema 0 → 1)

- If the old `file` path exists on this machine, read it into `content` and drop `file`.
- If it does not, set `content` to `""` and keep `content_sha256`. The next plan fills `content` from config `file` or `content`. One apply may PUT the same document to persist content in state; publish is not retriggered if the hash is unchanged.

### Docs

- Examples prefer `content = file(...)`.
- Schema text says `file` is still accepted and is not stored.
- mintlify-docs pages are a follow-up in that repo.

---

## 4. Alternatives not taken

**Replace `file` with `content` only.** Breaking HCL. Rejected in review.

**Keep `file` in state and suppress path diffs.** Refresh still uses the path in state. Rejected.

---

## 5. Out of scope

- Enrichment API, YAML schema, or `costory_enrichments_publish` behavior (examples only).
- Changing how `content_sha256` is used as the publish trigger.
- Normalizing YAML before hashing (byte-for-byte, as today).
- mintlify-docs (separate repo).

---

## 6. Tests

- Hash a content string (no temp file).
- `resolveYAML`: content wins; file is read when content is empty; missing file errors.
- State upgrade: existing path → `content` populated; missing path → empty `content`, no error.
- `go test ./...` and regenerate docs.

---

## 7. Checklist

- [x] Spec reviewed — compatibility path (`file` write-only + `content`)
- [x] Provider schema, upgrade, examples, registry docs
- [x] Tests listed above
- [ ] mintlify-docs Terraform / enrichments pages (follow-up)

History:

| Date | Status | Notes |
| --- | --- | --- |
| 2026-11-25 | Spec drafted | COS-2804. `yarn spec:new` first. |
| 2026-11-25 | Compat path chosen | Keep `file`; store YAML in `content`. Do not break existing HCL. |
| 2026-11-25 | Implemented | Write-only `file`, stored `content`, state upgrade 0→1. |
| 2026-11-25 | Review fixes | `File` on the model (always null); TF 1.11 noted for remaining `file` users. |
