package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const interfaceHumanName = "interface"

func routerLockKey(routerUUID string) string {
	return fmt.Sprintf("cloudscale/router/%s", routerUUID)
}

// interfaceLockKey serializes interface operations on their parent router.
var interfaceLockKey = uuidLockKey("router_uuid", routerLockKey)

var (
	resourceCloudscaleInterfaceCreate = getCreateOperation(createInterface, interfaceLockKey)
	resourceCloudscaleInterfaceRead   = getReadOperation(interfaceHumanName, getInterfaceResourceIdentifierFromSchema, readInterface, gatherInterfaceResourceData)
	resourceCloudscaleInterfaceDelete = getDeleteOperation(interfaceHumanName, getInterfaceResourceIdentifierFromSchema, deleteInterface, interfaceLockKey)
)

// resourceCloudscaleInterface exposes the generic "cloudscale_interface" resource.
// Today an interface can only be attached to a router, so this resource is backed
// by the router API: the required router_uuid field, the router-based lock key, and
// the router-scan read below are the router coupling points to revisit once a generic
// interfaces endpoint exists.
func resourceCloudscaleInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleInterfaceCreate,
		ReadContext:   resourceCloudscaleInterfaceRead,
		DeleteContext: resourceCloudscaleInterfaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(
				ctx context.Context,
				d *schema.ResourceData,
				m any,
			) ([]*schema.ResourceData, error) {
				routerID, id, err := splitImportID(d.Id(), "router_uuid", "interface_uuid")
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
		Schema: getInterfaceSchema(),
	}
}

type InterfaceResourceIdentifier struct {
	Id       string
	RouterID string
}

func getInterfaceResourceIdentifierFromSchema(d *schema.ResourceData) InterfaceResourceIdentifier {
	return InterfaceResourceIdentifier{
		Id:       d.Id(),
		RouterID: d.Get("router_uuid").(string),
	}
}

func getInterfaceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"router_uuid": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"network_uuid": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
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
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: addressSchema(true),
			},
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"mac_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func createInterface(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	routerUUID := d.Get("router_uuid").(string)

	opts := cloudscale.CreateInterfaceRequest{
		Network: d.Get("network_uuid").(string),
	}

	for _, address := range d.Get("addresses").([]any) {
		a := address.(map[string]any)
		req := cloudscale.CreateAddressRequest{
			Subnet:  a["subnet_uuid"].(string),
			Address: a["address"].(string),
		}
		opts.Addresses = append(opts.Addresses, req)
	}

	log.Printf("[DEBUG] Interface create configuration: %#v", opts)

	// Router-backed create: interfaces are created through their parent router.
	iface, err := client.Routers.CreateInterface(ctx, routerUUID, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating interface: %s", err))
	}

	d.SetId(iface.UUID)

	log.Printf("[INFO] Interface ID %s", d.Id())
	// Populate state from the create response rather than re-reading through the
	// parent router. There is no per-interface GET endpoint, so the read scans the
	// router's interface list; an immediately-following Get can be eventually
	// consistent and not yet list the new interface, which would make the read
	// return a 404 and clear the freshly-created ID.
	fillResourceData(d, gatherInterfaceResourceData(iface))
	return nil
}

func gatherInterfaceResourceData(iface *cloudscale.RouterInterface) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = iface.UUID
	m["network_uuid"] = iface.Network.UUID
	m["network_name"] = iface.Network.Name
	m["network_href"] = iface.Network.HREF
	m["type"] = iface.Type
	m["mac_address"] = iface.MACAddress

	m["addresses"] = gatherAddresses(iface.Addresses)
	return m
}

func readInterface(ctx context.Context, rId InterfaceResourceIdentifier, meta any) (*cloudscale.RouterInterface, error) {
	client := meta.(*cloudscale.Client)

	// Router-backed read: there is no per-interface GET endpoint, so scan the
	// parent router's interface list for the interface we manage.
	router, err := client.Routers.Get(ctx, rId.RouterID)
	if err != nil {
		return nil, err
	}

	for i := range router.Interfaces {
		if router.Interfaces[i].UUID == rId.Id {
			return &router.Interfaces[i], nil
		}
	}

	// The interface no longer exists on the router; signal a 404 so it is
	// removed from state via CheckDeleted.
	return nil, &cloudscale.ErrorResponse{StatusCode: http.StatusNotFound}
}

func deleteInterface(ctx context.Context, rId InterfaceResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	return client.Routers.DeleteInterface(ctx, rId.RouterID, rId.Id)
}
