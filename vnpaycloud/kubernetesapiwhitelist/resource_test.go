package kubernetesapiwhitelist

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateWhitelistCIDR(t *testing.T) {
	cases := map[string]bool{
		"203.0.113.0/24":  true,
		"198.51.100.7/32": true,
		"203.0.113.5/24":  false,
		"0.0.0.0/0":       false,
		"203.0.113.5":     false,
		"2001:db8::/32":   false,
		"not-a-cidr":      false,
	}
	for in, valid := range cases {
		_, errs := validateWhitelistCIDR(in, "cidr")
		if valid && len(errs) > 0 {
			t.Errorf("%s should be valid: %v", in, errs)
		}
		if !valid && len(errs) == 0 {
			t.Errorf("%s should be rejected", in)
		}
	}
}

func TestResourceApiWhitelistSchema(t *testing.T) {
	res := ResourceKubernetesApiWhitelist()
	if !res.Schema["cluster_id"].Required || !res.Schema["cluster_id"].ForceNew {
		t.Error("expected cluster_id to be Required+ForceNew")
	}
	if e := res.Schema["entry"]; e.Type != schema.TypeSet || !e.Required || e.MinItems != 1 || e.MaxItems != 50 {
		t.Error("expected entry to be a required set of 1 to 50 blocks")
	}
	for _, key := range []string{"endpoint", "private_access", "max_entries", "supported", "unsupported_reason"} {
		if !res.Schema[key].Computed {
			t.Errorf("expected %s to be Computed", key)
		}
	}
}

func TestResourceApiWhitelistRead(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ClusterApiWhitelist(testhelpers.TestProjectID, "clu-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ClusterApiWhitelistResponse{
				Entries:    []dto.ApiWhitelistEntry{{Cidr: "203.0.113.0/24", Note: "office"}},
				Endpoint:   "203.0.113.200",
				MaxEntries: 50,
				Supported:  true,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesApiWhitelist()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("clu-001")

	if diags := res.ReadContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Get("cluster_id").(string) != "clu-001" {
		t.Errorf("expected cluster_id clu-001, got %s", d.Get("cluster_id"))
	}
	if d.Get("endpoint").(string) != "203.0.113.200" || d.Get("max_entries").(int) != 50 || !d.Get("supported").(bool) {
		t.Errorf("computed attributes not set: %v", d.State().Attributes)
	}
	entries := d.Get("entry").(*schema.Set).List()
	if len(entries) != 1 || entries[0].(map[string]interface{})["cidr"] != "203.0.113.0/24" {
		t.Errorf("unexpected entries: %v", entries)
	}
}

func TestResourceApiWhitelistReadNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ClusterApiWhitelist(testhelpers.TestProjectID, "gone"),
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesApiWhitelist()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("gone")

	if diags := res.ReadContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "" {
		t.Error("expected ID to be cleared on 404")
	}
}

func TestResourceApiWhitelistDeleteSendsEmptyList(t *testing.T) {
	var body dto.UpdateClusterApiWhitelistRequest
	var raw map[string]interface{}
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.ClusterApiWhitelist(testhelpers.TestProjectID, "clu-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				var buf json.RawMessage
				json.NewDecoder(r.Body).Decode(&buf)
				json.Unmarshal(buf, &body)
				json.Unmarshal(buf, &raw)
				testhelpers.JSONHandler(t, http.StatusOK, dto.ClusterApiWhitelistResponse{})(w, r)
			},
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.ClusterWithID(testhelpers.TestProjectID, "clu-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: dto.K8sCluster{ID: "clu-001", Status: "active"}}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesApiWhitelist()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"cluster_id": "clu-001"})
	d.SetId("clu-001")

	if diags := res.DeleteContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if v, ok := raw["entries"]; !ok || v == nil {
		t.Errorf("expected entries to be sent as an empty array, got %v", raw)
	}
	if len(body.Entries) != 0 {
		t.Errorf("expected no entries, got %v", body.Entries)
	}
}
