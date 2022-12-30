package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

const poolHumanName = "load balancer pool"

var resourceCloudscaleLoadBalancerPoolRead = getReadOperation(poolHumanName, getGenericResourceIdentifierFromSchema, readLoadBalancerPool, gatherLoadBalancerPoolResourceData)
var resourceCloudscaleLoadBalancerPoolDelete = getDeleteOperation(poolHumanName, deleteLoadBalancerPool)

func resourceCloudscaleLoadBalancerPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleLoadBalancerPoolCreate,
		Read:   resourceCloudscaleLoadBalancerPoolRead,
		Update: resourceCloudscaleLoadBalancerPoolUpdate,
		Delete: resourceCloudscaleLoadBalancerPoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getLoadBalancerPoolSchema(RESOURCE),
	}
}

func getLoadBalancerPoolSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"load_balancer_uuid": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
			ForceNew: true,
		},
		"load_balancer_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"load_balancer_href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"algorithm": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"protocol": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
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

func resourceCloudscaleLoadBalancerPoolCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerPoolRequest{
		Name:         d.Get("name").(string),
		LoadBalancer: d.Get("load_balancer_uuid").(string),
		Algorithm:    d.Get("algorithm").(string),
		Protocol:     d.Get("protocol").(string),
	}

	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] LoadBalancerPool create configuration: %#v", opts)

	loadBalancerPool, err := client.LoadBalancerPools.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating LoadBalancerPool: %s", err)
	}

	d.SetId(loadBalancerPool.UUID)

	log.Printf("[INFO] LoadBalancerPool ID: %s", d.Id())
	err = resourceCloudscaleLoadBalancerPoolRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the load balancer pool (%s): %s", d.Id(), err)
	}
	return nil
}

func gatherLoadBalancerPoolResourceData(loadbalancerpool *cloudscale.LoadBalancerPool) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = loadbalancerpool.UUID
	m["href"] = loadbalancerpool.HREF
	m["name"] = loadbalancerpool.Name
	m["load_balancer_uuid"] = loadbalancerpool.LoadBalancer.UUID
	m["load_balancer_name"] = loadbalancerpool.LoadBalancer.Name
	m["load_balancer_href"] = loadbalancerpool.LoadBalancer.HREF
	m["algorithm"] = loadbalancerpool.Algorithm
	m["protocol"] = loadbalancerpool.Protocol
	m["tags"] = loadbalancerpool.Tags
	return m
}

func readLoadBalancerPool(rId GenericResourceIdentifier, meta any) (*cloudscale.LoadBalancerPool, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPools.Get(context.Background(), rId.Id)
}

func resourceCloudscaleLoadBalancerPoolUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "algorithm", "protocol", "tags"} {
		if d.HasChange(attribute) {
			opts := &cloudscale.LoadBalancerPoolRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "algorithm" {
				opts.Algorithm = d.Get(attribute).(string)
			} else if attribute == "protocol" {
				opts.Protocol = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
			err := client.LoadBalancerPools.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Load Balancer Pool (%s): %s", id, err)
			}
		}
	}
	return resourceCloudscaleLoadBalancerPoolRead(d, meta)
}

func deleteLoadBalancerPool(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.LoadBalancerPools.Delete(context.Background(), id)
}
