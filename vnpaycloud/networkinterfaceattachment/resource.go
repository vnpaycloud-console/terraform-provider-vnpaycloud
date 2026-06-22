package networkinterfaceattachment

import (
	"context"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

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

	_, err := cfg.Client.Post(ctx, client.ApiPath.NetworkInterfaceAttach(cfg.ProjectID, nicID), attachOpts, nil, nil)
	if err != nil {
		return diag.Errorf("Error attaching network interface %s to server %s: %s", nicID, serverID, err)
	}

	d.SetId(nicID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"attaching"},
		Target:     []string{"attached"},
		Refresh:    serverNICAttachRefreshFunc(ctx, cfg.Client, cfg.ProjectID, serverID, nicID, true),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for network interface %s to attach to server %s: %s", nicID, serverID, err)
	}

	return resourceNetworkInterfaceAttachmentRead(ctx, d, meta)
}

func resourceNetworkInterfaceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	serverID := d.Get("server_id").(string)
	attached, err := serverHasNIC(ctx, cfg.Client, cfg.ProjectID, serverID, d.Id())
	if err != nil {
		if util.ResponseCodeIs(err, http.StatusNotFound) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error checking server %s for network interface %s attachment: %s", serverID, d.Id(), err)
	}
	if !attached {
		d.SetId("")
		return nil
	}

	niResp := &dto.NetworkInterfaceResponse{}
	_, err = cfg.Client.Get(ctx, client.ApiPath.NetworkInterfaceWithID(cfg.ProjectID, d.Id()), niResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_network_interface_attachment"))
	}

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

	_, err := cfg.Client.Post(ctx, client.ApiPath.NetworkInterfaceDetach(cfg.ProjectID, d.Id()), detachOpts, nil, nil)
	if err != nil {
		if util.ResponseCodeIs(err, http.StatusNotFound) {
			return nil
		}
		return diag.Errorf("Error detaching network interface %s from server %s: %s", d.Id(), serverID, err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"detaching"},
		Target:     []string{"detached"},
		Refresh:    serverNICAttachRefreshFunc(ctx, cfg.Client, cfg.ProjectID, serverID, d.Id(), false),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for network interface %s to detach: %s", d.Id(), err)
	}

	return nil
}

func serverHasNIC(ctx context.Context, c *client.Client, projectID, serverID, nicID string) (bool, error) {
	instResp := &dto.InstanceResponse{}
	if _, err := c.Get(ctx, client.ApiPath.InstanceWithID(projectID, serverID), instResp, nil); err != nil {
		return false, err
	}
	for _, id := range instResp.Instance.NetworkInterfaceIDs {
		if id == nicID {
			return true, nil
		}
	}
	return false, nil
}

func serverNICAttachRefreshFunc(ctx context.Context, c *client.Client, projectID, serverID, nicID string, wantAttached bool) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attached, err := serverHasNIC(ctx, c, projectID, serverID, nicID)
		if err != nil {
			return nil, "", err
		}
		if wantAttached {
			if attached {
				return serverID, "attached", nil
			}
			return serverID, "attaching", nil
		}
		if !attached {
			return serverID, "detached", nil
		}
		return serverID, "detaching", nil
	}
}
