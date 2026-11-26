package cloudscale

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const routerExtraRouteHumanName = "router extra route"

// extraRouteLockKey serializes extra-route operations on their parent router.
var extraRouteLockKey = uuidLockKey("router_uuid", routerLockKey)

var (
	resourceCloudscaleRouterExtraRouteCreate = getCreateOperation(createRouterExtraRoute, extraRouteLockKey)
	resourceCloudscaleRouterExtraRouteRead   = getReadOperation(routerExtraRouteHumanName, getRouterExtraRouteResourceIdentifierFromSchema, readRouterExtraRoute, gatherRouterExtraRouteResourceData)
	resourceCloudscaleRouterExtraRouteDelete = getDeleteOperation(routerExtraRouteHumanName, getRouterExtraRouteResourceIdentifierFromSchema, deleteRouterExtraRoute, extraRouteLockKey)
)

func resourceCloudscaleRouterExtraRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleRouterExtraRouteCreate,
		ReadContext:   resourceCloudscaleRouterExtraRouteRead,
		DeleteContext: resourceCloudscaleRouterExtraRouteDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(
				ctx context.Context,
				d *schema.ResourceData,
				m any,
			) ([]*schema.ResourceData, error) {
				routerID, id, err := splitImportID(d.Id(), "router_uuid", "extra_route_uuid")
				if err != nil {
					return nil, err
				}
				err = d.Set("router_uuid", routerID)
				if err != nil {
					return nil, err
				}
				d.SetId(id)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: getRouterExtraRouteSchema(),
	}
}

// RouterExtraRouteResourceIdentifier identifies a single extra route on a router.
type RouterExtraRouteResourceIdentifier struct {
	Id       string
	RouterId string
}

func getRouterExtraRouteResourceIdentifierFromSchema(d *schema.ResourceData) RouterExtraRouteResourceIdentifier {
	return RouterExtraRouteResourceIdentifier{
		Id:       d.Id(),
		RouterId: d.Get("router_uuid").(string),
	}
}

func getRouterExtraRouteSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"router_uuid": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"destination": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"next_hop": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func createRouterExtraRoute(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	routerUUID := d.Get("router_uuid").(string)

	req := cloudscale.AddRouterExtraRouteRequest{
		Destination: d.Get("destination").(string),
		NextHop:     d.Get("next_hop").(string),
	}

	extraRoute, err := client.Routers.AddRouterExtraRoute(ctx, routerUUID, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating router extra route: %s", err))
	}

	d.SetId(extraRoute.UUID)

	fillResourceData(d, gatherRouterExtraRouteResourceData(extraRoute))
	return nil
}

func gatherRouterExtraRouteResourceData(extraRoute *cloudscale.RouterExtraRoute) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = extraRoute.UUID
	m["destination"] = extraRoute.Destination
	m["next_hop"] = extraRoute.NextHop
	m["href"] = extraRoute.HREF
	return m
}

func readRouterExtraRoute(ctx context.Context, rId RouterExtraRouteResourceIdentifier, meta any) (*cloudscale.RouterExtraRoute, error) {
	client := meta.(*cloudscale.Client)

	// We get the list of routes from the parent router
	router, err := client.Routers.Get(ctx, rId.RouterId)
	if err != nil {
		return nil, err
	}

	for i := range router.ExtraRoutes {
		if router.ExtraRoutes[i].UUID == rId.Id {
			return &router.ExtraRoutes[i], nil
		}
	}

	return nil, &cloudscale.ErrorResponse{StatusCode: http.StatusNotFound}
}

func deleteRouterExtraRoute(ctx context.Context, rId RouterExtraRouteResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	return client.Routers.RemoveRouterExtraRoute(ctx, rId.RouterId, rId.Id)
}
