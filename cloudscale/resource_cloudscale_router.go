package cloudscale

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const routerHumanName = "router"

var (
	resourceCloudscaleRouterCreate = getCreateOperation(createRouter, nil)
	resourceCloudscaleRouterRead   = getReadOperation(routerHumanName, getGenericResourceIdentifierFromSchema, readRouter, gatherRouterResourceData)
	// update not implemented yet
	// resourceCloudscaleRouterUpdate = getUpdateOperation(routerHumanName, getGenericResourceIdentifierFromSchema, updateRouter, resourceCloudscaleRouterRead, gatherRouterUpdateRequest, nil)
	resourceCloudscaleRouterDelete = getDeleteOperation(routerHumanName, getGenericResourceIdentifierFromSchema, deleteRouter, nil)
)

func resourceCloudscaleRouter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleRouterCreate,
		ReadContext:   resourceCloudscaleRouterRead,
		// UpdateContext: resourceCloudscaleRouterUpdate,
		DeleteContext: resourceCloudscaleRouterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getRouterSchema(RESOURCE),
	}
}

// addressSchema returns the Elem schema for a list of router IP addresses.
// writable makes subnet_uuid/address user-settable (router-interface input);
// otherwise every field is Computed (router read-only representation).
func addressSchema(writable bool) map[string]*schema.Schema {
	subnetUUID := &schema.Schema{
		Type:     schema.TypeString,
		Required: writable,
		Computed: !writable,
		ForceNew: true,
	}
	address := &schema.Schema{
		Type:     schema.TypeString,
		Optional: writable,
		Computed: true,
		ForceNew: true,
	}
	return map[string]*schema.Schema{
		"address":     address,
		"subnet_uuid": subnetUUID,
		"subnet_href": {Type: schema.TypeString, Computed: true},
		"subnet_cidr": {Type: schema.TypeString, Computed: true},
		"version":     {Type: schema.TypeInt, Computed: true},
		"reverse_ptr": {Type: schema.TypeString, Computed: true},
	}
}

func gatherAddresses(in []cloudscale.IPAddress) []map[string]any {
	out := make([]map[string]any, len(in))
	for i, addr := range in {
		out[i] = map[string]any{
			"address":     addr.Address,
			"subnet_uuid": addr.Subnet.UUID,
			"subnet_href": addr.Subnet.HREF,
			"subnet_cidr": addr.Subnet.CIDR,
			"version":     addr.Version,
			"reverse_ptr": addr.ReversePTR,
		}
	}
	return out
}

func getRouterSchema(t SchemaType) map[string]*schema.Schema {
	// FIXME: update not implemented yet, tags needs ForceNew until this is implemented
	tagsSchema := TagsSchema
	tagsSchema.ForceNew = true

	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
			Computed: t.isDataSource(),
			ForceNew: true, // update not implemented yet
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
			Computed: t.isDataSource(),
			ForceNew: true,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"tags": &tagsSchema,
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"internet_gateway": {
			Type:     schema.TypeBool,
			Optional: t.isResource(),
			Computed: t.isDataSource(),
			ForceNew: true, // update not implemented yet
		},
		"internet_gateway_addresses": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: addressSchema(false),
			},
			Computed: true,
		},
		"interfaces": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"uuid": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"network_uuid": {
						Type:     schema.TypeString,
						Computed: true,
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
							Schema: addressSchema(false),
						},
						Computed: true,
					},
					"type": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"mac_address": {
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
	}
	return m
}

func createRouter(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.RouterCreateRequest{
		Name: d.Get("name").(string),
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
	}
	if attr, ok := d.GetOk("internet_gateway"); ok {
		opts.InternetGateway = attr.(bool)
	}
	opts.Tags = TagsFromState(d)

	log.Printf("[DEBUG] Router create configuration: %#v", opts)

	router, err := client.Routers.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating router: %s", err))
	}

	d.SetId(router.UUID)

	log.Printf("[INFO] Router ID %s", d.Id())
	return resourceCloudscaleRouterRead(ctx, d, meta)
}

func gatherRouterResourceData(router *cloudscale.Router) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = router.UUID
	m["name"] = router.Name
	m["zone_slug"] = router.Zone.Slug
	m["href"] = router.HREF
	m["tags"] = TagsToState(router.Tags)
	m["status"] = router.Status
	m["internet_gateway"] = router.InternetGateway

	m["internet_gateway_addresses"] = gatherAddresses(router.InternetGatewayAddresses)

	RouterInterface := make([]map[string]any, len(router.Interfaces))
	for i := range router.Interfaces {
		// Reuse the interface gather; it keys the UUID as "id", but the
		// router's interfaces schema exposes it as "uuid".
		g := gatherInterfaceResourceData(&router.Interfaces[i])
		g["uuid"] = g["id"]
		delete(g, "id")
		RouterInterface[i] = g
	}
	m["interfaces"] = RouterInterface
	return m
}

func readRouter(ctx context.Context, rId GenericResourceIdentifier, meta any) (*cloudscale.Router, error) {
	client := meta.(*cloudscale.Client)
	return client.Routers.Get(ctx, rId.Id)
}

func updateRouter(ctx context.Context, rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.RouterUpdateRequest) error {
	// client := meta.(*cloudscale.Client)
	// return client.Routers.Update(ctx, rId.Id, updateRequest)
	return errors.New("not implemented")
}

func gatherRouterUpdateRequest(d *schema.ResourceData) []*cloudscale.RouterUpdateRequest {
	requests := make([]*cloudscale.RouterUpdateRequest, 0)
	// FIXME: not implemented yet
	//
	// for _, attribute := range []string{"name", "internet_gateway", "tags"} {
	// 	if d.HasChange(attribute) {
	// 		log.Printf("[INFO] Attribute %s changed", attribute)
	// 		opts := &cloudscale.RouterUpdateRequest{}
	// 		requests = append(requests, opts)
	//
	// 		if attribute == "name" {
	// 			opts.Name = d.Get(attribute).(string)
	// 		} else if attribute == "internet_gateway" {
	// 			opts.InternetGateway = d.Get(attribute).(bool)
	// 		} else if attribute == "tags" {
	// 			opts.Tags = TagsFromState(d)
	// 		}
	// 	}
	// }
	return requests
}

func deleteRouter(ctx context.Context, rId GenericResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	if err := client.Routers.Delete(ctx, rId.Id); err != nil {
		return err
	}
	err := waitForDeleted(ctx, func() (bool, error) {
		router, err := client.Routers.Get(ctx, rId.Id)
		if err != nil {
			if cerr, ok := errors.AsType[*cloudscale.ErrorResponse](err); ok && cerr.StatusCode == http.StatusNotFound {
				return false, nil
			}
			// Tolerate transient (non-404) errors: keep polling rather than
			// failing if there's a transient error. The ctx deadline ensures it's not doing requests forever.
			tflog.Warn(ctx, "transient error polling router during delete, retrying", map[string]any{"id": rId.Id, "error": err.Error()})
			return true, nil
		}
		tflog.Info(ctx, "router still exists", map[string]any{"id": rId.Id, "status": router.Status})
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("error waiting for router (%s) to be deleted: %s", rId.Id, err)
	}
	return nil
}
