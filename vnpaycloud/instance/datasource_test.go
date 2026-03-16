package instance

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceInstanceRead_ByID(t *testing.T) {
	inst := testInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/inst-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "inst-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "inst-001" {
		t.Errorf("expected ID inst-001, got %s", d.Id())
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

func TestDataSourceInstanceRead_ByName(t *testing.T) {
	inst := testInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListInstancesResponse{
				Instances: []dto.Instance{inst},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-instance",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "inst-001" {
		t.Errorf("expected ID inst-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-instance" {
		t.Errorf("expected name test-instance, got %s", v)
	}
}

func TestDataSourceInstanceRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListInstancesResponse{
				Instances: []dto.Instance{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent instance, got none")
	}
}

func TestDataSourceInstancesRead(t *testing.T) {
	inst1 := testInstance()
	inst2 := dto.Instance{
		ID:            "inst-002",
		Name:          "test-instance-2",
		ImageName:     "centos-9",
		FlavorName:    "v1.medium",
		Status:        "active",
		PowerState:    "running",
		KeyPairID:     "kp-002",
		ServerGroupID: "",
		ZoneID:        testhelpers.TestZoneID,
		CreatedAt:     "2025-01-16T12:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListInstancesResponse{
				Instances: []dto.Instance{inst1, inst2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	instances := d.Get("instances").([]interface{})
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	first := instances[0].(map[string]interface{})
	if first["id"] != "inst-001" {
		t.Errorf("expected first instance id inst-001, got %v", first["id"])
	}
	if first["name"] != "test-instance" {
		t.Errorf("expected first instance name test-instance, got %v", first["name"])
	}

	second := instances[1].(map[string]interface{})
	if second["id"] != "inst-002" {
		t.Errorf("expected second instance id inst-002, got %v", second["id"])
	}
	if second["name"] != "test-instance-2" {
		t.Errorf("expected second instance name test-instance-2, got %v", second["name"])
	}
}
