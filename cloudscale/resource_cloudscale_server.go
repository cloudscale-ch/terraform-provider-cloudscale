package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const serverHumanName = "server"

var resourceCloudscaleServerRead = getReadOperation(serverHumanName, getGenericResourceIdentifierFromSchema, readServer, gatherServerResourceData)
var resourceCloudscaleServerDelete = getDeleteOperation(serverHumanName, deleteServer)

func resourceCloudscaleServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleServerCreate,
		Read:   resourceCloudscaleServerRead,
		Update: resourceCloudscaleServerUpdate,
		Delete: resourceCloudscaleServerDelete,

		Schema: getServerSchema(RESOURCE),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceCloudscaleServerImport,
		},
	}
}

func getServerSchema(t SchemaType) map[string]*schema.Schema {
	imageConflictsWith := []string{}
	if t.isResource() {
		imageConflictsWith = append(imageConflictsWith, "image_uuid")
	}
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"flavor_slug": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"image_slug": {
			Type:          schema.TypeString,
			Optional:      t.isResource(),
			ForceNew:      true,
			ConflictsWith: imageConflictsWith,
			Computed:      true,
		},
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
			Optional: t.isResource(),
			Computed: true,
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
			Optional: t.isResource(),
			Computed: true,
		},
		"tags": &TagsSchema,
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
	if t.isDataSource() {
		m["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	} else {
		m["image_uuid"] = &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"image_slug"},
		}
		m["ssh_keys"] = &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		}
		m["password"] = &schema.Schema{
			Type:      schema.TypeString,
			Optional:  true,
			Elem:      &schema.Schema{Type: schema.TypeString},
			ForceNew:  true,
			Sensitive: true,
		}
		m["volume_size_gb"] = &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
		}
		m["bulk_volume_size_gb"] = &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		}
		m["user_data"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		}
		m["use_public_network"] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		}
		m["use_private_network"] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		}
		m["use_ipv6"] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		}
		m["allow_stopping_for_update"] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
		}
		m["skip_waiting_for_ssh_host_keys"] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true,
		}
		m["server_group_ids"] = &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		}
	}
	return m
}

func resourceCloudscaleServerCreate(d *schema.ResourceData, meta any) error {
	timeout := d.Timeout(schema.TimeoutCreate)
	startTime := time.Now()

	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ServerRequest{
		Name:   d.Get("name").(string),
		Flavor: d.Get("flavor_slug").(string),
		Image:  createImageOption(d),
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
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] Server create configuration: %#v", opts)

	server, err := client.Servers.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}

	d.SetId(server.UUID)

	log.Printf("[INFO] Server ID %s", d.Id())

	remainingTime := timeout - time.Since(startTime)
	_, err = waitForStatus([]string{"changing"}, "running", &remainingTime, newServerRefreshFunc(d, "status", meta))
	if err != nil {
		return fmt.Errorf("error waiting for server (%s) to become ready: %s", d.Id(), err)
	}

	remainingTime = timeout - time.Since(startTime)
	err = waitForSSHHostKeys(d, meta, &remainingTime)
	if err != nil {
		return fmt.Errorf("error waiting for SSH host keys (%s) to be available: %s", d.Id(), err)
	}

	if originalStatus == "stopped" {
		updateRequest := &cloudscale.ServerUpdateRequest{
			Status: originalStatus,
		}
		err := client.Servers.Update(context.Background(), server.UUID, updateRequest)
		if err != nil {
			return fmt.Errorf("error stopping the server (%s) status (%s) ", server.UUID, err)
		}

		remainingTime = timeout - time.Since(startTime)
		_, err = waitForStatus([]string{"changing", "running"}, "stopped", &remainingTime, newServerRefreshFunc(d, "status", meta))
		if err != nil {
			return fmt.Errorf("error waiting for server status (%s) (%s) ", server.UUID, err)
		}
	}

	err = resourceCloudscaleServerRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the server (%s): %s", d.Id(), err)
	}
	return nil
}

