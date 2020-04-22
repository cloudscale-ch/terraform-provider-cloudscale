package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCloudScaleServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: getServerSchema(),
	}
}

func getServerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		// Required attributes

		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"flavor_slug": {
			Type:     schema.TypeString,
			Required: true,
		},
		"image_slug": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"ssh_keys": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		},
		"password": {
			Type:      schema.TypeString,
			Optional:  true,
			Elem:      &schema.Schema{Type: schema.TypeString},
			ForceNew:  true,
			Sensitive: true,
		},

		// Optional attributes

		"volume_size_gb": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"bulk_volume_size_gb": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"user_data": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"use_public_network": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"use_private_network": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"use_ipv6": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"allow_stopping_for_update": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"server_group_ids": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		},

		// Computed attributes

		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"volumes": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"device_path": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"uuid": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"size_gb": {
						Type:     schema.TypeInt,
						Computed: true,
					},
				},
			},
			Computed: true,
		},
		"public_ipv4_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"public_ipv6_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"private_ipv4_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"interfaces": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Required: true,
					},
					"network_uuid": {
						Type:     schema.TypeString,
						Computed: true,
						Optional: true,
					},
					"network_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"network_href": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"addresses": {
						Type: schema.TypeList,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"version": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"address": {
									Type:     schema.TypeString,
									Computed: true,
									Optional: true,
								},
								"prefix_length": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"gateway": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"reverse_ptr": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"subnet_uuid": {
									Type:     schema.TypeString,
									Computed: true,
									Optional: true,
								},
								"subnet_cidr": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"subnet_href": {
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
						Computed: true,
						Optional: true,
					},
					"no_address": {
						Type:     schema.TypeBool,
						Optional: true,
					},
				},
			},
			Computed: true,
			Optional: true,
		},
		"ssh_fingerprints": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
		},
		"ssh_host_keys": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"server_groups": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"href": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"uuid": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
			Computed: true,
		},
	}
}

func resourceServerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ServerRequest{
		Name:   d.Get("name").(string),
		Flavor: d.Get("flavor_slug").(string),
		Image:  d.Get("image_slug").(string),
	}

	sshKeys := d.Get("ssh_keys").(*schema.Set).List()
	k := make([]string, len(sshKeys))

	for i := range sshKeys {
		k[i] = sshKeys[i].(string)
	}
	opts.SSHKeys = k

	serverGroupIds := d.Get("server_group_ids").(*schema.Set).List()
	g := make([]string, len(serverGroupIds))
	for i := range serverGroupIds {
		g[i] = serverGroupIds[i].(string)
	}
	opts.ServerGroups = g

	interfacesCount := d.Get("interfaces.#").(int)
	if interfacesCount > 0 {
		interfaceRequests := createInterfaceOptions(d)
		opts.Interfaces = &interfaceRequests
	}

	if attr, ok := d.GetOk("volume_size_gb"); ok {
		opts.VolumeSizeGB = attr.(int)
	}

	if attr, ok := d.GetOk("password"); ok {
		opts.Password = attr.(string)
	}

	if attr, ok := d.GetOk("bulk_volume_size_gb"); ok {
		opts.BulkVolumeSizeGB = attr.(int)
	}

	if attr, ok := d.GetOkExists("use_public_network"); ok {
		val := attr.(bool)
		opts.UsePublicNetwork = &val
	}

	if attr, ok := d.GetOkExists("use_private_network"); ok {
		val := attr.(bool)
		opts.UsePrivateNetwork = &val
	}

	if attr, ok := d.GetOkExists("use_ipv6"); ok {
		val := attr.(bool)
		opts.UseIPV6 = &val
	}

	if attr, ok := d.GetOk("user_data"); ok {
		opts.UserData = attr.(string)
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
	}

	originalStatus := ""
	if attr, ok := d.GetOk("status"); ok {
		originalStatus = attr.(string)
	}

	log.Printf("[DEBUG] Server create configuration: %#v", opts)

	server, err := client.Servers.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}

	d.SetId(server.UUID)

	log.Printf("[INFO] Server ID %s", d.Id())

	_, err = waitForServerStatus(d, meta, []string{"changing"}, "status", "running")
	if err != nil {
		return fmt.Errorf("Error waiting for server (%s) to become ready %s", d.Id(), err)
	}

	if originalStatus == "stopped" {
		updateRequest := &cloudscale.ServerUpdateRequest{
			Status: originalStatus,
		}
		err := client.Servers.Update(context.Background(), server.UUID, updateRequest)
		if err != nil {
			return fmt.Errorf("Error stopping the server (%s) status (%s) ", server.UUID, err)
		}

		_, err = waitForServerStatus(d, meta, []string{"changing", "running"}, "status", "stopped")
		if err != nil {
			return fmt.Errorf("Error waiting for server status (%s) (%s) ", server.UUID, err)
		}
	}

	return resourceServerRead(d, meta)
}

