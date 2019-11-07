package cloudscale

import (
	"context"
	"fmt"
	"net/http"
)

const serverGroupsBasePath = "v1/server-groups"

type ServerGroup struct {
	ZonalResource
	HREF    string       `json:"href"`
	UUID    string       `json:"uuid"`
	Name    string       `json:"name"`
	Type    string       `json:"type"`
	Servers []ServerStub `json:"servers"`
}

type ServerGroupRequest struct {
	ZonalResourceRequest
	Name    string       `json:"name"`
	Type    string       `json:"type"`
}

type ServerGroupService interface {
	Create(ctx context.Context, createRequest *ServerGroupRequest) (*ServerGroup, error)
	Get(ctx context.Context, serverGroupID string) (*ServerGroup, error)
	Delete(ctx context.Context, serverGroupID string) error
	List(ctx context.Context) ([]ServerGroup, error)
}

type ServerGroupServiceOperations struct {
	client *Client
}

func (s ServerGroupServiceOperations) Create(ctx context.Context, createRequest *ServerGroupRequest) (*ServerGroup, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverGroupsBasePath, createRequest)
	if err != nil {
		return nil, err
	}

	serverGroup := new(ServerGroup)
	err = s.client.Do(ctx, req, serverGroup)
	if err != nil {
		return nil, err
	}

	return serverGroup, nil
}

func (s ServerGroupServiceOperations) Get(ctx context.Context, serverGroupID string) (*ServerGroup, error) {
	path := fmt.Sprintf("%s/%s", serverGroupsBasePath, serverGroupID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	serverGroup := new(ServerGroup)
	err = s.client.Do(ctx, req, serverGroup)
	if err != nil {
		return nil, err
	}

	return serverGroup, nil
}

func (s ServerGroupServiceOperations) Delete(ctx context.Context, serverGroupID string) error {
	return genericDelete(s.client, ctx, serverGroupsBasePath, serverGroupID)
}

func (s ServerGroupServiceOperations) List(ctx context.Context) ([]ServerGroup, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverGroupsBasePath, nil)
	if err != nil {
		return nil, err
	}
	serverGroups := []ServerGroup{}
	err = s.client.Do(ctx, req, &serverGroups)
	if err != nil {
		return nil, err
	}

	return serverGroups, nil
}