func createImageOption(d *schema.ResourceData) string {

	if imageName := d.Get("image_slug").(string); imageName != "" {
		return imageName
	}
	return d.Get("image_uuid").(string)
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
		addresses := d.Get(addressKey).([]any)
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

func createAddressesOptions(addresses []any) []cloudscale.AddressRequest {
	result := make([]cloudscale.AddressRequest, len(addresses))
	for i, address := range addresses {
		a := address.(map[string]any)
		if a["subnet_uuid"] != "" {
			result[i].Subnet = a["subnet_uuid"].(string)
		}
		if a["address"] != "" {
			result[i].Address = a["address"].(string)
		}
	}
	return result
}

func gatherServerResourceData(server *cloudscale.Server) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = server.UUID
	m["href"] = server.HREF
	m["name"] = server.Name
	m["flavor_slug"] = server.Flavor.Slug
	m["image_slug"] = server.Image.Slug
	m["zone_slug"] = server.Zone.Slug
	m["status"] = server.Status
	m["tags"] = server.Tags

	if volumes := len(server.Volumes); volumes > 0 {
		volumesMaps := make([]map[string]any, 0, volumes)
		for _, volume := range server.Volumes {
			v := make(map[string]any)
			v["type"] = volume.Type
			v["device_path"] = volume.DevicePath
			v["size_gb"] = volume.SizeGB
			v["uuid"] = volume.UUID
			volumesMaps = append(volumesMaps, v)
		}
		m["volumes"] = volumesMaps
	}
	serverGroupMaps := make([]map[string]any, 0, len(server.ServerGroups))
	for _, serverGroup := range server.ServerGroups {
		g := make(map[string]any)
		g["uuid"] = serverGroup.UUID
		g["name"] = serverGroup.Name
		g["href"] = serverGroup.HREF
		serverGroupMaps = append(serverGroupMaps, g)
	}
	m["server_groups"] = serverGroupMaps

	if addrss := len(server.Interfaces); addrss > 0 {
		intsMap := make([]map[string]any, 0, addrss)
		for _, intr := range server.Interfaces {

			intMap := make(map[string]any)

			intMap["network_href"] = intr.Network.HREF
			intMap["network_name"] = intr.Network.Name
			intMap["network_uuid"] = intr.Network.UUID

			addrssMap := make([]map[string]any, 0, len(intr.Addresses))
			for _, addr := range intr.Addresses {
				i := make(map[string]any)
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
		m["interfaces"] = intsMap
	}

	m["ssh_fingerprints"] = server.SSHFingerprints

	m["ssh_host_keys"] = server.SSHHostKeys

	m["public_ipv4_address"] = findIPv4AddrByType(server, "public")
	m["public_ipv6_address"] = findIPv6AddrByType(server, "public")
	m["private_ipv4_address"] = findIPv4AddrByType(server, "private")
	return m
}

func readServer(rId GenericResourceIdentifier, meta any) (*cloudscale.Server, error) {
	client := meta.(*cloudscale.Client)
	return client.Servers.Get(context.Background(), rId.Id)
}

func resourceCloudscaleServerUpdate(d *schema.ResourceData, meta any) error {
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
	}

	if d.HasChange("flavor_slug") {
		if !d.Get("allow_stopping_for_update").(bool) {
			return fmt.Errorf("Changing the flavor requires stopping the server. " +
				"To acknowledge this, please set allow_stopping_for_update = true in your config.")
		}

		server, err := client.Servers.Get(context.Background(), id)
		if err != nil {
			return fmt.Errorf("Error retrieving server (%s) for update %s", id, err)
		}
		if server.Status != cloudscale.ServerStopped {
			updateRequest := &cloudscale.ServerUpdateRequest{
				Status: cloudscale.ServerStopped,
			}
			err := client.Servers.Update(context.Background(), id, updateRequest)
			if err != nil {
				return fmt.Errorf("Error updating server (%s), %s", server.Status, err)
			}

			_, err = waitForStatus([]string{"changing", "running"}, "stopped", nil, newServerRefreshFunc(d, "status", meta))
			if err != nil {
				return fmt.Errorf("Error waiting for server (%s) to change status %s", d.Id(), err)
			}
		}

		updateRequest := &cloudscale.ServerUpdateRequest{Flavor: wantedFlavor}

		err = client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error scaling the Server (%s) status (%s) ", id, err)
		}
		_, err = waitForStatus([]string{"changing"}, "stopped", nil, newServerRefreshFunc(d, "status", meta))

		// Signal that we want to start the server again
		if wantedStatus == "running" {
			needStart = true
		}
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
			_, err = waitForStatus([]string{"changing", "running"}, "stopped", nil, newServerRefreshFunc(d, "status", meta))
		} else {
			_, err = waitForStatus([]string{"changing", "stopped"}, "running", nil, newServerRefreshFunc(d, "status", meta))
		}

		if err != nil {
			return fmt.Errorf("Error waiting for server (%s) to change status %s", d.Id(), err)
		}
	}

	if d.HasChange("name") {
		updateRequest := &cloudscale.ServerUpdateRequest{Name: wantedName}
		err := client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error renaming the Server (%s) status (%s) ", id, err)
		}
	}

	if d.HasChange("interfaces") {
		interfaceRequests := createInterfaceOptions(d)
		updateRequest := &cloudscale.ServerUpdateRequest{Interfaces: &interfaceRequests}
		err := client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error changing the Server (%s) interfaces (%s) ", id, err)
		}
	}

	if d.HasChange("tags") {
		updateRequest := &cloudscale.ServerUpdateRequest{}
		updateRequest.Tags = CopyTags(d)
		err := client.Servers.Update(context.Background(), id, updateRequest)
		if err != nil {
			return fmt.Errorf("Error tagging the Server (%s) status (%s) ", id, err)
		}
	}

	return resourceCloudscaleServerRead(d, meta)
}

