package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

const poolHumanName = "load balancer pool"

var (
	resourceCloudscaleLoadBalancerPoolRead   = getReadOperation(poolHumanName, getGenericResourceIdentifierFromSchema, readLoadBalancerPool, gatherLoadBalancerPoolResourceData)
	resourceCloudscaleLoadBalancerPoolUpdate = getUpdateOperation(poolHumanName, getGenericResourceIdentifierFromSchema, updateLoadBalancerPool, resourceCloudscaleLoadBalancerPoolRead, gatherLoadBalancerPoolUpdateRequest)
	resourceCloudscaleLoadBalancerPoolDelete = getDeleteOperation(poolHumanName, deleteLoadBalancerPool)
)

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
			ForceNew: true,
		},
		"protocol": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
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

func updateLoadBalancerPool(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.LoadBalancerPoolRequest) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPools.Update(context.Background(), rId.Id, updateRequest)
}

func gatherLoadBalancerPoolUpdateRequest(d *schema.ResourceData) []*cloudscale.LoadBalancerPoolRequest {
	requests := make([]*cloudscale.LoadBalancerPoolRequest, 0)

	for _, attribute := range []string{"name", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.LoadBalancerPoolRequest{}
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

func deleteLoadBalancerPool(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.LoadBalancerPools.Delete(context.Background(), id)
}
