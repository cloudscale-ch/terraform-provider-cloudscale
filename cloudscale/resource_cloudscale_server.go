package cloudscale

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudscale-ch/cloudscale"
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
			Type:     schema.TypeList,
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
		"anti_affinity_with": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		},
		"user_data": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"state": {
			Type:     schema.TypeString,
			Optional: true,
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

	sshKeys := d.Get("ssh_keys.#").(int)
	if sshKeys > 0 {
		opts.SSHKeys = make([]string, 0, sshKeys)
		for i := 0; i < sshKeys; i++ {
			key := fmt.Sprintf("ssh_keys.%d", i)
			sshkey := d.Get(key).(string)
			opts.SSHKeys = append(opts.SSHKeys, sshkey)

		}
	}

	if attr, ok := d.GetOk("volume_size_gb"); ok {
		opts.VolumeSizeGB = attr.(int)
	}

	if attr, ok := d.GetOk("bulk_volume_size_gb"); ok {
		opts.BulkVolumeSizeGB = attr.(int)
	}

	if attr, ok := d.GetOk("use_public_network"); ok {
		val := attr.(bool)
		opts.UsePublicNetwork = &val
	}

	if attr, ok := d.GetOk("use_private_network"); ok {
		val := attr.(bool)
		opts.UsePrivateNetwork = &val
	}

	if attr, ok := d.GetOk("use_ipv6"); ok {
		val := attr.(bool)
		opts.UseIPV6 = &val
	}

	antiAffinityUUIDs := d.Get("anti_affinity_with.#").(int)
	if antiAffinityUUIDs > 0 {
		opts.AntiAffinityWith = make([]string, 0, antiAffinityUUIDs)

		for i := 0; i < antiAffinityUUIDs; i++ {
			key := fmt.Sprintf("anti_affinity_with.%d", i)
			antiAffinityUUID := d.Get(key).(string)
			opts.AntiAffinityWith = append(opts.AntiAffinityWith, antiAffinityUUID)

		}
	}

	if attr, ok := d.GetOk("user_data"); ok {
		opts.UserData = attr.(string)
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
		d.Set("volumes", volumesMaps)
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
				i["prefix_length"] = addr.PrefixLenght
				i["gateway"] = addr.Gateway
				i["reverse_ptr"] = addr.ReversePtr

				addrssMap = append(addrssMap, i)
			}

			intMap["type"] = intr.Type
			intMap["addresses"] = addrssMap

			intsMap = append(intsMap, intMap)
		}
		d.Set("interfaces", intsMap)
	}

	if publicIPv4 := findIPv4AddrByType(server, "public"); publicIPv4 != "" {
		d.Set("ipv4_address", publicIPv4)
	}
	if publicIPv6 := findIPv6AddrByType(server, "public"); publicIPv6 != "" {
		d.Set("ipv6_address", publicIPv6)
	}
	if privateIPv4 := findIPv4AddrByType(server, "private"); privateIPv4 != "" {
		d.Set("ipv4_private_address", privateIPv4)
	}
	if privateIPv6 := findIPv4AddrByType(server, "private"); privateIPv6 != "" {
		d.Set("ipv6_private_address", privateIPv6)
	}

	d.Set("ssh_fingerprints", server.SSHFingerprints)

	d.Set("ssh_host_keys", server.SSHHostKeys)

	if antiAffinities := len(server.AntiAfinityWith); antiAffinities > 0 {
		antiAfs := make([]string, 0, antiAffinities)
		for _, antiAf := range server.AntiAfinityWith {
			antiAfs = append(antiAfs, antiAf.UUID)
		}
		d.Set("anti_affinity_with", antiAfs)
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

	if d.HasChange("state") {
		state := d.Get("state").(string)
		err := client.Servers.Update(context.Background(), id, state)
		if err != nil {
			return fmt.Errorf("Error updating the Server (%s) state (%s) ", id, err)
		}

		if state == "stopped" {
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
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting Server: %s", err)
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
		Timeout:    60 * time.Minute,
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
