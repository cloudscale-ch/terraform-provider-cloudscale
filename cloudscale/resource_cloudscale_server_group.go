package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const serverGroupHumanName = "server group"

var (
	resourceCloudscaleServerGroupCreate = getCreateOperation(createServerGroup, nil)
	resourceCloudscaleServerGroupRead   = getReadOperation(serverGroupHumanName, getGenericResourceIdentifierFromSchema, readServerGroup, gatherServerGroupResourceData)
	resourceCloudscaleServerGroupUpdate = getUpdateOperation(serverGroupHumanName, getGenericResourceIdentifierFromSchema, updateServerGroup, resourceCloudscaleServerGroupRead, gatherServerGroupUpdateRequest, nil)
	resourceCloudscaleServerGroupDelete = getDeleteOperation(serverGroupHumanName, getGenericResourceIdentifierFromSchema, deleteServerGroup, nil)
)

func resourceCloudscaleServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleServerGroupCreate,
		ReadContext:   resourceCloudscaleServerGroupRead,
		UpdateContext: resourceCloudscaleServerGroupUpdate,
		DeleteContext: resourceCloudscaleServerGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getServerGroupSchema(RESOURCE),
	}
}

func getServerGroupSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"type": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
			ForceNew: true,
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
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

func createServerGroup(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ServerGroupRequest{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
	}
	opts.Tags = TagsFromState(d)

	log.Printf("[DEBUG] ServerGroup create configuration: %#v", opts)

	serverGroup, err := client.ServerGroups.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating server group: %s", err))
	}

	d.SetId(serverGroup.UUID)

	log.Printf("[INFO] ServerGroup ID %s", d.Id())

	return resourceCloudscaleServerGroupRead(ctx, d, meta)
}

func gatherServerGroupResourceData(serverGroup *cloudscale.ServerGroup) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = serverGroup.UUID
	m["href"] = serverGroup.HREF
	m["name"] = serverGroup.Name
	m["type"] = serverGroup.Type
	m["zone_slug"] = serverGroup.Zone.Slug
	m["tags"] = TagsToState(serverGroup.Tags)
	return m
}

func readServerGroup(ctx context.Context, rId GenericResourceIdentifier, meta any) (*cloudscale.ServerGroup, error) {
	client := meta.(*cloudscale.Client)
	return client.ServerGroups.Get(ctx, rId.Id)
}

func updateServerGroup(ctx context.Context, rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.ServerGroupRequest) error {
	client := meta.(*cloudscale.Client)
	return client.ServerGroups.Update(ctx, rId.Id, updateRequest)
}

func gatherServerGroupUpdateRequest(d *schema.ResourceData) []*cloudscale.ServerGroupRequest {
	requests := make([]*cloudscale.ServerGroupRequest, 0)

	for _, attribute := range []string{"name", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.ServerGroupRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = TagsFromState(d)
			}
		}
	}
	return requests
}

func deleteServerGroup(ctx context.Context, rId GenericResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	return client.ServerGroups.Delete(ctx, rId.Id)
}
