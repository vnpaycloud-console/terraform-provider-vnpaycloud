// This set of code handles all functions required to configure networking
// on an vnpaycloud_compute_instance_v2 resource.
//
// This is a complicated task because it's not possible to obtain all
// information in a single API call. In fact, it even traverses multiple
// VNPAYCLOUD services.
//
// The end result, from the user's point of view, is a structured set of
// understandable network information within the instance resource.
package server

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type InstanceNIC struct {
	FixedIPv4 string
	FixedIPv6 string
	MAC       string
}

type InstanceAddresses struct {
	NetworkName  string
	InstanceNICs []InstanceNIC
}

type InstanceNetwork struct {
	UUID          string
	Name          string
	Port          string
	FixedIP       string
	AccessNetwork bool
}

func getAllInstanceNetworks(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]InstanceNetwork, error) {
	networks := d.Get("network").([]interface{})

	instanceNetworks := make([]InstanceNetwork, 0, len(networks))
	for _, v := range networks {
		network := v.(map[string]interface{})
		networkID := network["uuid"].(string)
		networkName := network["name"].(string)
		portID := network["port"].(string)

		if networkID == "" && networkName == "" && portID == "" {
			return nil, fmt.Errorf(
				"At least one of network.uuid, network.name, or network.port must be set")
		}

		if networkID != "" && networkName != "" {
			v := InstanceNetwork{
				UUID:          networkID,
				Name:          networkName,
				Port:          portID,
				FixedIP:       network["fixed_ip_v4"].(string),
				AccessNetwork: network["access_network"].(bool),
			}
			instanceNetworks = append(instanceNetworks, v)
			continue
		}

		queryType := "name"
		queryTerm := networkName
		if networkID != "" {
			queryType = "id"
			queryTerm = networkID
		}
		if portID != "" {
			queryType = "port"
			queryTerm = portID
		}

		networkInfo, err := getInstanceNetworkInfo(ctx, d, meta, queryType, queryTerm)
		if err != nil {
			return nil, err
		}

		v := InstanceNetwork{
			Port:          portID,
			FixedIP:       network["fixed_ip_v4"].(string),
			AccessNetwork: network["access_network"].(bool),
		}
		if networkInfo["uuid"] != nil {
			v.UUID = networkInfo["uuid"].(string)
		}
		if networkInfo["name"] != nil {
			v.Name = networkInfo["name"].(string)
		}

		instanceNetworks = append(instanceNetworks, v)
	}

	log.Printf("[DEBUG] getAllInstanceNetworks: %#v", instanceNetworks)
	return instanceNetworks, nil
}

func getInstanceNetworkInfo(ctx context.Context, d *schema.ResourceData, meta interface{}, queryType, queryTerm string) (map[string]interface{}, error) {
	config := meta.(*config.Config)
	networkClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	networkInfo, err := getInstanceNetworkInfoNeutron(ctx, networkClient, queryType, queryTerm)
	if err != nil {
		return nil, fmt.Errorf("Error trying to get network information from the Network API: %s", err)
	}

	return networkInfo, nil
}

func getInstanceNetworkInfoNeutron(ctx context.Context, networkClient *client.Client, queryType, queryTerm string) (map[string]interface{}, error) {
	if queryType == "port" {
		listOpts := dto.ListPortOpts{
			ID: queryTerm,
		}

		portsResp := &dto.ListPortsResponse{}
		_, err := networkClient.All(ctx, client.ApiPath.PortWithParams(listOpts), portsResp, nil)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve networks from the Network API: %s", err)
		}

		var port dto.Port
		switch len(portsResp.Ports) {
		case 0:
			return nil, fmt.Errorf("Could not find any matching port for %s %s", queryType, queryTerm)
		case 1:
			port = portsResp.Ports[0]
		default:
			return nil, fmt.Errorf("More than one port found for %s %s", queryType, queryTerm)
		}

		queryType = "id"
		queryTerm = port.NetworkID
	}

	listNetworkOpts := dto.ListNetworkParams{
		Status: "ACTIVE",
	}

	switch queryType {
	case "name":
		listNetworkOpts.Name = queryTerm
	default:
		listNetworkOpts.ID = queryTerm
	}

	networksResp := &dto.ListNetworksResponse{}
	_, err := networkClient.All(ctx, client.ApiPath.NetworkWithParams(listNetworkOpts), networksResp, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve networks from the Network API: %s", err)
	}

	var network dto.Network
	switch len(networksResp.Networks) {
	case 0:
		return nil, fmt.Errorf("Could not find any matching network for %s %s", queryType, queryTerm)
	case 1:
		network = networksResp.Networks[0]
	default:
		return nil, fmt.Errorf("More than one network found for %s %s", queryType, queryTerm)
	}

	v := map[string]interface{}{
		"uuid": network.ID,
		"name": network.Name,
	}

	log.Printf("[DEBUG] getInstanceNetworkInfoNeutron: %#v", v)
	return v, nil
}

