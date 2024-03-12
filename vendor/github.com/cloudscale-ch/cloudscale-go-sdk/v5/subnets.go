package cloudscale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

const subnetBasePath = "v1/subnets"

var UseCloudscaleDefaults = []string{"CLOUDSCALE_DEFAULTS"}

type Subnet struct {
	TaggedResource
	// Just use omitempty everywhere. This makes it easy to use restful. Errors
	// will be coming from the API if something is disabled.
	HREF           string      `json:"href,omitempty"`
	UUID           string      `json:"uuid,omitempty"`
	CIDR           string      `json:"cidr,omitempty"`
	Network        NetworkStub `json:"network,omitempty"`
	GatewayAddress string      `json:"gateway_address,omitempty"`
	DNSServers     []string    `json:"dns_servers,omitempty"`
}

type SubnetStub struct {
	HREF string `json:"href,omitempty"`
	CIDR string `json:"cidr,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

type SubnetCreateRequest struct {
	TaggedResourceRequest
	CIDR           string    `json:"cidr,omitempty"`
	Network        string    `json:"network,omitempty"`
	GatewayAddress string    `json:"gateway_address,omitempty"`
	DNSServers     *[]string `json:"dns_servers,omitempty"`
}

type SubnetUpdateRequest struct {
	TaggedResourceRequest
	GatewayAddress string    `json:"gateway_address,omitempty"`
	DNSServers     *[]string `json:"dns_servers"`
}

func (request SubnetUpdateRequest) MarshalJSON() ([]byte, error) {
	type Alias SubnetUpdateRequest // Create an alias to avoid recursion

	if request.DNSServers == nil {
		return json.Marshal(&struct {
			Alias
			DNSServers []string `json:"dns_servers,omitempty"`
		}{
			Alias: (Alias)(request),
		})
	}

	if reflect.DeepEqual(*request.DNSServers, UseCloudscaleDefaults) {
		return json.Marshal(&struct {
			Alias
			DNSServers []string `json:"dns_servers"` // important: no omitempty
		}{
			Alias:      (Alias)(request),
			DNSServers: nil,
		})
	}

	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(request),
	})
}

func (request SubnetCreateRequest) MarshalJSON() ([]byte, error) {
	type Alias SubnetCreateRequest // Create an alias to avoid recursion

	if request.DNSServers == nil {
		return json.Marshal(&struct {
			Alias
			DNSServers []string `json:"dns_servers,omitempty"`
		}{
			Alias: (Alias)(request),
		})
	}

	if reflect.DeepEqual(*request.DNSServers, UseCloudscaleDefaults) {
		return json.Marshal(&struct {
			Alias
			DNSServers []string `json:"dns_servers"` // important: no omitempty
		}{
			Alias:      (Alias)(request),
			DNSServers: nil,
		})
	}

	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(request),
	})
}

type SubnetService interface {
	Create(ctx context.Context, createRequest *SubnetCreateRequest) (*Subnet, error)
	Get(ctx context.Context, subnetID string) (*Subnet, error)
	List(ctx context.Context, modifiers ...ListRequestModifier) ([]Subnet, error)
	Update(ctx context.Context, subnetID string, updateRequest *SubnetUpdateRequest) error
	Delete(ctx context.Context, subnetID string) error
}

type SubnetServiceOperations struct {
	client *Client
}

func (s SubnetServiceOperations) Create(ctx context.Context, createRequest *SubnetCreateRequest) (*Subnet, error) {
	path := subnetBasePath

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, createRequest)
	if err != nil {
		return nil, err
	}

	subnet := new(Subnet)

	err = s.client.Do(ctx, req, subnet)
	if err != nil {
		return nil, err
	}

	return subnet, nil
}

func (s SubnetServiceOperations) Update(ctx context.Context, subnetID string, updateRequest *SubnetUpdateRequest) error {
	path := fmt.Sprintf("%s/%s", subnetBasePath, subnetID)

	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, updateRequest)
	if err != nil {
		return err
	}

	err = s.client.Do(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s SubnetServiceOperations) Get(ctx context.Context, subnetID string) (*Subnet, error) {
	path := fmt.Sprintf("%s/%s", subnetBasePath, subnetID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	subnet := new(Subnet)
	err = s.client.Do(ctx, req, subnet)
	if err != nil {
		return nil, err
	}

	return subnet, nil
}

func (s SubnetServiceOperations) Delete(ctx context.Context, subnetID string) error {
	path := fmt.Sprintf("%s/%s", subnetBasePath, subnetID)

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return s.client.Do(ctx, req, nil)
}

func (s SubnetServiceOperations) List(ctx context.Context, modifiers ...ListRequestModifier) ([]Subnet, error) {
	path := subnetBasePath

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	for _, modifier := range modifiers {
		modifier(req)
	}

	subnets := []Subnet{}
	err = s.client.Do(ctx, req, &subnets)
	if err != nil {
		return nil, err
	}

	return subnets, nil
}
