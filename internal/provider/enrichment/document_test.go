package enrichment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDocumentsEqualIgnoresKeyOrderAndNull(t *testing.T) {
	t.Parallel()

	file := map[string]any{
		"name":       "namespace",
		"keyField":   "cos_namespace",
		"priority":   float64(2),
		"conditions": nil,
		"targetColumns": []any{
			map[string]any{"name": "owner_org", "mode": "merge"},
		},
		"entries": []any{},
	}
	stored := map[string]any{
		"entries": []any{},
		"targetColumns": []any{
			map[string]any{"mode": "merge", "name": "owner_org"},
		},
		"priority": float64(2),
		"keyField": "cos_namespace",
		"name":     "namespace",
	}

	if !documentsEqual(file, stored) {
		t.Fatal("expected the file and the stored document to match")
	}
}

func TestParseDocumentTurnsTargetColumnsIntoAnOrderedList(t *testing.T) {
	t.Parallel()

	document, err := parseDocument("01.yaml", []byte(`
name: 01_pod_label_to_org
keyField: cos_namespace
priority: 1
targetColumns:
  owner_org_code: { mode: replace }
  owner_org: { mode: merge }
entries:
  - key: fluent-bit
    then: { owner_org_code: "3282", owner_org: SRE }
`))
	if err != nil {
		t.Fatal(err)
	}
	columns, ok := document["targetColumns"].([]any)
	if !ok || len(columns) != 2 {
		t.Fatalf("unexpected targetColumns: %#v", document["targetColumns"])
	}
	first := columns[0].(map[string]any)
	second := columns[1].(map[string]any)
	if first["name"] != "owner_org_code" || first["mode"] != "replace" || second["name"] != "owner_org" || second["mode"] != "merge" {
		t.Fatalf("expected file order with modes, got %#v", columns)
	}
}

func TestParseDocumentKeepsAListAsIs(t *testing.T) {
	t.Parallel()

	document, err := parseDocument("a.yaml", []byte(`
name: a
targetColumns:
  - { name: owner_org, mode: merge }
`))
	if err != nil {
		t.Fatal(err)
	}
	columns := document["targetColumns"].([]any)
	if columns[0].(map[string]any)["name"] != "owner_org" {
		t.Fatalf("unexpected targetColumns: %#v", columns)
	}
}

func TestDocumentsEqualIgnoresAnEmptyStoredDescription(t *testing.T) {
	t.Parallel()

	file := map[string]any{"name": "a", "priority": float64(0)}
	stored := map[string]any{"name": "a", "priority": float64(0), "description": ""}
	if !documentsEqual(file, stored) {
		t.Fatal("an empty stored description should match a file without one")
	}
	stored["description"] = "Owner"
	if documentsEqual(file, stored) {
		t.Fatal("a description should count")
	}
}

func TestRemovedNames(t *testing.T) {
	t.Parallel()

	removed := removedNames(
		map[string]string{"a": "1", "b": "2", "c": "3"},
		map[string]string{"b": "2", "d": "4"},
	)
	if len(removed) != 2 || removed[0] != "a" || removed[1] != "c" {
		t.Fatalf("unexpected removed names: %v", removed)
	}
}

func TestDocsExampleParses(t *testing.T) {
	t.Parallel()

	document, err := readDocument("../../../examples/resources/costory_enrichment/enrichments/namespace.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if document["name"] != "namespace" || len(document["targetColumns"].([]any)) != 1 {
		t.Fatalf("unexpected document: %#v", document)
	}
}

func TestHashBytes(t *testing.T) {
	t.Parallel()

	got := hashBytes([]byte("abc"))
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("hashBytes(abc) = %q, want %q", got, want)
	}
	if hashBytes([]byte("abd")) == got {
		t.Fatal("different content must not hash the same")
	}
}

func TestResolveYAMLPrefersContent(t *testing.T) {
	t.Parallel()

	raw, ok, err := resolveYAML(types.StringValue("name: from-content\n"), types.StringValue("/no/such/file.yaml"))
	if err != nil || !ok {
		t.Fatalf("unexpected: ok=%v err=%v", ok, err)
	}
	if string(raw) != "name: from-content\n" {
		t.Fatalf("unexpected content: %q", raw)
	}
}

func TestResolveYAMLReadsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "namespace.yaml")
	if err := os.WriteFile(path, []byte("name: namespace\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, ok, err := resolveYAML(types.StringNull(), types.StringValue(path))
	if err != nil || !ok {
		t.Fatalf("unexpected: ok=%v err=%v", ok, err)
	}
	if string(raw) != "name: namespace\n" {
		t.Fatalf("unexpected content: %q", raw)
	}
}

func TestResolveYAMLUnknownWaits(t *testing.T) {
	t.Parallel()

	raw, ok, err := resolveYAML(types.StringUnknown(), types.StringValue("x.yaml"))
	if err != nil || ok || raw != nil {
		t.Fatalf("unknown content must wait, got ok=%v err=%v raw=%q", ok, err, raw)
	}
}

func TestResolveYAMLMissingFile(t *testing.T) {
	t.Parallel()

	_, ok, err := resolveYAML(types.StringNull(), types.StringValue("/no/such/enrichment.yaml"))
	if err == nil || ok {
		t.Fatal("expected an error for a missing file")
	}
}

func TestUpgradeEnrichmentStateV0toV1ReadsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "namespace.yaml")
	if err := os.WriteFile(path, []byte("name: namespace\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := upgradeEnrichmentStateV0toV1(enrichmentResourceModelV0{
		ID:            types.StringValue("namespace"),
		Name:          types.StringValue("namespace"),
		File:          types.StringValue(path),
		ContentSHA256: types.StringValue("abc"),
	})
	if got.Content.ValueString() != "name: namespace\n" {
		t.Fatalf("expected file bytes in content, got %q", got.Content.ValueString())
	}
	if !got.File.IsNull() {
		t.Fatalf("file must stay null in upgraded state, got %q", got.File.ValueString())
	}
	if got.ContentSHA256.ValueString() != "abc" || got.Name.ValueString() != "namespace" {
		t.Fatalf("expected identity fields kept, got name=%q sha=%q", got.Name.ValueString(), got.ContentSHA256.ValueString())
	}
}

func TestUpgradeEnrichmentStateV0toV1MissingFile(t *testing.T) {
	t.Parallel()

	got := upgradeEnrichmentStateV0toV1(enrichmentResourceModelV0{
		ID:            types.StringValue("namespace"),
		Name:          types.StringValue("namespace"),
		File:          types.StringValue("/no/such/enrichment.yaml"),
		ContentSHA256: types.StringValue("abc"),
	})
	if got.Content.ValueString() != "" {
		t.Fatalf("missing path should leave content empty, got %q", got.Content.ValueString())
	}
	if !got.File.IsNull() {
		t.Fatalf("file must stay null in upgraded state, got %q", got.File.ValueString())
	}
	if got.ContentSHA256.ValueString() != "abc" {
		t.Fatalf("expected stored hash kept, got %q", got.ContentSHA256.ValueString())
	}
}