func getInstanceAddresses(addresses map[string]interface{}) []InstanceAddresses {
	allInstanceAddresses := make([]InstanceAddresses, 0, len(addresses))

	networkNames := make([]string, len(addresses))
	i := 0
	for k := range addresses {
		networkNames[i] = k
		i++
	}

	if len(networkNames) == 2 {
		if networkNames[0] == "private" && networkNames[1] == "public" {
			networkNames[0] = "public"
			networkNames[1] = "private"
		}
	}

	for _, networkName := range networkNames {
		v := addresses[networkName]
		instanceAddresses := InstanceAddresses{
			NetworkName: networkName,
		}

		for _, v := range v.([]interface{}) {
			instanceNIC := InstanceNIC{}
			var exists bool

			v := v.(map[string]interface{})
			if v, ok := v["OS-EXT-IPS-MAC:mac_addr"].(string); ok {
				instanceNIC.MAC = v
			}

			if v["OS-EXT-IPS:type"] == "fixed" || v["OS-EXT-IPS:type"] == nil {
				switch v["version"].(float64) {
				case 6:
					instanceNIC.FixedIPv6 = fmt.Sprintf("[%s]", v["addr"].(string))
				default:
					instanceNIC.FixedIPv4 = v["addr"].(string)
				}
			}

			for i, v := range instanceAddresses.InstanceNICs {
				if v.MAC == instanceNIC.MAC {
					exists = true
					if instanceNIC.FixedIPv6 != "" {
						instanceAddresses.InstanceNICs[i].FixedIPv6 = instanceNIC.FixedIPv6
					}
					if instanceNIC.FixedIPv4 != "" {
						instanceAddresses.InstanceNICs[i].FixedIPv4 = instanceNIC.FixedIPv4
					}
				}
			}

			if !exists {
				instanceAddresses.InstanceNICs = append(instanceAddresses.InstanceNICs, instanceNIC)
			}
		}

		allInstanceAddresses = append(allInstanceAddresses, instanceAddresses)
	}

	log.Printf("[DEBUG] Addresses: %#v", addresses)
	log.Printf("[DEBUG] allInstanceAddresses: %#v", allInstanceAddresses)

	return allInstanceAddresses
}

func expandInstanceNetworks(allInstanceNetworks []InstanceNetwork) []dto.ServerNetwork {
	networks := make([]dto.ServerNetwork, 0, len(allInstanceNetworks))
	for _, v := range allInstanceNetworks {
		n := dto.ServerNetwork{
			UUID:    v.UUID,
			Port:    v.Port,
			FixedIP: v.FixedIP,
		}
		networks = append(networks, n)
	}

	return networks
}

func flattenInstanceNetworks(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]map[string]interface{}, error) {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Error creating VNPAYCLOUD client: %s", err)
	}

	serverResp := &dto.GetServerResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.ServerWithId(d.Id()), serverResp, nil)
	if err != nil {
		return nil, util.CheckDeleted(d, err, "server")
	}

	allInstanceAddresses := getInstanceAddresses(serverResp.Server.Addresses)
	allInstanceNetworks, err := getAllInstanceNetworks(ctx, d, meta)
	if err != nil {
		return nil, err
	}

	networks := []map[string]interface{}{}

	if len(allInstanceNetworks) == 0 {
		for _, instanceAddresses := range allInstanceAddresses {
			for _, instanceNIC := range instanceAddresses.InstanceNICs {
				v := map[string]interface{}{
					"name":        instanceAddresses.NetworkName,
					"fixed_ip_v4": instanceNIC.FixedIPv4,
					"fixed_ip_v6": instanceNIC.FixedIPv6,
					"mac":         instanceNIC.MAC,
				}

				networkInfo, err := getInstanceNetworkInfo(ctx, d, meta, "name", instanceAddresses.NetworkName)
				if err != nil {
					log.Printf("[WARN] Error getting default network uuid: %s", err)
				} else {
					if v["uuid"] != nil {
						v["uuid"] = networkInfo["uuid"].(string)
					} else {
						log.Printf("[WARN] Could not get default network uuid")
					}
				}

				networks = append(networks, v)
			}
		}

		log.Printf("[DEBUG] flattenInstanceNetworks: %#v", networks)
		return networks, nil
	}

	for _, instanceNetwork := range allInstanceNetworks {
		for _, instanceAddresses := range allInstanceAddresses {
			if len(instanceAddresses.InstanceNICs) == 0 {
				continue
			}

			if instanceNetwork.Name == instanceAddresses.NetworkName {
				instanceNIC := instanceAddresses.InstanceNICs[0]
				copy(instanceAddresses.InstanceNICs, instanceAddresses.InstanceNICs[1:])
				v := map[string]interface{}{
					"name":           instanceAddresses.NetworkName,
					"fixed_ip_v4":    instanceNIC.FixedIPv4,
					"fixed_ip_v6":    instanceNIC.FixedIPv6,
					"mac":            instanceNIC.MAC,
					"uuid":           instanceNetwork.UUID,
					"port":           instanceNetwork.Port,
					"access_network": instanceNetwork.AccessNetwork,
				}
				networks = append(networks, v)
			}
		}
	}

	log.Printf("[DEBUG] flattenInstanceNetworks: %#v", networks)
	return networks, nil
}

func getInstanceAccessAddresses(d *schema.ResourceData, networks []map[string]interface{}) (string, string) {
	var hostv4, hostv6 string

	for _, n := range networks {
		var accessNetwork bool

		if an, ok := n["access_network"].(bool); ok && an {
			accessNetwork = true
		}

		if fixedIPv4, ok := n["fixed_ip_v4"].(string); ok && fixedIPv4 != "" {
			if hostv4 == "" || accessNetwork {
				hostv4 = fixedIPv4
			}
		}

		if fixedIPv6, ok := n["fixed_ip_v6"].(string); ok && fixedIPv6 != "" {
			if hostv6 == "" || accessNetwork {
				hostv6 = fixedIPv6
			}
		}
	}

	log.Printf("[DEBUG] VNPAYCLOUD Instance Network Access Addresses: %s, %s", hostv4, hostv6)

	return hostv4, hostv6
}

func computeInstanceReadTags(d *schema.ResourceData, tags []string) {
	util.ExpandObjectReadTags(d, tags)
}

func computeInstanceUpdateTags(d *schema.ResourceData) []string {
	return util.ExpandObjectUpdateTags(d)
}

func computeInstanceTags(d *schema.ResourceData) []string {
	return util.ExpandObjectTags(d)
}
