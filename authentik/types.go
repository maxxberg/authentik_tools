package authentik

import (
	"context"
	"fmt"
	"goauthentik.io/api/v3"
	"time"
)

type Config struct {
	ApiKey string
}

func NewConfig(apiKey string) *Config {
	return &Config{ApiKey: apiKey}
}

type AuthentikAPI struct {
	client *api.APIClient
	config *Config
}

func NewAuthentikAPI(aClient *api.APIClient, cfg *Config) *AuthentikAPI {
	return &AuthentikAPI{aClient, cfg}
}

func (a *AuthentikAPI) ListGroups() ([]api.Group, error) {
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), api.ContextAccessToken, a.config.ApiKey), 10*time.Second)
	defer cancel()
	groupsRequest := a.client.CoreApi.CoreGroupsList(ctx)
	groupsResult, _, err := groupsRequest.Execute()
	if err != nil {
		return nil, err
	}
	return groupsResult.Results, nil
}

func (a *AuthentikAPI) GetGroupByName(groupName string) (*api.Group, error) {
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), api.ContextAccessToken, a.config.ApiKey), 10*time.Second)
	defer cancel()
	request := a.client.CoreApi.CoreGroupsList(ctx)
	request = request.Name(groupName)
	result, _, err := request.Execute()
	if err != nil {
		return nil, err
	}
	return &result.Results[0], nil
}

func (a *AuthentikAPI) AddUsersToGroup(group *api.Group, users []int32) error {
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), api.ContextAccessToken, a.config.ApiKey), 10*time.Second)
	defer cancel()
	request := a.client.CoreApi.CoreGroupsAddUserCreate(ctx, group.Pk)
	for _, user := range users {
		uRequest := request.UserAccountRequest(*api.NewUserAccountRequest(user))
		_, err := uRequest.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AuthentikAPI) AddUsersToGroupByGroupString(group string, users []int32) error {
	groupsObject, err := a.GetGroupByName(group)
	if err != nil {
		return err
	}
	return a.AddUsersToGroup(groupsObject, users)
}

func (a *AuthentikAPI) RemoveUsersFromGroup(group *api.Group, users []int32) error {
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), api.ContextAccessToken, a.config.ApiKey), 10*time.Second)
	defer cancel()
	request := a.client.CoreApi.CoreGroupsRemoveUserCreate(ctx, group.Pk)
	for _, user := range users {
		uRequest := request.UserAccountRequest(*api.NewUserAccountRequest(user))
		_, err := uRequest.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AuthentikAPI) RemoveUsersFromGroupByGroupString(group string, users []int32) error {
	groupsObject, err := a.GetGroupByName(group)
	if err != nil {
		return err
	}
	return a.RemoveUsersFromGroup(groupsObject, users)
}

func (a *AuthentikAPI) ListUsers() []api.User {
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), api.ContextAccessToken, a.config.ApiKey), 10*time.Second)
	defer cancel()
	request := a.client.CoreApi.CoreUsersList(ctx)
	result, _, err := request.Execute()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return result.Results
}
