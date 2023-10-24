package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

const loadBalancerHumanName = "load balancer"

var (
	resourceCloudscaleLoadBalancerRead   = getReadOperation(loadBalancerHumanName, getGenericResourceIdentifierFromSchema, readLoadBalancer, gatherLoadBalancerResourceData)
	resourceCloudscaleLoadBalancerUpdate = getUpdateOperation(loadBalancerHumanName, getGenericResourceIdentifierFromSchema, updateLoadBalancer, resourceCloudscaleLoadBalancerRead, gatherLoadBalancerUpdateRequest)
	resourceCloudscaleLoadBalancerDelete = getDeleteOperation(loadBalancerHumanName, deleteLoadBalancer)
)

func resourceCloudscaleLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleLoadBalancerCreate,
		Read:   resourceCloudscaleLoadBalancerRead,
		Update: resourceCloudscaleLoadBalancerUpdate,
		Delete: resourceCloudscaleLoadBalancerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getLoadBalancerSchema(RESOURCE),
	}
}

func getLoadBalancerSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"flavor_slug": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"vip_addresses": {
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
			Optional: t.isResource(),
			ForceNew: true,
			Computed: true,
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
			ForceNew: true,
		},
		"tags": &TagsSchema,
	}
	if t.isDataSource() {
		m["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	return m
}

func resourceCloudscaleLoadBalancerCreate(d *schema.ResourceData, meta any) error {
	timeout := d.Timeout(schema.TimeoutCreate)
	startTime := time.Now()

	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerRequest{
		ZonalResourceRequest: cloudscale.ZonalResourceRequest{
			Zone: d.Get("zone_slug").(string),
		},
		Name:   d.Get("name").(string),
		Flavor: d.Get("flavor_slug").(string),
	}

	vipAddressCount := d.Get("vip_addresses.#").(int)
	if vipAddressCount > 0 {
		vipAddressRequests := createVipAddressOptions(d)
		opts.VIPAddresses = &vipAddressRequests
	}

	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] LoadBalancer create configuration: %#v", opts)

	loadbalancer, err := client.LoadBalancers.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating LoadBalancer: %s", err)
	}

	d.SetId(loadbalancer.UUID)

	log.Printf("[INFO] LoadBalancer ID: %s", d.Id())

	remainingTime := timeout - time.Since(startTime)
	_, err = waitForStatus([]string{"changing"}, "running", &remainingTime, newLoadBalancerRefreshFunc(d, "status", meta))
	if err != nil {
		return fmt.Errorf("error waiting for load balancer (%s) to become ready: %s", d.Id(), err)
	}

	err = resourceCloudscaleLoadBalancerRead(d, meta)
	if err != nil {
		return fmt.Errorf("error reading the load balancer (%s): %s", d.Id(), err)
	}
	return nil
}

func newLoadBalancerRefreshFunc(d *schema.ResourceData, attribute string, meta any) resource.StateRefreshFunc {
	client := meta.(*cloudscale.Client)
	return func() (any, string, error) {
		id := d.Id()

		// read the latest data into d
		err := resourceCloudscaleLoadBalancerRead(d, meta)
		if err != nil {
			return nil, "", err
		}
		// get the instance
		loadBalancer, err := client.LoadBalancers.Get(context.Background(), id)
		if err != nil {
			return nil, "", fmt.Errorf("error retrieving load balancer(%s) (refresh) %s", id, err)
		}

		attr, ok := d.GetOk(attribute)
		if !ok {
			return nil, "", nil
		}

		// return attr
		return loadBalancer, attr.(string), nil
	}
}

func createVipAddressOptions(d *schema.ResourceData) []cloudscale.VIPAddressRequest {
	vipAddressCount := d.Get("vip_addresses.#").(int)
	result := make([]cloudscale.VIPAddressRequest, vipAddressCount)
	for i := 0; i < vipAddressCount; i++ {
		prefix := fmt.Sprintf("vip_addresses.%d", i)
		result[i] = cloudscale.VIPAddressRequest{
			Address: d.Get(prefix + ".address").(string),
			Subnet:  d.Get(prefix + ".subnet_uuid").(string),
		}
	}
	return result
}

func gatherLoadBalancerResourceData(loadbalancer *cloudscale.LoadBalancer) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = loadbalancer.UUID
	m["href"] = loadbalancer.HREF
	m["name"] = loadbalancer.Name
	m["flavor_slug"] = loadbalancer.Flavor.Slug
	m["zone_slug"] = loadbalancer.Zone.Slug
	m["status"] = loadbalancer.Status
	m["tags"] = loadbalancer.Tags

	if addrss := len(loadbalancer.VIPAddresses); addrss > 0 {
		vipAddressesMap := make([]map[string]any, 0, addrss)
		for _, vip := range loadbalancer.VIPAddresses {

			vipMap := make(map[string]any)

			vipMap["version"] = vip.Version
			vipMap["address"] = vip.Address
			vipMap["subnet_uuid"] = vip.Subnet.UUID
			vipMap["subnet_cidr"] = vip.Subnet.CIDR
			vipMap["subnet_href"] = vip.Subnet.HREF

			vipAddressesMap = append(vipAddressesMap, vipMap)
		}
		m["vip_addresses"] = vipAddressesMap
	}

	return m
}

func readLoadBalancer(rId GenericResourceIdentifier, meta any) (*cloudscale.LoadBalancer, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancers.Get(context.Background(), rId.Id)
}

func updateLoadBalancer(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.LoadBalancerRequest) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancers.Update(context.Background(), rId.Id, updateRequest)
}

func gatherLoadBalancerUpdateRequest(d *schema.ResourceData) []*cloudscale.LoadBalancerRequest {
	requests := make([]*cloudscale.LoadBalancerRequest, 0)

	for _, attribute := range []string{"name", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.LoadBalancerRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteLoadBalancer(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.LoadBalancers.Delete(context.Background(), id)
}
