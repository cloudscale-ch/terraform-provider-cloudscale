package cloudscale

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testDSSchema is a minimal schema for dataSourceResourceRead unit tests.
var testDSSchema = map[string]*schema.Schema{
	"name":     {Type: schema.TypeString, Optional: true},
	"ssh_keys": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	"tags":     &TagsSchema,
}

func mockFetch(rows ...ResourceDataRaw) func(*schema.ResourceData, any) ([]ResourceDataRaw, error) {
	return func(_ *schema.ResourceData, _ any) ([]ResourceDataRaw, error) {
		return rows, nil
	}
}

func TestDataSourceRead_SingleMatch(t *testing.T) {
	// two resources with distinct names; the name filter selects exactly one

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{}},
		{"id": "bbb", "name": "beta", "tags": map[string]interface{}{}},
	}
	filter := ResourceDataRaw{"name": "alpha"}
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary)
	}
	if resourceData.Id() != "aaa" {
		t.Errorf("got id=%q, want aaa", resourceData.Id())
	}
}

func TestDataSourceRead_NoMatch(t *testing.T) {
	// name filter matches none of the available resources; returns a zero-match diagnostic

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{}},
	}
	filter := ResourceDataRaw{"name": "gamma"}
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if !diags.HasError() {
		t.Fatal("expected error, got none")
	}
	if diags[0].Summary != "Found zero things" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

func TestDataSourceRead_MultipleMatches(t *testing.T) {
	// no filter attributes set; every resource qualifies and triggers an ambiguity diagnostic

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{}},
		{"id": "bbb", "name": "beta", "tags": map[string]interface{}{}},
	}
	filter := ResourceDataRaw{} // no filter → all match
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if !diags.HasError() {
		t.Fatal("expected error, got none")
	}
	if diags[0].Summary != "Found 2 things, expected one" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

// TestDataSourceRead_TagSubsetMatch is a regression test for
// https://github.com/cloudscale-ch/terraform-provider-cloudscale/pull/143.
//
// Before that fix the filter loop used `m[key] != attr` for all field types.
// Comparing two interface{} values whose dynamic type is map[string]interface{}
// causes a runtime panic ("comparing uncomparable type") in Go.
func TestDataSourceRead_TagSubsetMatch(t *testing.T) {
	// filter holds a strict subset of a resource's tags; the resource still qualifies

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{"env": "prod", "team": "infra"}},
		{"id": "bbb", "name": "beta", "tags": map[string]interface{}{"env": "dev"}},
	}
	filter := ResourceDataRaw{"tags": map[string]interface{}{"env": "prod"}} // strict subset of "aaa"'s tags
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary)
	}
	if resourceData.Id() != "aaa" {
		t.Errorf("got id=%q, want aaa", resourceData.Id())
	}
}

func TestDataSourceRead_TagExactMatch(t *testing.T) {
	// filter tags match a resource's tags exactly; the resource qualifies

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{"env": "prod"}},
		{"id": "bbb", "name": "beta", "tags": map[string]interface{}{"env": "dev"}},
	}
	filter := ResourceDataRaw{"tags": map[string]interface{}{"env": "prod"}}
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary)
	}
	if resourceData.Id() != "aaa" {
		t.Errorf("got id=%q, want aaa", resourceData.Id())
	}
}

func TestDataSourceRead_TagSupersetNoMatch(t *testing.T) {
	// filter demands more tags than the resource carries; nothing qualifies

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{"env": "prod"}},
	}
	filter := ResourceDataRaw{"tags": map[string]interface{}{"env": "prod", "team": "infra"}} // resource only has one of these
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if !diags.HasError() {
		t.Fatal("expected error, got none")
	}
	if diags[0].Summary != "Found zero things" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

func TestDataSourceRead_TagMismatch(t *testing.T) {
	// filter and resource disagree on the value of a shared tag key; nothing qualifies

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{"env": "prod"}},
	}
	filter := ResourceDataRaw{"tags": map[string]interface{}{"env": "dev"}}
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if !diags.HasError() {
		t.Fatal("expected error, got none")
	}
	if diags[0].Summary != "Found zero things" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

func TestDataSourceRead_EmptyTagFilter(t *testing.T) {
	// empty tag map is a no-op filter; all resources qualify and trigger an ambiguity diagnostic

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "tags": map[string]interface{}{"env": "prod"}},
		{"id": "bbb", "name": "beta", "tags": map[string]interface{}{"env": "dev"}},
	}
	filter := ResourceDataRaw{"tags": map[string]interface{}{}} // empty tags = no filter
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if !diags.HasError() {
		t.Fatal("expected ambiguity error, got none")
	}
	if diags[0].Summary != "Found 2 things, expected one" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

func TestDataSourceRead_ListMatch(t *testing.T) {
	// gather returns []string (from the SDK struct); d.GetOk returns []interface{}.
	// The filter must normalise the types before comparing, otherwise reflect.DeepEqual
	// returns false and the resource is silently excluded.

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "ssh_keys": []string{"key-a"}, "tags": map[string]interface{}{}},
		{"id": "bbb", "name": "beta", "ssh_keys": []string{"key-b"}, "tags": map[string]interface{}{}},
	}
	filter := ResourceDataRaw{"ssh_keys": []interface{}{"key-a"}}
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary)
	}
	if resourceData.Id() != "aaa" {
		t.Errorf("got id=%q, want aaa", resourceData.Id())
	}
}

func TestDataSourceRead_ListNoMatch(t *testing.T) {
	// list filter with a value that matches no resource returns a zero-match diagnostic

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "ssh_keys": []string{"key-a"}, "tags": map[string]interface{}{}},
	}
	filter := ResourceDataRaw{"ssh_keys": []interface{}{"key-z"}}
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if !diags.HasError() {
		t.Fatal("expected error, got none")
	}
	if diags[0].Summary != "Found zero things" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

func TestDataSourceRead_ListSubsetNoMatch(t *testing.T) {
	// filter is a strict subset of the resource's list: the resource does NOT match.

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "ssh_keys": []string{"key-a", "key-b"}, "tags": map[string]interface{}{}},
	}
	filter := ResourceDataRaw{"ssh_keys": []interface{}{"key-a"}} // subset, not full list
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert: exact match means ["key-a"] ≠ ["key-a", "key-b"]
	if !diags.HasError() {
		t.Fatal("expected zero-match error (exact match semantics), got none")
	}
	if diags[0].Summary != "Found zero things" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

func TestDataSourceRead_ListOrderMatters(t *testing.T) {
	// same elements, different order — the resource does NOT match.
	// reflect.DeepEqual on slices is order-sensitive.

	// Arrange
	rows := []ResourceDataRaw{
		{"id": "aaa", "name": "alpha", "ssh_keys": []string{"key-a", "key-b"}, "tags": map[string]interface{}{}},
	}
	filter := ResourceDataRaw{"ssh_keys": []interface{}{"key-b", "key-a"}} // reversed
	resourceData := schema.TestResourceDataRaw(t, testDSSchema, filter)

	// Act
	diags := dataSourceResourceRead("things", testDSSchema, mockFetch(rows...))(context.Background(), resourceData, nil)

	// Assert
	if !diags.HasError() {
		t.Fatal("expected zero-match error (order-sensitive), got none")
	}
	if diags[0].Summary != "Found zero things" {
		t.Errorf("unexpected error: %s", diags[0].Summary)
	}
}

