# COS-2804 — Enrichment filepath between CI and local

Status: Spec In Progress

Created with `yarn spec:new cos-2804-enrichment-filepath` before implementation. Waiting on review before any provider code change.

**Linear:** [COS-2804](https://linear.app/costory/issue/COS-2804/terraform-issue-with-enrichments-filepath-between-ci-and-local)

Make `costory_enrichment` portable across local and CI by storing YAML content instead of a machine-specific file path.

Related files:

- `internal/provider/enrichment/resource.go` — `file` attribute, `os.ReadFile` in Create/Update/Read and `fileHashModifier`
- `examples/resources/costory_enrichment/resource.tf`
- `examples/resources/costory_enrichments_publish/resource.tf`
- `docs/resources/enrichment.md` / `docs/resources/enrichments_publish.md` (generated)
- `templates/resources/enrichment.md.tmpl`
- mintlify-docs `setup/terraform.mdx` and `features/tagging/enrichments.mdx` (follow-up in that repo)

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

The provider does not call `filepath.Abs`; it does not need to. Asking for a filesystem path and persisting it is enough to pin the resource to one machine.

---

## 2. Goal

Same git checkout, same YAML, no path in state: `terraform plan` is empty on CI after a local apply, and the other way around. Refresh never opens a path from another machine.

Operators keep writing YAML files in git. Terraform's built-in `file()` reads them on the machine that is planning.

---

## 3. Proposed change

Replace the `file` attribute with `content`. Documented usage becomes:

```hcl
resource "costory_enrichment" "all" {
  for_each = fileset("${path.module}/enrichments", "*.yaml")

  name    = trimsuffix(each.value, ".yaml")
  content = file("${path.module}/enrichments/${each.value}")
}
```

`file()` is evaluated by Terraform, not by the provider. The YAML bytes in config and state are the same on every machine.

### Schema

| Attribute | Required | In state | Role |
| --- | --- | --- | --- |
| `name` | yes | yes | Unchanged. Still must match `name` in the YAML. |
| `content` | yes | yes | Raw YAML body. |
| `file` | removed | upgraded away | — |
| `content_sha256` | computed | yes | Unchanged. Hash of the YAML bytes, empty when live API document ≠ parsed content. Still passed to `costory_enrichments_publish`. |
| `id` | computed | yes | Unchanged. |

Do not mark `content` sensitive: enrichment YAML is org mapping, and plans should show the diff.

### Provider behavior

- Create / Update: `parseDocument` from `content` (already used by tests). No `os.ReadFile`.
- Plan modifier: hash `content` into `content_sha256` (today it hashes the file on disk).
- Read: parse `state.Content`, compare to `GetEnrichment`, set `content_sha256` to the hash or `""`. Never touch the filesystem.
- Name mismatch error quotes the enrichment name, not a path.

### State upgrade (schema 0 → 1)

Follow the Azure resource pattern (`UpgradeState` in `azure_resource.go`).

- If the old `file` path exists on this machine, read it into `content` and drop `file`.
- If it does not (typical CI after a local apply), set `content` to `""` and keep `content_sha256`. The next plan fills `content` from `file(...)` in config. If the YAML is unchanged, `content_sha256` stays the same, so publish is not retriggered unless Update still PUTs. **Prefer:** when upgrade content is empty, Read skips the local/API compare and leaves `content_sha256` as stored, so a content-only backfill does not flip the hash. Update may still PUT the same document once to persist `content` in state; that is acceptable.

Existing configs must change `file = path` to `content = file(path)` before or in the same apply as the provider upgrade. Terraform will reject unknown `file` once it is removed from the schema.

### Docs

- Provider examples and generated registry docs.
- mintlify-docs `setup/terraform.mdx` and `features/tagging/enrichments.mdx` (separate PR in that repo, same wording).

---

## 4. Alternatives considered

**A. Keep `file`, suppress path diffs when hashes match.** Refresh still uses the path in state, so CI still errors. Rejected.

**B. Keep `file` and rewrite it to a path relative to cwd.** `terraform apply -chdir` and Terraform Cloud make cwd/module paths diverge. HashiCorp's guidance is not to persist cwd-relative paths. Rejected.

**C. `content` plus optional write-only `file` (`ExactlyOneOf`).** Old HCL keeps working; `file` never lands in state; a plan modifier reads the path at plan time into `content`. Needs Terraform ≥ 1.11 (`WriteOnly` on plugin-framework 1.16+, we are on 1.17.0). More code, two ways to say the same thing. Use only if we must avoid a breaking config change.

**D. Store only the hash, not the YAML.** Refresh cannot compare the local document to the API without the body. Rejected.

Recommended: **replace `file` with `content`** (section 3). The resource shipped in 0.3.1 (PR #17); the HCL change is one function call. Alternative C is the fallback if review wants a deprecation window.

---

## 5. Out of scope

- Enrichment API, YAML schema, or `costory_enrichments_publish` behavior (examples only).
- Changing how `content_sha256` is used as the publish trigger.
- Normalizing YAML before hashing (byte-for-byte, as today).

---

## 6. Tests

- `parseDocument` / `documentsEqual` stay; add hashing of a content string (no temp file).
- Read never calls `os.ReadFile`.
- State upgrade: existing path → `content` populated; missing path → empty `content`, no error.
- Example HCL uses `content = file(...)`.
- `go test ./...` and regenerate docs.

---

## 7. Checklist

- [ ] Spec reviewed — **do not implement until this is confirmed**
- [ ] Confirm breaking `file` → `content` vs write-only `file` compatibility (alternative C)
- [ ] Provider schema, upgrade, examples, registry docs
- [ ] mintlify-docs Terraform / enrichments pages
- [ ] Tests listed above

History:

| Date | Status | Notes |
| --- | --- | --- |
| 2026-11-25 | Spec drafted | COS-2804. `yarn spec:new` first; waiting for review. |
