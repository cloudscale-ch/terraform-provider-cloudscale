package cloudscale

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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

		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"flavor_slug": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"image_slug": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"ssh_keys": {
			Type:     schema.TypeSet,
			Required: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		},

		// Optional attributes

		"volume_size_gb": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"bulk_volume_size_gb": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"anti_affinity_uuid": {
			Type:     schema.TypeString,
			Optional: true,
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
			Default:  true,
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
			Default:  true,
		},

		// Computed attributes

		"href": &schema.Schema{
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
					"size_gb": {
						Type:     schema.TypeInt,
						Computed: true,
					},
				},
			},
			Computed: true,
		},
		"interfaces": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
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
							},
						},
						Computed: true,
					},
				},
			},
			Computed: true,
		},
		"ssh_fingerprints": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
		},
		"ssh_host_keys": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
		},
		"anti_affinity_with": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Optional: true,
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

	if attr, ok := d.GetOk("volume_size_gb"); ok {
		opts.VolumeSizeGB = attr.(int)
	}

	if attr, ok := d.GetOk("bulk_volume_size_gb"); ok {
		opts.BulkVolumeSizeGB = attr.(int)
	}

	use_public_network := d.Get("use_public_network")
	use_public_network_bool := use_public_network.(bool)
	opts.UsePublicNetwork = &use_public_network_bool

	use_ipv6 := d.Get("use_ipv6")
	use_ipv6_bool := use_ipv6.(bool)
	opts.UseIPV6 = &use_ipv6_bool

	if attr, ok := d.GetOk("use_private_network"); ok {
		val := attr.(bool)
		opts.UsePrivateNetwork = &val
	}

	if attr, ok := d.GetOk("anti_affinity_uuid"); ok {
		opts.AntiAffinityWith = attr.(string)
	}

	if attr, ok := d.GetOk("user_data"); ok {
		opts.UserData = attr.(string)
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
		err := client.Servers.Update(context.Background(), server.UUID, originalStatus)
		if err != nil {
			return fmt.Errorf("Error updating the Server (%s) status (%s) ", server.UUID, err)
		}

		_, err = waitForServerStatus(d, meta, []string{"changing", "running"}, "status", "stopped")
		if err != nil {
			return fmt.Errorf("Error updating the Server (%s) status (%s) ", server.UUID, err)
		}
	}

	return resourceServerRead(d, meta)
}

func resourceServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	id := d.Id()

	server, err := client.Servers.Get(context.Background(), id)
	if err != nil {
		if err.Error() == "detail: Not Found." {
			log.Printf("[WARN] Cloudscale Server (%s) not found", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving server: %s", err)
	}

	d.Set("href", server.HREF)
	d.Set("name", server.Name)
	d.Set("flavor_slug", server.Flavor.Slug)
	d.Set("image_slug", server.Image.Slug)

	if volumes := len(server.Volumes); volumes > 0 {
		volumesMaps := make([]map[string]interface{}, 0, volumes)
		for _, volume := range server.Volumes {
			v := make(map[string]interface{})
			v["type"] = volume.Type
			v["device_path"] = volume.DevicePath
			v["size_gb"] = volume.SizeGB
			volumesMaps = append(volumesMaps, v)
		}
		err = d.Set("volumes", volumesMaps)
		if err != nil {
			log.Printf("[DEBUG] Error setting volumes attribute: %#v, error: %#v", volumesMaps, err)
			return fmt.Errorf("Error setting volumes attribute: %#v, error: %#v", volumesMaps, err)
		}

	}

	d.Set("status", server.Status)

	if addrss := len(server.Interfaces); addrss > 0 {

		intsMap := make([]map[string]interface{}, 0, addrss)
		for _, intr := range server.Interfaces {

			intMap := make(map[string]interface{})
			addrssMap := make([]map[string]interface{}, 0, len(intr.Adresses))
			for _, addr := range intr.Adresses {
				i := make(map[string]interface{})
				i["address"] = addr.Address
				i["version"] = addr.Version
				i["prefix_length"] = addr.PrefixLength
				i["gateway"] = addr.Gateway
				i["reverse_ptr"] = addr.ReversePtr

				addrssMap = append(addrssMap, i)
			}

			intMap["type"] = intr.Type
			intMap["addresses"] = addrssMap

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

	var antiAfs []string
	for _, antiAf := range server.AntiAfinityWith {
		antiAfs = append(antiAfs, antiAf.UUID)
	}
	err = d.Set("anti_affinity_with", antiAfs)
	if err != nil {
		log.Printf("[DEBUG] Error setting anti_affinity_with attribute: %#v, error: %#v", antiAfs, err)
		return fmt.Errorf("Error setting anti_affinity_with attribute: %#v, error: %#v", antiAfs, err)
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

	return nil
}

func resourceServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	if d.HasChange("status") {
		status := d.Get("status").(string)
		err := client.Servers.Update(context.Background(), id, status)
		if err != nil {
			return fmt.Errorf("Error updating the Server (%s) status (%s) ", id, err)
		}

		if status == "rebooted" {
			return fmt.Errorf("Status (%s) not supported", status)
		}

		if status == "stopped" {
			_, err = waitForServerStatus(d, meta, []string{"changing", "running"}, "status", "stopped")
		} else {
			_, err = waitForServerStatus(d, meta, []string{"changing", "stopped"}, "status", "running")
		}

		if err != nil {
			return fmt.Errorf("Error waiting for server (%s) to change status %s", d.Id(), err)
		}

	}

	return resourceServerRead(d, meta)
}

func resourceServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Server: %s", d.Id())
	err := client.Servers.Delete(context.Background(), id)

	if err != nil && strings.Contains(err.Error(), "Not found") {
		log.Printf("[WARN] Cloudscale Server (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting Server: %s", err)
	}

	d.SetId("")

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
				return nil, "", fmt.Errorf("Error retrieving server %s", err)
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
			for _, addr := range interf.Adresses {
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
			for _, addr := range interf.Adresses {
				if addr.Version == 4 {
					return addr.Address
				}
			}
		}
	}
	return ""
}
