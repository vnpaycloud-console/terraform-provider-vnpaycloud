package subnetsnat

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testSubnetWithSNAT returns a dto.Subnet with SNAT enabled for use in tests.
func testSubnetWithSNAT() dto.Subnet {
	return dto.Subnet{
		ID:           "subnet-001",
		Name:         "test-subnet",
		VpcID:        "vpc-001",
		CIDR:         "10.0.1.0/24",
		GatewayIP:    "10.0.1.1",
		EnableDHCP:   true,
		EnableSnat:   true,
		ExternalIpID: "fip-001",
		Status:       "active",
		CreatedAt:    "2025-01-15T10:00:00Z",
		ProjectID:    testhelpers.TestProjectID,
		ZoneID:       testhelpers.TestZoneID,
	}
}

func TestResourceSubnetSNATCreate(t *testing.T) {
	subnet := testSubnetWithSNAT()
	subnetBeforeEnable := subnet
	subnetBeforeEnable.EnableSnat = false
	subnetBeforeEnable.ExternalIpID = ""
	getCalls := 0
	enableCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				getCalls++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				resp := dto.SubnetResponse{Subnet: subnet}
				if getCalls == 1 {
					resp.Subnet = subnetBeforeEnable
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					t.Errorf("failed to encode response body: %v", err)
				}
			},
		},
		{
			Method:  "PUT",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001/enable-snat",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				enableCalled = true
				w.WriteHeader(http.StatusOK)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnetSNAT()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"subnet_id":      "subnet-001",
		"floating_ip_id": "fip-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "subnet-001/snat" {
		t.Errorf("expected ID subnet-001/snat, got %s", d.Id())
	}
	if getCalls != 2 {
		t.Errorf("expected two subnet GET calls, got %d", getCalls)
	}
	if !enableCalled {
		t.Error("expected enable-snat PUT to have been called")
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("floating_ip_id").(string); v != "fip-001" {
		t.Errorf("expected floating_ip_id fip-001, got %s", v)
	}
}

func TestResourceSubnetSNATCreate_DuplicateSameFloatingIP(t *testing.T) {
	subnet := testSubnetWithSNAT()
	enableCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: subnet}),
		},
		{
			Method:  "PUT",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001/enable-snat",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				enableCalled = true
				w.WriteHeader(http.StatusInternalServerError)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnetSNAT()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"subnet_id":      "subnet-001",
		"floating_ip_id": "fip-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected duplicate SNAT create to fail")
	}
	if !strings.Contains(diags[0].Summary, "duplicate vnpaycloud_subnet_snat resources are not allowed") {
		t.Fatalf("expected duplicate error, got %v", diags)
	}
	if enableCalled {
		t.Error("did not expect enable-snat PUT for duplicate SNAT")
	}
}

func TestResourceSubnetSNATCreate_AlreadyEnabledDifferentFloatingIP(t *testing.T) {
	subnet := testSubnetWithSNAT()
	enableCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: subnet}),
		},
		{
			Method:  "PUT",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001/enable-snat",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				enableCalled = true
				w.WriteHeader(http.StatusInternalServerError)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnetSNAT()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"subnet_id":      "subnet-001",
		"floating_ip_id": "fip-002",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected already-enabled SNAT create to fail")
	}
	if !strings.Contains(diags[0].Summary, "disable it before switching to floating IP fip-002") {
		t.Fatalf("expected switch error, got %v", diags)
	}
	if enableCalled {
		t.Error("did not expect enable-snat PUT when SNAT is already enabled")
	}
}

func TestResourceSubnetSNATRead_Enabled(t *testing.T) {
	subnet := testSubnetWithSNAT()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: subnet}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnetSNAT()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"subnet_id":      "subnet-001",
		"floating_ip_id": "",
	})
	d.SetId("subnet-001/snat")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "subnet-001/snat" {
		t.Errorf("expected ID to remain subnet-001/snat, got %s", d.Id())
	}
	if v := d.Get("floating_ip_id").(string); v != "fip-001" {
		t.Errorf("expected floating_ip_id fip-001, got %s", v)
	}
}

func TestResourceSubnetSNATRead_Disabled(t *testing.T) {
	subnet := testSubnetWithSNAT()
	subnet.EnableSnat = false
	subnet.ExternalIpID = ""

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: subnet}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnetSNAT()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"subnet_id":      "subnet-001",
		"floating_ip_id": "fip-001",
	})
	d.SetId("subnet-001/snat")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared when SNAT is disabled, got %s", d.Id())
	}
}

func TestResourceSubnetSNATDelete(t *testing.T) {
	disableCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "PUT",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001/disable-snat",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				disableCalled = true
				w.WriteHeader(http.StatusOK)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnetSNAT()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"subnet_id":      "subnet-001",
		"floating_ip_id": "fip-001",
	})
	d.SetId("subnet-001/snat")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !disableCalled {
		t.Error("expected disable-snat PUT to have been called")
	}
}
