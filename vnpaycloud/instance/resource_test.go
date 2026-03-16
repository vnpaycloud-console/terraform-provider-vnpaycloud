package instance

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testInstance returns a fully populated dto.Instance for use in tests.
func testInstance() dto.Instance {
	return dto.Instance{
		ID:                  "inst-001",
		Name:                "test-instance",
		ImageName:           "ubuntu-22.04",
		ImageID:             "img-001",
		FlavorName:          "v1.small",
		VolumeIDs:           []string{"vol-001"},
		Status:              "active",
		PowerState:          "running",
		NetworkInterfaceIDs: []string{"ni-001"},
		KeyPairID:           "kp-001",
		SecurityGroupIDs:    []string{"sg-001"},
		ServerGroupID:       "sgrp-001",
		CreatedAt:           "2025-01-15T10:00:00Z",
		ProjectID:           testhelpers.TestProjectID,
		ZoneID:              testhelpers.TestZoneID,
	}
}

func TestResourceInstanceCreate(t *testing.T) {
	inst := testInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/inst-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                  "test-instance",
		"image":                 "ubuntu-22.04",
		"snapshot_id":           "",
		"flavor":                "v1.small",
		"root_disk_gb":          20,
		"root_disk_type":        "SSD",
		"key_pair":              "kp-001",
		"server_group_id":       "sgrp-001",
		"user_data":             "",
		"is_user_data_base64":   false,
		"is_custom_flavor":      false,
		"custom_vcpus":          0,
		"custom_ram_mb":         0,
		"security_groups":       []interface{}{"sg-001"},
		"network_interface_ids": []interface{}{"ni-001"},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "inst-001" {
		t.Errorf("expected ID inst-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-instance" {
		t.Errorf("expected name test-instance, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("power_state").(string); v != "running" {
		t.Errorf("expected power_state running, got %s", v)
	}
}

func TestResourceInstanceRead(t *testing.T) {
	inst := testInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/inst-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                "test-instance",
		"image":               "",
		"snapshot_id":         "",
		"flavor":              "",
		"root_disk_gb":        20,
		"root_disk_type":      "SSD",
		"key_pair":            "",
		"server_group_id":     "",
		"user_data":           "",
		"is_user_data_base64": false,
		"is_custom_flavor":    false,
		"custom_vcpus":        0,
		"custom_ram_mb":       0,
	})
	d.SetId("inst-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-instance" {
		t.Errorf("expected name test-instance, got %s", v)
	}
	if v := d.Get("image_name").(string); v != "ubuntu-22.04" {
		t.Errorf("expected image_name ubuntu-22.04, got %s", v)
	}
	if v := d.Get("image_id").(string); v != "img-001" {
		t.Errorf("expected image_id img-001, got %s", v)
	}
	if v := d.Get("flavor_name").(string); v != "v1.small" {
		t.Errorf("expected flavor_name v1.small, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("power_state").(string); v != "running" {
		t.Errorf("expected power_state running, got %s", v)
	}
	if v := d.Get("key_pair").(string); v != "kp-001" {
		t.Errorf("expected key_pair kp-001, got %s", v)
	}
	if v := d.Get("server_group_id").(string); v != "sgrp-001" {
		t.Errorf("expected server_group_id sgrp-001, got %s", v)
	}
	if v := d.Get("zone_id").(string); v != testhelpers.TestZoneID {
		t.Errorf("expected zone_id %s, got %s", testhelpers.TestZoneID, v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceInstanceRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/inst-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                "",
		"image":               "",
		"snapshot_id":         "",
		"flavor":              "",
		"root_disk_gb":        20,
		"root_disk_type":      "SSD",
		"key_pair":            "",
		"server_group_id":     "",
		"user_data":           "",
		"is_user_data_base64": false,
		"is_custom_flavor":    false,
		"custom_vcpus":        0,
		"custom_ram_mb":       0,
	})
	d.SetId("inst-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceInstanceDelete(t *testing.T) {
	inst := testInstance()
	inst.Status = "active"

	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/instances/inst-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: inst})(w, r)
				case "DELETE":
					deletedCalled = true
					w.WriteHeader(http.StatusAccepted)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                "test-instance",
		"image":               "",
		"snapshot_id":         "",
		"flavor":              "",
		"root_disk_gb":        20,
		"root_disk_type":      "SSD",
		"key_pair":            "",
		"server_group_id":     "",
		"user_data":           "",
		"is_user_data_base64": false,
		"is_custom_flavor":    false,
		"custom_vcpus":        0,
		"custom_ram_mb":       0,
	})
	d.SetId("inst-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
