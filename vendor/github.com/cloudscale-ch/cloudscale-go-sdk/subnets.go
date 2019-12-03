package cloudscale

import (
	"context"
	"fmt"
	"net/http"
)

const subnetBasePath = "v1/subnets"

type Subnet struct {
	ZonalResource
	// Just use omitempty everywhere. This makes it easy to use restful. Errors
	// will be coming from the API if something is disabled.
	HREF    string      `json:"href,omitempty"`
	UUID    string      `json:"uuid,omitempty"`
	CIDR    string      `json:"cidr,omitempty"`
	Network NetworkStub `json:"network,omitempty"`
}

type SubnetStub struct {
	HREF string `json:"href,omitempty"`
	CIDR string `json:"cidr,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

type SubnetService interface {
	Get(ctx context.Context, subnetID string) (*Subnet, error)
	List(ctx context.Context) ([]Subnet, error)
}

type SubnetServiceOperations struct {
	client *Client
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

func (s SubnetServiceOperations) List(ctx context.Context) ([]Subnet, error) {
	path := subnetBasePath
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	subnets := []Subnet{}
	err = s.client.Do(ctx, req, &subnets)
	if err != nil {
		return nil, err
	}

	return subnets, nil
}
