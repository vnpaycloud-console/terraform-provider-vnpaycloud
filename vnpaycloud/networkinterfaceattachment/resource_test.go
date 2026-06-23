package networkinterfaceattachment

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testNetworkInterface returns a dto.NetworkInterface in attached state for use in tests.
func testNetworkInterface() dto.NetworkInterface {
	return dto.NetworkInterface{
		ID:                  "nic-001",
		Name:                "test-nic",
		NetworkID:           "net-001",
		SubnetID:            "subnet-001",
		IPAddress:           "10.0.0.5",
		MACAddress:          "fa:16:3e:aa:bb:cc",
		Status:              "active",
		SecurityGroups:      []string{"sg-001"},
		PortSecurityEnabled: true,
		NetworkType:         "vxlan",
		Description:         "a test network interface",
		CreatedAt:           "2025-01-15T10:00:00Z",
		ProjectID:           testhelpers.TestProjectID,
		ZoneID:              testhelpers.TestZoneID,
	}
}

func testInstance(nicIDs ...string) dto.Instance {
	return dto.Instance{
		ID:                  "srv-001",
		Name:                "test-server",
		Status:              "active",
		NetworkInterfaceIDs: nicIDs,
		ProjectID:           testhelpers.TestProjectID,
		ZoneID:              testhelpers.TestZoneID,
	}
}

func TestResourceNetworkInterfaceAttachmentCreate(t *testing.T) {
	ni := testNetworkInterface()
	inst := testInstance("nic-001")

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001/attach",
			Handler: testhelpers.EmptyHandler(http.StatusAccepted),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/srv-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterfaceAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"network_interface_id": "nic-001",
		"server_id":            "srv-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "nic-001" {
		t.Errorf("expected ID nic-001, got %s", d.Id())
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("ip_address").(string); v != "10.0.0.5" {
		t.Errorf("expected ip_address 10.0.0.5, got %s", v)
	}
}

func TestResourceNetworkInterfaceAttachmentRead(t *testing.T) {
	ni := testNetworkInterface()
	inst := testInstance("nic-001")

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/srv-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterfaceAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"network_interface_id": "nic-001",
		"server_id":            "srv-001",
	})
	d.SetId("nic-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("network_interface_id").(string); v != "nic-001" {
		t.Errorf("expected network_interface_id nic-001, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("ip_address").(string); v != "10.0.0.5" {
		t.Errorf("expected ip_address 10.0.0.5, got %s", v)
	}
}

func TestResourceNetworkInterfaceAttachmentRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/srv-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: testInstance("nic-gone")}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterfaceAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"network_interface_id": "nic-gone",
		"server_id":            "srv-001",
	})
	d.SetId("nic-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceNetworkInterfaceAttachmentDelete(t *testing.T) {
	detachCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001/detach",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				detachCalled = true
				w.WriteHeader(http.StatusAccepted)
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/srv-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if detachCalled {
					testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: testInstance()})(w, r)
					return
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: testInstance("nic-001")})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterfaceAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"network_interface_id": "nic-001",
		"server_id":            "srv-001",
	})
	d.SetId("nic-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !detachCalled {
		t.Error("expected detach POST to have been called")
	}
}

func TestResourceNetworkInterfaceAttachmentRead_DetachedClearsState(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/instances/srv-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InstanceResponse{Instance: testInstance()}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterfaceAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"network_interface_id": "nic-001",
		"server_id":            "srv-001",
	})
	d.SetId("nic-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected attachment ID to be cleared after NIC is no longer on server, got %s", d.Id())
	}
}

func TestResourceNetworkInterfaceAttachmentCreate_APIError(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001/attach",
			Handler: testhelpers.EmptyHandler(http.StatusBadRequest),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterfaceAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"network_interface_id": "nic-001",
		"server_id":            "srv-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for attach API failure, got none")
	}
}
