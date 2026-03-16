package networkinterface

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceNetworkInterfaceRead_ByID(t *testing.T) {
	ni := testNetworkInterface()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "nic-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "nic-001" {
		t.Errorf("expected ID nic-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-nic" {
		t.Errorf("expected name test-nic, got %s", v)
	}
	if v := d.Get("network_id").(string); v != "net-001" {
		t.Errorf("expected network_id net-001, got %s", v)
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("ip_address").(string); v != "10.0.0.5" {
		t.Errorf("expected ip_address 10.0.0.5, got %s", v)
	}
	if v := d.Get("mac_address").(string); v != "fa:16:3e:aa:bb:cc" {
		t.Errorf("expected mac_address fa:16:3e:aa:bb:cc, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("port_security_enabled").(bool); !v {
		t.Error("expected port_security_enabled true, got false")
	}
	if v := d.Get("network_type").(string); v != "vxlan" {
		t.Errorf("expected network_type vxlan, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test network interface" {
		t.Errorf("expected description 'a test network interface', got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceNetworkInterfaceRead_ByName(t *testing.T) {
	ni := testNetworkInterface()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListNetworkInterfacesResponse{
				NetworkInterfaces: []dto.NetworkInterface{ni},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-nic",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "nic-001" {
		t.Errorf("expected ID nic-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-nic" {
		t.Errorf("expected name test-nic, got %s", v)
	}
}

func TestDataSourceNetworkInterfaceRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListNetworkInterfacesResponse{
				NetworkInterfaces: []dto.NetworkInterface{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent network interface, got none")
	}
}