func createInterfaceOptions(d *schema.ResourceData) []cloudscale.InterfaceRequest {
	interfacesCount := d.Get("interfaces.#").(int)
	result := make([]cloudscale.InterfaceRequest, interfacesCount)
	for i := 0; i < interfacesCount; i++ {
		prefix := fmt.Sprintf("interfaces.%d", i)
		intType := d.Get(prefix + ".type").(string)

		if intType == "public" {
			result[i] = cloudscale.InterfaceRequest{
				Network: "public",
			}
		} else {
			result[i] = createPrivateInterfaceOptions(d, prefix)
		}
	}
	return result
}

func createPrivateInterfaceOptions(d *schema.ResourceData, prefix string) cloudscale.InterfaceRequest {
	result := cloudscale.InterfaceRequest{}

	addressKey := prefix + ".addresses"
	if d.HasChange(addressKey) {
		addresses := d.Get(addressKey).([]interface{})
		if len(addresses) > 0 {
			addresses := createAddressesOptions(addresses)
			result.Addresses = &addresses
		}
		// we don't need to update the network
		return result
	}

	networkUUID := d.Get(prefix + ".network_uuid").(string)
	if networkUUID != "" {
		result.Network = networkUUID
	}

	if d.Get(prefix + ".no_address").(bool) {
		result.Addresses = &[]cloudscale.AddressRequest{}
	}

	return result
}

func createAddressesOptions(addresses []interface{}) []cloudscale.AddressRequest {
	result := make([]cloudscale.AddressRequest, len(addresses))
	for i, address := range addresses {
		a := address.(map[string]interface{})
		if a["subnet_uuid"] != "" {
			result[i].Subnet = a["subnet_uuid"].(string)
		}
		if a["address"] != "" {
			result[i].Address = a["address"].(string)
		}
	}
	return result
}

func resourceServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	id := d.Id()

	server, err := client.Servers.Get(context.Background(), id)
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving server")
	}

	d.Set("href", server.HREF)
	d.Set("name", server.Name)
	d.Set("flavor_slug", server.Flavor.Slug)
	d.Set("image_slug", server.Image.Slug)
	d.Set("zone_slug", server.Zone.Slug)

	if volumes := len(server.Volumes); volumes > 0 {
		volumesMaps := make([]map[string]interface{}, 0, volumes)
		for _, volume := range server.Volumes {
			v := make(map[string]interface{})
			v["type"] = volume.Type
			v["device_path"] = volume.DevicePath
			v["size_gb"] = volume.SizeGB
			v["uuid"] = volume.UUID
			volumesMaps = append(volumesMaps, v)
		}
		err = d.Set("volumes", volumesMaps)
		if err != nil {
			log.Printf("[DEBUG] Error setting volumes attribute: %#v, error: %#v", volumesMaps, err)
			return fmt.Errorf("Error setting volumes attribute: %#v, error: %#v", volumesMaps, err)
		}

	}
	serverGroupMaps := make([]map[string]interface{}, 0, len(server.ServerGroups))
	for _, serverGroup := range server.ServerGroups {
		g := make(map[string]interface{})
		g["uuid"] = serverGroup.UUID
		g["name"] = serverGroup.Name
		g["href"] = serverGroup.HREF
		serverGroupMaps = append(serverGroupMaps, g)
	}
	err = d.Set("server_groups", serverGroupMaps)
	if err != nil {
		log.Printf("[DEBUG] Error setting server_groups attribute: %#v, error: %#v", serverGroupMaps, err)
		return fmt.Errorf("Error setting server_groups attribute: %#v, error: %#v", serverGroupMaps, err)
	}

	d.Set("status", server.Status)

	if addrss := len(server.Interfaces); addrss > 0 {

		intsMap := make([]map[string]interface{}, 0, addrss)
		for _, intr := range server.Interfaces {

			intMap := make(map[string]interface{})

			intMap["network_href"] = intr.Network.HREF
			intMap["network_name"] = intr.Network.Name
			intMap["network_uuid"] = intr.Network.UUID

			addrssMap := make([]map[string]interface{}, 0, len(intr.Addresses))
			for _, addr := range intr.Addresses {
				i := make(map[string]interface{})
				i["address"] = addr.Address
				i["version"] = addr.Version
				i["prefix_length"] = addr.PrefixLength
				i["gateway"] = addr.Gateway
				i["reverse_ptr"] = addr.ReversePtr
				i["subnet_uuid"] = addr.Subnet.UUID
				i["subnet_cidr"] = addr.Subnet.CIDR
				i["subnet_href"] = addr.Subnet.HREF

				addrssMap = append(addrssMap, i)
			}

			intMap["type"] = intr.Type
			intMap["addresses"] = addrssMap
			intMap["no_address"] = len(addrssMap) == 0

			intsMap = append(intsMap, intMap)
		}
		err = d.Set("interfaces", intsMap)
		if err != nil {
			log.Printf("[DEBUG] Error setting interfaces attribute: %#v, error: %#v", intsMap, err)
			return fmt.Errorf("Error setting interfaces attribute: %#v, error: %#v", intsMap, err)
		}
	}

	err = d.Set("ssh_fingerprints", server.SSHFingerprints)
	if err != nil {
		log.Printf("[DEBUG] Error setting ssh_fingerprins attribute: %#v, error: %#v", server.SSHFingerprints, err)
		return fmt.Errorf("Error setting ssh_fingerprins attribute: %#v, error: %#v", server.SSHFingerprints, err)
	}

	err = d.Set("ssh_host_keys", server.SSHHostKeys)
	if err != nil {
		log.Printf("[DEBUG] Error setting ssh_host_keys attribute: %#v, error: %#v", server.SSHHostKeys, err)
		return fmt.Errorf("Error setting ssh_host_keys attribute: %#v, error: %#v", server.SSHHostKeys, err)
	}

	if publicIPV4 := findIPv4AddrByType(server, "public"); publicIPV4 != "" {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": publicIPV4,
		})
	} else {
		if publicIPV6 := findIPv6AddrByType(server, "private"); publicIPV6 != "" {
			d.SetConnInfo(map[string]string{
				"type": "ssh",
				"host": publicIPV6,
			})
		}
	}

	d.Set("public_ipv4_address", findIPv4AddrByType(server, "public"))
	d.Set("public_ipv6_address", findIPv6AddrByType(server, "public"))
	d.Set("private_ipv4_address", findIPv4AddrByType(server, "private"))

	return nil
}

func resourceServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	wantedStatus := d.Get("status").(string)
	// Since starting stoppin the server changes the state, get the wanted
	// things here.
	wantedFlavor := d.Get("flavor_slug").(string)
	wantedName := d.Get("name").(string)
	needStart := false

	if d.HasChange("volume_size_gb") {
		// The root volume is the first volume.
		volumeUUID := d.Get("volumes.0.uuid").(string)
		opts := &cloudscale.VolumeRequest{SizeGB: d.Get("volume_size_gb").(int)}
		err := client.Volumes.Update(context.Background(), volumeUUID, opts)
		if err != nil {
			return fmt.Errorf("Error scaling the Volume (%s) status (%s) ", volumeUUID, err)
		}
		d.SetPartial("volume_size_gb")
	}

	if d.HasChange("flavor_slug") {
		if !d.Get("allow_stopping_for_update").(bool) {
			return fmt.Errorf("Changing the flavor requires stopping the server. " +
				"To acknowledge this, please set allow_stopping_for_update = true in your config.")
		}

		server, err := client.Servers.Get(context.Background(), id)
		if err != nil {
			return fmt.Errorf("Error retrieving server for update %s", err)
		}
		if server.Status != cloudscale.ServerStopped {
			updateRequest := &cloudscale.ServerUpdateRequest{
				Status: cloudscale.ServerStopped,
			}
			err := client.Servers.Update(context.Background(), id, updateRequest)
			if err != nil {
				return fmt.Errorf("Error updating server (%s), %s", server.Status, err)
			}

			_, err = waitForServerStatus(d, meta, []string{"changing", "running"}, "status", "stopped")
			if err != nil {
				return fmt.Errorf("Error waiting for server (%s) to change status %s", d.Id(), err)
			}
		}

		updateRequest := &cloudscale.ServerUpdateRequest{Flavor: wantedFlavor}

		err = client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error scaling the Server (%s) status (%s) ", id, err)
		}
		_, err = waitForServerStatus(d, meta, []string{"changing"}, "status", "stopped")

		// Signal that we want to start the server again
		if wantedStatus == "running" {
			needStart = true
		}
		d.SetPartial("flavor_slug")
	}

	if d.HasChange("status") || needStart {
		updateRequest := &cloudscale.ServerUpdateRequest{
			Status: wantedStatus,
		}
		err := client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error changing status (%s) (%s) ", id, err)
		}

		if wantedStatus == "rebooted" {
			return fmt.Errorf("Status (%s) not supported", wantedStatus)
		}

		if wantedStatus == "stopped" {
			_, err = waitForServerStatus(d, meta, []string{"changing", "running"}, "status", "stopped")
		} else {
			_, err = waitForServerStatus(d, meta, []string{"changing", "stopped"}, "status", "running")
		}

		if err != nil {
			return fmt.Errorf("Error waiting for server (%s) to change status %s", d.Id(), err)
		}
		d.SetPartial("status")

	}

	if d.HasChange("name") {
		updateRequest := &cloudscale.ServerUpdateRequest{Name: wantedName}
		err := client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error renaming the Server (%s) status (%s) ", id, err)
		}
		d.SetPartial("name")
	}

	if d.HasChange("interfaces") {
		interfaceRequests := createInterfaceOptions(d)
		updateRequest := &cloudscale.ServerUpdateRequest{Interfaces: &interfaceRequests}
		err := client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error changing the Server (%s) interfaces (%s) ", id, err)
		}
		d.SetPartial("interfaces")
	}

	return resourceServerRead(d, meta)
}

func resourceServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Server: %s", d.Id())
	err := client.Servers.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting server")
	}

	return nil
}

func waitForServerStatus(d *schema.ResourceData, meta interface{}, pending []string, attribute, target string) (interface{}, error) {
	log.Printf(
		"[INFO] Waiting for server (%s) to have %s of %s",
		d.Id(), attribute, target)

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     []string{target},
		Refresh:    newServerRefreshFunc(d, attribute, meta),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func newServerRefreshFunc(d *schema.ResourceData, attribute string, meta interface{}) resource.StateRefreshFunc {
	client := meta.(*cloudscale.Client)
	return func() (interface{}, string, error) {
		id := d.Id()

		err := resourceServerRead(d, meta)
		if err != nil {
			return nil, "", err
		}

		if attr, ok := d.GetOk(attribute); ok {
			server, err := client.Servers.Get(context.Background(), id)
			if err != nil {
				return nil, "", fmt.Errorf("Error retrieving server (refresh) %s", err)
			}

			if server.Status == "errored" {
				return nil, "", fmt.Errorf("Server status %s, abort", server.Status)
			}

			if sshKeys := len(server.SSHHostKeys); sshKeys <= 0 {
				return nil, "", nil
			}

			return server, attr.(string), nil
		}
		return nil, "", nil
	}
}

func findIPv6AddrByType(s *cloudscale.Server, addrType string) string {
	for _, interf := range s.Interfaces {
		if interf.Type == addrType {
			for _, addr := range interf.Addresses {
				if addr.Version == 6 {
					return addr.Address
				}
			}
		}
	}
	return ""
}

func findIPv4AddrByType(s *cloudscale.Server, addrType string) string {
	for _, interf := range s.Interfaces {
		if interf.Type == addrType {
			for _, addr := range interf.Addresses {
				if addr.Version == 4 {
					return addr.Address
				}
			}
		}
	}
	return ""
}