func deleteServer(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.Servers.Delete(context.Background(), id)
}

func newServerRefreshFunc(d *schema.ResourceData, attribute string, meta any) resource.StateRefreshFunc {
	client := meta.(*cloudscale.Client)
	return func() (any, string, error) {
		id := d.Id()

		// read the latest data into d
		err := resourceCloudscaleServerRead(d, meta)
		if err != nil {
			return nil, "", err
		}
		// get the instance
		server, err := client.Servers.Get(context.Background(), id)
		if err != nil {
			return nil, "", fmt.Errorf("error retrieving server (%s) (refresh) %s", id, err)
		}

		attr, ok := d.GetOk(attribute)
		if !ok {
			return nil, "", nil
		}

		// return attr
		return server, attr.(string), nil
	}
}

func waitForSSHHostKeys(d *schema.ResourceData, meta any, timeout *time.Duration) error {
	if d.Get("skip_waiting_for_ssh_host_keys").(bool) {
		log.Printf("[INFO] Not waiting for server (%s) to have host keys available", d.Id())
		return nil
	}
	log.Printf("[INFO] Waiting %s for server (%s) to have host keys available", timeout, d.Id())

	err := resource.Retry(*timeout, func() *resource.RetryError {
		err := resourceCloudscaleServerRead(d, meta)
		if err != nil {
			return &resource.RetryError{
				Err:       err,
				Retryable: false,
			}
		}

		if attr, ok := d.GetOk("ssh_host_keys.#"); ok {
			count := attr.(int)
			if count <= 0 {
				return &resource.RetryError{
					Err:       fmt.Errorf("ssh_host_keys.# is %d", count),
					Retryable: true,
				}
			}
			return nil
		}

		return &resource.RetryError{
			Err:       fmt.Errorf("getting attribute is not ok"),
			Retryable: true,
		}
	})
	if err != nil {
		return err
	}
	return nil
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

func resourceCloudscaleServerImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	// this attribute is irrelevant for existing servers
	d.Set("skip_waiting_for_ssh_host_keys", false)
	return schema.ImportStatePassthroughContext(ctx, d, meta)
}
