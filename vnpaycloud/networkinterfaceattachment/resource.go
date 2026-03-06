package networkinterfaceattachment

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNetworkInterfaceAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInterfaceAttachmentCreate,
		ReadContext:   resourceNetworkInterfaceAttachmentRead,
		DeleteContext: resourceNetworkInterfaceAttachmentDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkInterfaceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	nicID := d.Get("network_interface_id").(string)
	serverID := d.Get("server_id").(string)

	attachOpts := dto.AttachNetworkInterfaceRequest{
		ServerID: serverID,
	}

	tflog.Debug(ctx, "vnpaycloud_network_interface_attachment create", map[string]interface{}{
		"network_interface_id": nicID,
		"server_id":            serverID,
	})

	_, err := cfg.Client.Post(ctx, client.ApiPath.NetworkInterfaceAttach(cfg.ProjectID, nicID), attachOpts, nil, nil)
	if err != nil {
		return diag.Errorf("Error attaching network interface %s to server %s: %s", nicID, serverID, err)
	}

	d.SetId(nicID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"attaching", "build", "creating"},
		Target:     []string{"active", "attached"},
		Refresh:    nicStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, nicID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for network interface %s to attach: %s", nicID, err)
	}

	return resourceNetworkInterfaceAttachmentRead(ctx, d, meta)
}

func resourceNetworkInterfaceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	niResp := &dto.NetworkInterfaceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkInterfaceWithID(cfg.ProjectID, d.Id()), niResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_network_interface_attachment"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_network_interface_attachment "+d.Id(), map[string]interface{}{
		"network_interface": niResp.NetworkInterface,
	})

	d.Set("network_interface_id", niResp.NetworkInterface.ID)
	d.Set("status", niResp.NetworkInterface.Status)
	d.Set("ip_address", niResp.NetworkInterface.IPAddress)

	return nil
}

func resourceNetworkInterfaceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	serverID := d.Get("server_id").(string)

	detachOpts := dto.DetachNetworkInterfaceRequest{
		ServerID: serverID,
	}

	tflog.Debug(ctx, "vnpaycloud_network_interface_attachment delete", map[string]interface{}{
		"network_interface_id": d.Id(),
		"server_id":            serverID,
	})

	_, err := cfg.Client.Post(ctx, client.ApiPath.NetworkInterfaceDetach(cfg.ProjectID, d.Id()), detachOpts, nil, nil)
	if err != nil {
		return diag.Errorf("Error detaching network interface %s from server %s: %s", d.Id(), serverID, err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"detaching"},
		Target:     []string{"active", "created", "detached"},
		Refresh:    nicStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for network interface %s to detach: %s", d.Id(), err)
	}

	return nil
}

func nicStateRefreshFunc(ctx context.Context, c *client.Client, projectID, nicID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		niResp := &dto.NetworkInterfaceResponse{}
		_, err := c.Get(ctx, client.ApiPath.NetworkInterfaceWithID(projectID, nicID), niResp, nil)
		if err != nil {
			return nil, "", err
		}
		return niResp.NetworkInterface, niResp.NetworkInterface.Status, nil
	}
}
