package volumetype

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceVolumeTypeRead_ByID(t *testing.T) {
	vtResp := dto.VolumeTypeResponse{
		VolumeType: dto.VolumeType{
			ID:            "vt-001",
			Name:          "ssd-premium",
			IOPS:          3000,
			IsEncrypted:   true,
			IsMultiattach: false,
			Zone:          "zone-a",
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.VolumeTypeWithID("vt-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, vtResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceVolumeType()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"id": "vt-001",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vt-001" {
		t.Errorf("expected ID 'vt-001', got '%s'", d.Id())
	}
	if v := d.Get("name").(string); v != "ssd-premium" {
		t.Errorf("expected name 'ssd-premium', got '%s'", v)
	}
	if v := d.Get("iops").(int); v != 3000 {
		t.Errorf("expected iops 3000, got %d", v)
	}
	if v := d.Get("is_encrypted").(bool); !v {
		t.Error("expected is_encrypted true, got false")
	}
	if v := d.Get("is_multiattach").(bool); v {
		t.Error("expected is_multiattach false, got true")
	}
	if v := d.Get("zone").(string); v != "zone-a" {
		t.Errorf("expected zone 'zone-a', got '%s'", v)
	}
}

func TestDataSourceVolumeTypeRead_ByName(t *testing.T) {
	listResp := dto.ListVolumeTypesResponse{
		VolumeTypes: []dto.VolumeType{
			{
				ID:            "vt-001",
				Name:          "ssd-premium",
				IOPS:          3000,
				IsEncrypted:   true,
				IsMultiattach: false,
				Zone:          "zone-a",
			},
			{
				ID:            "vt-002",
				Name:          "hdd-standard",
				IOPS:          500,
				IsEncrypted:   false,
				IsMultiattach: true,
				Zone:          "zone-a",
			},
		},
	}

	// The client calls GET /v2/iac/volume-types?zone=test-zone-id.
	// http.ServeMux strips query params before matching, so register the base path.
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/volume-types",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceVolumeType()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "hdd-standard",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vt-002" {
		t.Errorf("expected ID 'vt-002', got '%s'", d.Id())
	}
	if v := d.Get("name").(string); v != "hdd-standard" {
		t.Errorf("expected name 'hdd-standard', got '%s'", v)
	}
	if v := d.Get("iops").(int); v != 500 {
		t.Errorf("expected iops 500, got %d", v)
	}
	if v := d.Get("is_encrypted").(bool); v {
		t.Error("expected is_encrypted false, got true")
	}
	if v := d.Get("is_multiattach").(bool); !v {
		t.Error("expected is_multiattach true, got false")
	}
	if v := d.Get("zone").(string); v != "zone-a" {
		t.Errorf("expected zone 'zone-a', got '%s'", v)
	}
}

func TestDataSourceVolumeTypesRead(t *testing.T) {
	listResp := dto.ListVolumeTypesResponse{
		VolumeTypes: []dto.VolumeType{
			{
				ID:            "vt-001",
				Name:          "ssd-premium",
				IOPS:          3000,
				IsEncrypted:   true,
				IsMultiattach: false,
				Zone:          "zone-a",
			},
			{
				ID:            "vt-002",
				Name:          "hdd-standard",
				IOPS:          500,
				IsEncrypted:   false,
				IsMultiattach: true,
				Zone:          "zone-a",
			},
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/volume-types",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceVolumeTypes()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	expectedID := "volume-types-" + testhelpers.TestZoneID
	if d.Id() != expectedID {
		t.Errorf("expected ID '%s', got '%s'", expectedID, d.Id())
	}

	volumeTypes := d.Get("volume_types").([]interface{})
	if len(volumeTypes) != 2 {
		t.Fatalf("expected 2 volume types, got %d", len(volumeTypes))
	}

	first := volumeTypes[0].(map[string]interface{})
	if first["id"] != "vt-001" {
		t.Errorf("expected first volume type id 'vt-001', got '%s'", first["id"])
	}
	if first["name"] != "ssd-premium" {
		t.Errorf("expected first volume type name 'ssd-premium', got '%s'", first["name"])
	}
	if first["is_encrypted"] != true {
		t.Error("expected first volume type is_encrypted true, got false")
	}

	second := volumeTypes[1].(map[string]interface{})
	if second["id"] != "vt-002" {
		t.Errorf("expected second volume type id 'vt-002', got '%s'", second["id"])
	}
	if second["name"] != "hdd-standard" {
		t.Errorf("expected second volume type name 'hdd-standard', got '%s'", second["name"])
	}
	if second["is_multiattach"] != true {
		t.Error("expected second volume type is_multiattach true, got false")
	}
}
