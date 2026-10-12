package enrichment

import "testing"

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
